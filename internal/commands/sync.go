package commands

import (
	"fmt"

	"github.com/barthv/imgsync/internal/config"
	"github.com/barthv/imgsync/internal/repo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newSyncCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "sync",
		Short: "sync images from sources to target",

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runSyncCommand(); err != nil {
				return fmt.Errorf("sync command: %w", err)
			}

			return nil
		},
	}

	return &cmd
}

func runSyncCommand() error {
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	conf, err := config.Get(viper.GetString("confpath"))
	if err != nil {
		return err
	}

	targetAddr := conf.Target.GetRepositoryAddress()

	if conf.Target.Auth.Username != "" {
		log.Debugln("Encoding target credentials")
		err := repo.SetHostCredentials(targetAddr, conf.Target.Auth.Username, conf.Target.Auth.Password)
		if err != nil {
			log.Errorf("target auth failed : %s", err)
			return err
		}
	}

	if err := conf.Target.Healthcheck(); err != nil {
		log.Debugln("target test : %s", err)
		log.Errorln("Target registry is unavailable. Stopping")
		return err
	}

	log.Debugln("Target registry is healthy")
	log.Infof("Images will be synced to %s", targetAddr)

	for _, source := range conf.Sources {
		log.Infof("Starting sync : %s", source.Source.Repository)

		sourceRepoAddr := source.Source.GetRepositoryAddress()

		if source.Source.Auth.Username != "" {
			log.Debugf("%s : encoding source credentials", sourceRepoAddr)
			err := repo.SetHostCredentials(sourceRepoAddr, source.Source.Auth.Username, source.Source.Auth.Password)
			if err != nil {
				log.Errorf("source auth failed : %s", err)
				return err
			}
		}

		sourceRepoTags, err := repo.ListRepo(sourceRepoAddr)
		if conf.ContinueOnSyncError && err != nil {
			log.Debugf("%s", err)
			log.Warnln("continueOnSyncError flag enabled : List source error ignored.")
			continue
		}
		if err != nil {
			return err
		}

		targetRepoAddr := source.GetTargetRepositoryAddress(conf.Target)
		log.Infof("%s : target repo is %s", sourceRepoAddr, targetRepoAddr)

		sourceFilteredTags, err := source.FilterTags(sourceRepoTags)
		if err != nil {
			return err
		}
		log.Infof("%s : %d/%d tags matching selectors", sourceRepoAddr, len(sourceFilteredTags), len(sourceRepoTags))

		targetRepoTags, _ := repo.ListRepo(targetRepoAddr)
		missingTags := config.MissingTags(sourceFilteredTags, targetRepoTags)
		allSyncTags := append(missingTags, source.MutableTags...)
		if len(missingTags) > 0 {
			log.Infof("%s : %d missing tags to sync", sourceRepoAddr, len(missingTags))
		}
		if len(source.MutableTags) > 0 {
			log.Infof("%s : %d tags forced to sync", sourceRepoAddr, len(source.MutableTags))
		}

		if len(allSyncTags) == 0 {
			log.Infof("%s : target is up-to-date", sourceRepoAddr)
			continue
		}

		for _, tag := range allSyncTags {
			log.Infof("%s : syncing %s to %s:%s", sourceRepoAddr, tag, targetRepoAddr, tag)
			err = repo.SyncTagBetweenRepos(tag, sourceRepoAddr, targetRepoAddr)
			if conf.ContinueOnSyncError && err != nil {
				log.Errorf("%s", err)
				log.Warnln("continueOnSyncError flag enabled : Sync error ignored.")
				continue
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
