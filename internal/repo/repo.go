package repo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
)

type loginOptions struct {
	serverAddress string
	user          string
	password      string
	passwordStdin bool
}

func login(opts loginOptions) error {
	if opts.passwordStdin {
		contents, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		opts.password = strings.TrimSuffix(string(contents), "\n")
		opts.password = strings.TrimSuffix(opts.password, "\r")
	}
	if opts.user == "" && opts.password == "" {
		return errors.New("username and password required")
	}
	cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return err
	}
	creds := cf.GetCredentialsStore(opts.serverAddress)
	if opts.serverAddress == name.DefaultRegistry {
		opts.serverAddress = authn.DefaultAuthKey
	}
	if err := creds.Store(types.AuthConfig{
		ServerAddress: opts.serverAddress,
		Username:      opts.user,
		Password:      opts.password,
	}); err != nil {
		return err
	}

	return cf.Save()
}

// ListRepo return the complete list of all existing tags for
// a given repository.
func ListRepo(r string) ([]string, error) {
	tags, err := crane.ListTags(r)
	if err != nil {
		err = fmt.Errorf("repo list tags : %w", err)
	}

	return tags, err
}

// SetHostCredentials registers credentials for a given registry address.
// Credentials are persisted in local userdir (as docker cli would do).
func SetHostCredentials(repoAddress string, user string, pass string) error {
	if user == "" || pass == "" {
		return fmt.Errorf("host login : username and password required")
	}

	cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return fmt.Errorf("host auth : %w", err)
	}

	creds := cf.GetCredentialsStore(repoAddress)
	serverAddress := authn.DefaultAuthKey
	if repoAddress != "" {
		serverAddress = repoAddress
	}

	err = creds.Store(types.AuthConfig{
		ServerAddress: serverAddress,
		Username:      user,
		Password:      pass,
	})
	if err != nil {
		return fmt.Errorf("host auth : %w", err)
	}

	err = cf.Save()
	if err != nil {
		err = fmt.Errorf("host auth save : %w", err)
	}

	return err
}

// SyncTagBetweenRepos copies a single tag from a repo to another
func SyncTagBetweenRepos(tag string, source string, target string) error {
	src := source + ":" + tag
	dst := target + ":" + tag

	err := crane.Copy(src, dst)
	if err != nil {
		err = fmt.Errorf("repo copy tag : %w", err)
	}

	return err
}
