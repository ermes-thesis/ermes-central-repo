package caches

import (
	"encoding/json"
	"ermes-central-repo/caches/metadata"
	"ermes-central-repo/gateway/client"
	"fmt"
	"io"
	"net/http"
)

// FunctionBundle represents a packaged function, consisting of
// the function's compiled binary and its associated metadata.
type FunctionBundle struct {
	Binary   []byte
	Metadata metadata.FunctionMetadata
}

// Fetcher defines the behaviour for retrieving function bundles.
//
// Type Parameters:
//   - T: the type of the returned bundle
type Fetcher[T any] interface {
	// Fetch retrieves the function bundle by name and version.
	//
	// Parameters:
	//	 - functionName: the logical name of the function
	//	 - functionVersion: the specific version identifier to retrieve
	//
	// Returns:
	//	 - T: the retrieved function bundle
	//	 - error: non-nil if the bundle could not be fetched
	Fetch(functionName string, functionVersion string) (T, error)

	// FetchLatest retrieves the newest available version of a function.
	//
	// Parameters:
	//	 - functionName: the logical name of the function
	//	 - cachedVersion: the version already cached locally (may
	//	 avoid downloading if still up-to-date)
	//
	// Returns:
	//	 - T: the retrieved function bundle, or nil if unchanged
	//	 - error: non-nil if the bundle could not be fetched
	FetchLatest(functionName string, cachedVersion string) (T, error)
}

// RemoteBundleFetcher implements Fetcher by fetching bundles from a
// remote HTTP service.
//
// Authentication is done via a shared secret provided at creation
type RemoteBundleFetcher struct {
	// BaseURL is the root endpoint of the remote service.
	BaseURL string
	// ermesSecret is a secret token sent to authenticate with the remote service
	ermesSecret string
}

// NewRemoteBundleFetcher creates a new RemoteBundleFetcher with the given base URL and secret.
func NewRemoteBundleFetcher(baseURL string, ermesSecret string) *RemoteBundleFetcher {
	return &RemoteBundleFetcher{BaseURL: baseURL, ermesSecret: ermesSecret}
}

// Fetch retrieves a specific version of a function bundle via HTTP.
//
// NOTE: Uses X-ERMES-SECRET header for authentication
func (f *RemoteBundleFetcher) Fetch(functionName string, functionVersion string) (*FunctionBundle, error) {
	url := fmt.Sprintf("%s/retrieve-auth-function/%s/%s", f.BaseURL, functionName, functionVersion)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-ERMES-SECRET", f.ermesSecret)

	resp, err := client.CustomClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch function: %s", string(body))
	}

	var fnBundle FunctionBundle
	err = json.Unmarshal(body, &fnBundle)
	return &fnBundle, err
}

// FetchLatest retrieves the most recent version of a function bundle from the
// remote service.
//
// NOTE: Sends "If-None-Match" header with cached version and "X-ERMES-SECRET" header
// for authentication.
func (f *RemoteBundleFetcher) FetchLatest(functionName string, cachedVersion string) (*FunctionBundle, error) {
	url := fmt.Sprintf("%s/retrieve-auth-function/%s", f.BaseURL, functionName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("If-None-Match", cachedVersion)
	req.Header.Set("X-ERMES-SECRET", f.ermesSecret)

	resp, err := client.CustomClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotModified {
		return nil, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch function: %s", string(body))
	}

	var fnBundle FunctionBundle
	err = json.Unmarshal(body, &fnBundle)
	return &fnBundle, err
}
