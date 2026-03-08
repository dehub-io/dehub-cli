// Package model defines core domain models for dehub
package model

import "time"

// Package represents a package with its metadata
type Package struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description,omitempty"`
	License     string            `yaml:"license,omitempty"`
	Platform    string            `yaml:"platform,omitempty"`
	Board       string            `yaml:"board,omitempty"`
	Dependencies map[string]string `yaml:"dependencies,omitempty"`
}

// Repository represents a package repository
type Repository struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// ProjectConfig represents the package.yaml for a project
type ProjectConfig struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Platform     string            `yaml:"platform,omitempty"`
	Board        string            `yaml:"board,omitempty"`
	Dependencies map[string]string `yaml:"dependencies"`
	Repositories []Repository      `yaml:"repositories"`
}

// LockFile represents the package.lock file
type LockFile struct {
	Version      int             `yaml:"version"`
	Dependencies []LockedPackage `yaml:"dependencies"`
}

// LockedPackage represents a locked dependency
type LockedPackage struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	SHA256  string `yaml:"sha256"`
}

// Index represents the global repository index
type Index struct {
	SchemaVersion string   `json:"schema_version"`
	Packages      []string `json:"packages"`
	UpdatedAt     string   `json:"updated_at"`
}

// PackageIndex represents the index for a single package
type PackageIndex struct {
	SchemaVersion string        `json:"schema_version"`
	Package       string        `json:"package"`
	Versions      []VersionInfo `json:"versions"`
}

// VersionInfo represents information about a specific version
type VersionInfo struct {
	Version      string            `json:"version"`
	ArchiveURL   string            `json:"archive_url"`
	SHA256URL    string            `json:"sha256_url"`
	ReleaseDate  time.Time         `json:"release_date"`
	Dependencies []DependencySpec  `json:"dependencies,omitempty"`
}

// DependencySpec represents a dependency in the package index
type DependencySpec struct {
	Package           string `json:"package"`
	VersionConstraint string `json:"version_constraint"`
}
