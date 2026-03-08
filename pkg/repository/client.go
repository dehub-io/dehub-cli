// Package repository provides repository access abstraction
package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dehub-io/dehub-cli/pkg/model"
)

// Client provides access to a package repository
type Client struct {
	URL        string
	HTTPClient *http.Client
}

// NewClient creates a new repository client
func NewClient(url string) *Client {
	return &Client{
		URL: url,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetIndex fetches the global repository index
func (c *Client) GetIndex() (*model.Index, error) {
	url := c.URL + "index.json"
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch index: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var index model.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return &index, nil
}

// GetPackageIndex fetches the index for a specific package
func (c *Client) GetPackageIndex(name string) (*model.PackageIndex, error) {
	url := fmt.Sprintf("%spackages/%s/index.json", c.URL, name)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("package not found: %s", name)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read package index: %w", err)
	}

	var pkgIndex model.PackageIndex
	if err := json.Unmarshal(data, &pkgIndex); err != nil {
		return nil, fmt.Errorf("failed to parse package index: %w", err)
	}

	return &pkgIndex, nil
}

// DownloadArchive downloads a package archive
func (c *Client) DownloadArchive(name, version string) ([]byte, error) {
	url := fmt.Sprintf("%sreleases/download/%s/%s/package.tar.gz", c.URL, name, version)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("archive not found: %s@%s", name, version)
	}

	return io.ReadAll(resp.Body)
}

// GetSHA256 fetches the SHA256 checksum for a package
func (c *Client) GetSHA256(name, version string) (string, error) {
	url := fmt.Sprintf("%sreleases/download/%s/%s/sha256", c.URL, name, version)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch sha256: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("sha256 not found: %s@%s", name, version)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read sha256: %w", err)
	}

	return string(data), nil
}
