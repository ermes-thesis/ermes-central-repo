package caches

import (
	"caches/metadata"
	"encoding/json"
	"ermes-central-repo/config"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FunctionBundle struct {
	Binary   []byte
	Metadata metadata.FunctionMetadata
}

type Fetcher[T any] interface {
	Fetch(identifier string) (T, error)
}

type RemoteBundleFetcher struct {
	BaseURL string
}

func (f *RemoteBundleFetcher) Fetch(functionName string) (*FunctionBundle, error) {
	//TODO: Fix with the right decision of how taking functionName + version
	parts := strings.SplitN(functionName, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid function name %s, expected name:version", functionName)
	}
	name, version := parts[0], parts[1]

	url := fmt.Sprintf("%s/retrieve-auth-function/%s/%s", f.BaseURL, name, version)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-ERMES-SECRET", config.ErmesSecret())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch function %s:%s, status code: %d, body: %s", name, version, resp.StatusCode, string(body))
	}

	var fnBundle FunctionBundle
	err = json.Unmarshal(body, &fnBundle)
	return &fnBundle, err
}

func NewRemoteBundleFetcher(baseURL string) *RemoteBundleFetcher {
	return &RemoteBundleFetcher{BaseURL: baseURL}
}
