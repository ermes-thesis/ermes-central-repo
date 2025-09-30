package caches

import (
	"ermes-central-repo/caches/metadata"
	"ermes-central-repo/config"
	"ermes-central-repo/logger"
	"fmt"
	"golang.org/x/sync/singleflight"
	"slices"
)

type FunctionCache interface {
	GetBinary(functionName string, functionVersion string) ([]byte, error)
	GetMetadata(functionName string, functionVersion string) (*metadata.FunctionMetadata, error)
}

type FunctionCacheImpl struct {
	MetadataBinaries *MetadataBinaryCache
	LatestVersions   GenericCache[string]
	fetcher          Fetcher[*FunctionBundle]
	group            singleflight.Group
	log              *logger.Logger
}

func NewFunctionCache(cfg *config.Config, fetcher Fetcher[*FunctionBundle], log *logger.Logger) (*FunctionCacheImpl, error) {
	binCache, err := NewMetadataBinaryCache(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating binary cache: %v", err)
	}
	latest, err := NewBaseCache[string](cfg.FunctionCacheSize, cfg.FunctionCacheTtl)
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

func NewDefaultFunctionCache(cfg *config.Config, log *logger.Logger) (*FunctionCacheImpl, error) {
	binCache, err := NewMetadataBinaryCache(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating binary cache: %v", err)
	}
	latest, err := NewBaseCache[string](cfg.FunctionCacheSize, cfg.FunctionCacheTtl)
	if err != nil {
		return nil, fmt.Errorf("error creating latest versions cache: %v", err)
	}

	return &FunctionCacheImpl{
		MetadataBinaries: binCache,
		LatestVersions:   latest,
		fetcher:          NewRemoteBundleFetcher(fmt.Sprintf("http://%s:%s", cfg.FunctionRegistryAddress, cfg.FunctionRegistryPort), cfg.ErmesSecret),
		group:            singleflight.Group{},
		log:              log,
	}, nil
}

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

func (f *FunctionCacheImpl) GetMetadata(functionName string, functionVersion string) (*metadata.FunctionMetadata, error) {
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
