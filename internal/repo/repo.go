package repo

import (
	"github.com/google/go-containerregistry/pkg/crane"
)

// ListTags return the complete list of all existing tags for
// a given repository.
func ListTags(r string) ([]string, error) {
	tags, err := crane.ListTags(r)
	return tags, err
}

// SyncTags copies a single tag from a repo to another
func SyncTags(tag string, source string, target string) error {
	src := source + ":" + tag
	dst := target + ":" + tag
	err := crane.Copy(src, dst)
	return err
}
