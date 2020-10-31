package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFilename = ".imgsync.yaml"
)

// Config contains sources and target definition for imgsync job.
type Config struct {
	Target  Repo     `yaml:"target"`
	Sources []Source `yaml:"sources,omitempty"`
}

// Auth is a username and password to authenticate to a registry.
type Auth struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// Repo is the target registry where the images defined in
// the configuration will be pushed to.
type Repo struct {
	Repository string `yaml:"repository"`
	Host       string `yaml:"host,omitempty"`
	Auth       Auth   `yaml:"auth,omitempty"`
}

// Source is a container repo in the manifest. Synced tags can be
// selected with multiple options and strategies
type Source struct {
	Source            Repo     `yaml:"source"`
	Tags              []string `yaml:"tags,omitempty"`
	MutableTags       []string `yaml:"mutableTags,omitempty"`
	RegexTags         []string `yaml:"regexTags,omitempty"`
	SyncLatestSemver  bool     `yaml:"syncLatestSemver,omitempty"`
	LatestSemverRegex string   `yaml:"latestSemverRegex,omitempty"`
}

func getConfigLocation(path string) string {
	location := path
	if !strings.Contains(location, ".yaml") && !strings.Contains(location, ".yml") {
		location = filepath.Join(path, defaultConfigFilename)
	}

	return location
}

// Get returns the configuration found at the specified path.
func Get(path string) (Config, error) {
	configLocation := getConfigLocation(path)
	configContents, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(configContents, &config); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return config, nil
}

func hostSupportsNestedRepositories(host string) bool {
	// Quay.io
	if strings.Contains(host, "quay.io") {
		return false
	}

	// Docker Registry (Docker Hub)
	// An empty host is assumed to be Docker Hub.
	if strings.Contains(host, "docker.io") || host == "" {
		return false
	}

	return true
}

// GetRepositoryAddress returns a full repository path.
func (t *Repo) GetRepositoryAddress() string {
	repoPath := "index.docker.io/" + t.Repository

	if t.Host != "" {
		repoPath = t.Host + "/" + t.Repository
	}

	return repoPath
}
