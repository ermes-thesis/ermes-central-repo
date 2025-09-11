package caches

import (
	"caches/metadata"
	"ermes-central-repo/config"
	"fmt"

	"golang.org/x/sync/singleflight"
)

type FunctionCache interface {
	GetBinary(functionName string, functionVersion string) ([]byte, error)
	GetMetadata(functionName string, functionVersion string) (*metadata.FunctionMetadata, error)
}

type FunctionCacheImpl struct {
	MetadataBinaries *MetadataBinaryCache
	LatestVersions   *BaseCache[string]
	fetcher          Fetcher[*FunctionBundle]
	group            singleflight.Group
}

func NewFunctionCache(fetcher Fetcher[*FunctionBundle]) (*FunctionCacheImpl, error) {
	binCache, err := NewMetadataBinaryCache()
	if err != nil {
		return nil, fmt.Errorf("error creating binary cache: %v", err)
	}
	latest, err := NewBaseCache[string](config.FunctionCacheSize(), config.FunctionCacheTtl())
	if err != nil {
		return nil, fmt.Errorf("error creating latest versions cache: %v", err)
	}

	return &FunctionCacheImpl{
		MetadataBinaries: binCache,
		LatestVersions:   latest,
		fetcher:          fetcher,
	}, nil
}

func NewDefaultFunctionCache() (*FunctionCacheImpl, error) {
	binCache, err := NewMetadataBinaryCache()
	if err != nil {
		return nil, fmt.Errorf("error creating binary cache: %v", err)
	}
	latest, err := NewBaseCache[string](config.FunctionCacheSize(), config.FunctionCacheTtl())
	if err != nil {
		return nil, fmt.Errorf("error creating latest versions cache: %v", err)
	}

	return &FunctionCacheImpl{
		MetadataBinaries: binCache,
		LatestVersions:   latest,
		fetcher:          NewRemoteBundleFetcher(fmt.Sprintf("http://%s:%s", config.FunctionRegistryAddress(), config.FunctionRegistryPort())),
		group:            singleflight.Group{},
	}, nil
}

func (f *FunctionCacheImpl) GetBinary(functionName string, functionVersion string) ([]byte, error) {
	if functionVersion == "latest" {
		bundle, err := f.resolveLatest(functionName, "binary")
		if err != nil {
			return nil, err
		}
		if bundle != nil {
			return bundle.Binary, nil
		}
		panic("This should be unreachable!!!")
	}

	bin, _, ok, err := f.MetadataBinaries.GetWithBinary(functionName, functionVersion)
	if err != nil {
		return nil, err
	}
	if ok {
		return bin, nil
	}

	fnBundle, err := f.fetchAndStore(functionName, functionVersion)
	if err != nil {
		return nil, err
	}
	return fnBundle.Binary, err
}

func (f *FunctionCacheImpl) GetMetadata(functionName string, functionVersion string) (*metadata.FunctionMetadata, error) {
	if functionVersion == "latest" {
		bundle, err := f.resolveLatest(functionName, "metadata")
		if err != nil {
			return nil, err
		}
		if bundle != nil {
			return &bundle.Metadata, nil
		}
		panic("This should be unreachable!!!")
	}

	meta, ok := f.MetadataBinaries.Get(functionName, functionVersion)
	if ok {
		return &meta, nil
	}

	fnBundle, err := f.fetchAndStore(functionName, functionVersion)
	if err != nil {
		return nil, err
	}
	return &fnBundle.Metadata, err
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
		return err
	}
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
			return nil, err
		}

		if bundle == nil || len(bundle.Binary) == 0 {
			return nil, fmt.Errorf("invalid bundle from registry")
		}

		err = f.store(bundle)
		if err != nil {
			return nil, err
		}

		return bundle, nil
	})

	if err != nil {
		return nil, err
	}
	return fetchedValue.(*FunctionBundle), nil
}
