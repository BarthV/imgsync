package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
)

const (
	// pattern from https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	defaultSemverRegex = "^(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func containsSubstring(items []string, substring string) bool {
	for _, item := range items {
		if strings.Contains(substring, item) {
			return true
		}
	}
	return false
}

func (s *Source) matchingTags(tags []string) []string {
	matchingTags := []string{}
	for _, t := range tags {
		if t == "" {
			continue
		}
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
			if r == "" {
				continue
			}

			match, err := regexp.MatchString(r, t)
			if err != nil {
				return []string{}, fmt.Errorf("Matching regex \"%s\" %w", r, err)
			}

			if match {
				matchingRegexTags = append(matchingRegexTags, t)
				break
			}
		}
	}
	return matchingRegexTags, nil
}

func getHighestSemverTag(semvers []string) (string, error) {
	versions := semver.Collection{}

	for _, sv := range semvers {
		if sv == "" {
			continue
		}
		v, err := semver.NewVersion(sv)
		if err != nil {
			return "", fmt.Errorf("Semver parsing error %s : %w", sv, err)
		}
		versions = append(versions, v)
	}

	if len(versions) > 0 {
		// semver sorting is provided by semver dependency package.
		// I hope it's ok :)
		sort.Sort(semver.Collection(versions))
		highestSemver := versions[len(versions)-1].String()

		return highestSemver, nil
	}

	return "", nil
}

func getSemverTags(tags []string, regex string) ([]string, error) {
	semverTags := []string{}

	for _, t := range tags {
		match, err := regexp.MatchString(regex, t)
		if err != nil {
			return []string{}, fmt.Errorf("Matching semver regex %w", err)
		}
		if match {
			semverTags = append(semverTags, t)
		}
	}

	return semverTags, nil
}

func (s *Source) matchingLatestSemverTag(tags []string) (string, error) {
	if !(s.LatestSemverSync) {
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
	latestSemverTag, err = getHighestSemverTag(semverTags)
	if err != nil {
		return "", fmt.Errorf("finding highest semver %w", err)
	}

	return latestSemverTag, nil
}

func (s *Source) filterSpecialTags(tags []string) []string {
	if !s.OmitPreReleaseTags && !s.OmitDashedTags {
		return tags
	}

	var filteredSpecialTags []string
	for _, tag := range tags {
		// Remove tags that include prerelease related tags
		if s.OmitPreReleaseTags {
			allowedPreReleases := []string{"alpha", "beta", "rc"}
			if containsSubstring(allowedPreReleases, tag) {
				continue
			}
		}

		// Remove tags that include dash (special arch, custom builds, ...)
		if s.OmitDashedTags {
			if strings.Contains(tag, "-") {
				continue
			}
		}

		filteredSpecialTags = append(filteredSpecialTags, tag)
	}
	return filteredSpecialTags
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

	// select the highest existing semver tag based on defaut regex
	// pattern or provided regex in "latestSemverRegex" config.
	latestSemverTag, err := s.matchingLatestSemverTag(tags)
	if err != nil {
		return []string{}, err
	}
	if latestSemverTag != "" {
		filteredTags = append(filteredTags, latestSemverTag)
	}

	// remove omitted "special" tags if source specify this options.
	finalTags := s.filterSpecialTags(filteredTags)
	return finalTags, nil
}

// MissingTags return the missing srcTags from dstList
func MissingTags(srcTags []string, dstTags []string) []string {
	missingTags := []string{}

	for _, srcTag := range srcTags {
		if !stringInSlice(srcTag, dstTags) {
			missingTags = append(missingTags, srcTag)
		}
	}

	return missingTags
}
