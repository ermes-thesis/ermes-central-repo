package caches

import (
	"ermes-central-repo/config"
	"fmt"
	"slices"

	"github.com/ermes-thesis/ermes-common/caches"
	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/ermes-thesis/ermes-common/model"
	"golang.org/x/sync/singleflight"
)

// FunctionCache defins an interface for retrieving function binaries and metadata,
// with transparent local caching and optional remote fetching.
type FunctionCache interface {
	// GetBinary retrieves the compiled binary for the specified function and version.
	//
	// Parameters:
	//	 - functionName: the logical name of the function.
	//	 - functionVersion: the version to retrieve ("latest" is supported).
	//
	// Returns:
	//	 - []byte: the function's compiled binary data.
	//	 - error: non-nil if the binary could not be retrieved or fetched.
	GetBinary(functionName string, functionVersion string) ([]byte, error)

	// GetMetadata retrieves the metadata for the specified function and version.
	//
	// Parameters:
	//	 - functionName: the logical name of the function.
	//	 - functionVersion: the version to retrieve ("latest" is supported).
	//
	// Returns:
	//	 - *metadata.FunctionMetadata: the retrieved metadata object of the function.
	//	 - error: non-nil if the metadata could not be retrieved or fetched.
	GetMetadata(functionName string, functionVersion string) (*model.FunctionMetadata, error)
}

// FunctionCacheImpl is a concrete implementation of FunctionCache that uses
// multiple internal caches to store function binaries and metadata.
//
// It coordinates local caching with remote fetching via a Fetcher implementation,
// uses singleflight to avoid redundant requests, and maintains a "latest version"
// cache for quick resolution. Since the underlying MetadataBinaryCache and BaseCache
// are safe for concurrent use, also FunctionCacheImpl is safe for concurrent uses
// by multiple goroutines
type FunctionCacheImpl struct {
	MetadataBinaries *caches.MetadataBinaryCache
	LatestVersions   caches.GenericCache[string]
	fetcher          caches.Fetcher[*FunctionBundle]
	group            singleflight.Group
	log              *logger.Logger
}

// FunctionBundle represents a packaged function, consisting of
// the function's compiled binary and its associated metadata.
type FunctionBundle struct {
	Binary   []byte
	Metadata model.FunctionMetadata
}

// NewFunctionCache creates a FunctionCacheImpl using the provided fetcher.
//
// It initializes metadata/binary caches and a latest-version tracker based on
// the configuration parameters
func NewFunctionCache(cfg *config.Config, fetcher caches.Fetcher[*FunctionBundle], log *logger.Logger) (*FunctionCacheImpl, error) {
	binCache, err := caches.NewMetadataBinaryCache(cfg.FunctionCache)
	if err != nil {
		return nil, fmt.Errorf("error creating binary cache: %v", err)
	}
	latest, err := caches.NewBaseCache[string](cfg.FunctionCache.Size, cfg.FunctionCache.Ttl)
	if err != nil {
		return nil, fmt.Errorf("error creating latest versions cache: %v", err)
	}

	return &FunctionCacheImpl{
		MetadataBinaries: binCache,
		LatestVersions:   latest,
		fetcher:          fetcher,
		log:              log,
	}, nil
}

// NewDefaultFunctionCache creates a FunctionCacheImpl with a default remote Fetcher
// that will communicate with the Ermes Function Registry node
func NewDefaultFunctionCache(cfg *config.Config, log *logger.Logger) (*FunctionCacheImpl, error) {
	binCache, err := caches.NewMetadataBinaryCache(cfg.FunctionCache)
	if err != nil {
		return nil, fmt.Errorf("error creating binary cache: %v", err)
	}
	latest, err := caches.NewBaseCache[string](cfg.FunctionCache.Size, cfg.FunctionCache.Ttl)
	if err != nil {
		return nil, fmt.Errorf("error creating latest versions cache: %v", err)
	}

	return &FunctionCacheImpl{
		MetadataBinaries: binCache,
		LatestVersions:   latest,
		fetcher:          caches.NewRemoteBundleFetcher[FunctionBundle](fmt.Sprintf("http://%s:%s", cfg.FunctionRegistryAddress, cfg.FunctionRegistryPort), cfg.ErmesSecret),
		group:            singleflight.Group{},
		log:              log,
	}, nil
}

