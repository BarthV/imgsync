package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
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
	Scheme     string `yaml:"scheme:omitempty"`
	Host       string `yaml:"host,omitempty"`
	Auth       Auth   `yaml:"auth,omitempty"`
}

// Source is a container repo in the manifest. Synced tags can be
// selected with multiple options and strategies
type Source struct {
	Source             Repo     `yaml:"source"`
	Tags               []string `yaml:"tags,omitempty"`
	MutableTags        []string `yaml:"mutableTags,omitempty"`
	RegexTags          []string `yaml:"regexTags,omitempty"`
	LatestSemverSync   bool     `yaml:"latestSemverSync,omitempty"`
	LatestSemverRegex  string   `yaml:"latestSemverRegex,omitempty"`
	OmitPreReleaseTags bool     `yaml:"omitPreReleaseTags,omitempty"`
	OmitDashedTags     bool     `yaml:"omitDashedTags,omitempty"`
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

func (r *Repo) supportNestedRepositories() bool {
	// Quay.io
	if strings.Contains(r.Host, "quay.io") {
		return false
	}
	// Docker Registry (Docker Hub)
	// An empty host is assumed to be Docker Hub.
	if strings.Contains(r.Host, "docker.io") || r.Host == "" {
		return false
	}
	return true
}

// GetRepositoryAddress returns a full repository path.
func (r *Repo) GetRepositoryAddress() string {
	repoPath := "index.docker.io/" + r.Repository

	if r.Host != "" {
		repoPath = r.Host + "/" + r.Repository
	}

	return repoPath
}

// GetTargetRepositoryAddress compute the final target repo adress
// from a source repo with nested repo support if handled by target.
func (s *Source) GetTargetRepositoryAddress(targetRepo Repo) string {
	var target string

	if targetRepo.supportNestedRepositories() {
		target = s.Source.Repository
	} else {
		target = filepath.Base(s.Source.Repository)
	}

	target = targetRepo.GetRepositoryAddress() + target

	return target
}

// Healthcheck tests "/v2" registry url availability.
// 2xx and 401 response status codes are valid.
func (r *Repo) Healthcheck() error {
	repoHostURL := "https://index.docker.io/"

	if r.Host != "" {
		repoHostURL = r.Host
		if r.Scheme != "" {
			repoHostURL = r.Scheme + "://" + repoHostURL
		} else {
			repoHostURL = "http://" + repoHostURL
		}
	}

	httpRes, err := http.Get(repoHostURL)
	if err != nil {
		return err
	}
	if (httpRes.StatusCode >= 200 && httpRes.StatusCode <= 299) || httpRes.StatusCode == 401 {
		return nil
	}
	return fmt.Errorf("Repo host healthcheck: %s status code %d must be 2xx or 401", repoHostURL, httpRes.StatusCode)
}
