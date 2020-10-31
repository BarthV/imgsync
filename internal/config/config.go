package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFilename = ".imgsync.yaml"
	// pattern from https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	defaultSemverRegex = "^(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (s *Source) matchingTags(tags []string) []string {
	matchingTags := []string{}
	for _, t := range tags {
		if stringInSlice(t, s.Tags) {
			matchingTags = append(matchingTags, t)
		}
	}
	return matchingTags
}

func (s *Source) matchingRegexTags(tags []string) ([]string, error) {
	matchingRegexTags := []string{}
	for _, t := range tags {
		for _, r := range s.RegexTags {
			match, err := regexp.MatchString(r, t)
			if err != nil {
				return []string{}, fmt.Errorf("Matching regexp \"%s\" %v", r, err)
			}
			if match {
				matchingRegexTags = append(matchingRegexTags, t)
				break
			}
		}
	}
	return matchingRegexTags, nil
}

func getHighestSemver(semvers []string) (string, error) {
	versions := semver.Collection{}

	for _, sv := range semvers {
		v, err := semver.NewVersion(sv)
		if err != nil {
			return "", fmt.Errorf("Semver parsing error %s : %v", sv, err)
		}
		versions = append(versions, v)
	}

	sort.Sort(semver.Collection(versions))
	highestSemver := versions[len(versions)-1].String()

	return highestSemver, nil
}

func getSemverTags(tags []string, regex string) ([]string, error) {
	semverTags := []string{}

	for _, t := range tags {
		match, err := regexp.MatchString(regex, t)
		if err != nil {
			return []string{}, fmt.Errorf("Matching semver regexp %v", err)
		}
		if match {
			semverTags = append(semverTags, t)
		}
	}

	return semverTags, nil
}

func (s *Source) matchingLatestSemverTag(tags []string) (string, error) {
	if !(s.SyncLatestSemver) {
		return "", nil
	}

	latestSemverTag := ""
	regexString := defaultSemverRegex

	// use any provided semver regex pattern instead of default one
	if s.LatestSemverRegex != "" {
		regexString = s.LatestSemverRegex
	}

	// get all tags matching semver regex ...
	semverTags, err := getSemverTags(tags, regexString)
	// ... then sort them out to keep to bigger one !
	latestSemverTag, err = getHighestSemver(semverTags)
	if err != nil {
		return "", fmt.Errorf("finding highest semver %v", err)
	}

	return latestSemverTag, nil
}

// FilterTags compute filtering rules of a source and
// applies it against a list of tags.
func (s *Source) FilterTags(tags []string) ([]string, error) {
	filteredTags := []string{}

	// select tags based on listed "tags"
	filteredTags = append(filteredTags, s.matchingTags(tags)...)

	// select tags matching regex listed on "regexTags"
	matchingRegexTags, err := s.matchingRegexTags(tags)
	if err != nil {
		return []string{}, err
	}
	filteredTags = append(filteredTags, matchingRegexTags...)

	// select the highest existing semver tag based on defaut
	// regex or provided regex in "latestSemverRegex" config
	latestSemverTag, err := s.matchingLatestSemverTag(tags)
	if err != nil {
		return []string{}, err
	}
	if latestSemverTag != "" {
		filteredTags = append(filteredTags, latestSemverTag)
	}

	return filteredTags, nil
}