// GetBinary returns the function binary from cache or fetches it from the remote
// registry if not present or outdated. "latest" versions are resolved automatically.
func (f *FunctionCacheImpl) GetBinary(functionName string, functionVersion string) ([]byte, error) {
	if functionVersion == "latest" {
		bundle, err := f.resolveLatest(functionName, "binary")
		if err != nil {
			f.log.Error("Failed to resolve latest binary bundle.",
				"fn_name", functionName,
				"fn_version", functionVersion,
				"error", err,
			)
			return nil, err
		}
		if bundle != nil {
			f.log.Info("Returning latest binary bundle",
				"fn_name", functionName,
				"fn_version", functionVersion,
			)
			return bundle.Binary, nil
		}
		panic("This should be unreachable!!!")
	}

	bin, meta, ok, err := f.MetadataBinaries.GetWithBinary(functionName, functionVersion)
	if err != nil {
		f.log.Error("Failed to get binary from cache.",
			"fn_name", functionName,
			"fn_version", functionVersion,
			"fn_metadata", meta,
			"error", err,
		)
		return nil, err
	}
	if ok {
		f.log.Info("Binary retrieve from cache",
			"fn_name", functionName,
			"fn_version", functionVersion,
			"fn_metadata", meta,
		)
		return bin, nil
	}

	fnBundle, err := f.fetchAndStore(functionName, functionVersion)
	if err != nil {
		f.log.Error("Failed to fetch and store binary bundle.",
			"fn_name", functionName,
			"fn_version", functionVersion,
			"error", err)
		return nil, err
	}
	return fnBundle.Binary, err
}

// GetMetadata returns the function metadata from cache or fetches it from the
// remote registry if not present or outdated. Handles "latest" resolution.
func (f *FunctionCacheImpl) GetMetadata(functionName string, functionVersion string) (*model.FunctionMetadata, error) {
	if functionVersion == "latest" {
		bundle, err := f.resolveLatest(functionName, "metadata")
		if err != nil {
			f.log.Error("Failed to resolve latest metadata bundle.",
				"fn_name", functionName,
				"fn_version", functionVersion,
				"error", err,
			)
			return nil, err
		}
		if bundle != nil {
			f.log.Info("Returning latest metadata bundle",
				"fn_name", functionName,
				"fn_version", functionVersion,
			)
			return &bundle.Metadata, nil
		}
		panic("This should be unreachable!!!")
	}

	meta, ok := f.MetadataBinaries.Get(functionName, functionVersion)
	if ok {
		f.log.Info("Metadata retrieved from cache",
			"fn_name", functionName,
			"fn_version", functionVersion,
			"fn_metadata", meta,
		)
		return &meta, nil
	}

	fnBundle, err := f.fetchAndStore(functionName, functionVersion)
	if err != nil {
		f.log.Error("Failed to fetch and store metadata bundle.",
			"fn_name", functionName,
			"fn_version", functionVersion,
			"error", err,
		)
		return nil, err
	}
	return &fnBundle.Metadata, nil
}

// getSnapshot attempts to build a FunctionBundle directly from cached data,
// returning nil if the requested version is not found.
func (f *FunctionCacheImpl) getSnapshot(functionName string, functionVersion string, want string) (*FunctionBundle, error) {
	switch want {
	case "metadata":
		if meta, ok := f.MetadataBinaries.Get(functionName, functionVersion); ok {
			return &FunctionBundle{Metadata: meta}, nil
		}

	case "binary":
		if bin, _, okBin, err := f.MetadataBinaries.GetWithBinary(functionName, functionVersion); okBin && len(bin) > 0 {
			return &FunctionBundle{Binary: bin}, err
		}

	default:
		return nil, fmt.Errorf("unknown want type: %s", want)
	}
	return nil, nil
}

