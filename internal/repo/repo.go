package repo

import (
	"github.com/google/go-containerregistry/pkg/crane"
)

// ListRepo return the complete list of all existing tags for
// a given repository.
func ListRepo(r string) ([]string, error) {
	tags, err := crane.ListTags(r)
	return tags, err
}

// SyncTagBetweenRepos copies a single tag from a repo to another
func SyncTagBetweenRepos(tag string, source string, target string) error {
	src := source + ":" + tag
	dst := target + ":" + tag
	err := crane.Copy(src, dst)
	return err
}