// resolveLatest resolves the "latest" version of a function, updating caches as needed.
//
// It uses the LatestVersions cache to determine if a refresh is required,
// and may fetch from the remote registry if data is missing or expired.
func (f *FunctionCacheImpl) resolveLatest(functionName, want string) (*FunctionBundle, error) {
	cachedVersion, found, expired := f.LatestVersions.GetWithStatus(functionName)
	bundle, err := f.getSnapshot(functionName, cachedVersion, want)
	if err != nil {
		return nil, err
	}

	if bundle == nil {
		fetchedBundle, err := f.fetcher.Fetch(functionName, "latest")
		if err != nil {
			return nil, err
		}

		f.LatestVersions.Set(functionName, fetchedBundle.Metadata.Version)
		err = f.store(fetchedBundle)
		if err != nil {
			return nil, err
		}
		return fetchedBundle, nil
	} else {
		if found && !expired {
			return bundle, nil
		}

		fetchedBundle, err := f.fetcher.FetchLatest(functionName, cachedVersion)
		if fetchedBundle != nil {
			f.LatestVersions.Set(functionName, fetchedBundle.Metadata.Version)
			err = f.store(fetchedBundle)
			if err != nil {
				return nil, err
			}
			return fetchedBundle, nil
		}

		// refresh latest cache
		f.LatestVersions.Set(functionName, cachedVersion)
		return bundle, err
	}
}

// store writes a FunctionBundle to local caches (metadata and binary).
func (f *FunctionCacheImpl) store(bundle *FunctionBundle) error {
	if err := f.MetadataBinaries.Set(bundle.Metadata, bundle.Binary); err != nil {
		f.log.Error("Failed to store bundle in cache",
			"fn_binary", append(slices.Clone(bundle.Binary[:min(len(bundle.Binary), 30)])),
			"fn_version", bundle.Metadata.Version,
			"error", err,
		)
		return err
	}

	f.log.Info("Bundle stored in cache",
		"fn_binary", append(slices.Clone(bundle.Binary[:min(len(bundle.Binary), 30)])),
		"fn_version", bundle.Metadata.Version,
	)
	return nil
}

// fetchAndStore fetches a function bundle from the remote registry and stores it locally.
// It uses singleflight to ensure only one fetch per (name, version) key at a time.
func (f *FunctionCacheImpl) fetchAndStore(functionName string, functionVersion string) (*FunctionBundle, error) {
	if functionVersion == "latest" {
		panic("This should be unreachable!!!")
	}

	key := functionName + "@" + functionVersion
	fetchedValue, err, _ := f.group.Do(key, func() (any, error) {
		bundle, err := f.fetcher.Fetch(functionName, functionVersion)
		if err != nil {
			f.log.Error("Failed to fetch function bundle.",
				"fn_name", functionName,
				"fn_version", functionVersion,
				"error", err,
			)
			return nil, err
		}

		if bundle == nil || len(bundle.Binary) == 0 {
			f.log.Error("Fetched invalid bundle.",
				"fn_name", functionName,
				"fn_version", functionVersion,
			)
			return nil, fmt.Errorf("invalid bundle from registry")
		}

		err = f.store(bundle)
		if err != nil {
			f.log.Error("Failed to store bundle.",
				"fn_name", functionName,
				"fn_version", functionVersion,
				"error", err,
			)
			return nil, err
		}

		f.log.Info("Function bundle fetched and stored successfully.",
			"fn_name", functionName,
			"fn_version", functionVersion,
		)
		return bundle, nil
	})

	if err != nil {
		return nil, err
	}
	return fetchedValue.(*FunctionBundle), nil
}
