package main

import (
	"os"

	"github.com/barthv/imgsync/internal/config"
	"github.com/barthv/imgsync/internal/repo"
	log "github.com/sirupsen/logrus"
)

func runSync() error {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	conf, err := config.Get("")
	if err != nil {
		return err
	}

	targetAddr := conf.Target.GetRepositoryAddress()
	log.Infof("Images will be synced to %s", targetAddr)

	for _, source := range conf.Sources {
		log.Infof("Starting sync : %s", source.Source.Repository)

		sourceRepoAddr := source.Source.GetRepositoryAddress()
		sourceRepoTags, err := repo.ListRepo(sourceRepoAddr)
		if err != nil {
			return err
		}

		sourceFilteredTags, err := source.FilterTags(sourceRepoTags)
		if err != nil {
			return err
		}
		log.Infof("%s : %d/%d tags matching selectors", sourceRepoAddr, len(sourceFilteredTags), len(sourceRepoTags))

		targetRepoAddr := source.GetTargetRepositoryAddress(conf.Target)
		log.Infof("%s : target is %s", sourceRepoAddr, targetRepoAddr)

		if err := conf.Target.Healthcheck(); err != nil {
			return err
		}

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
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func main() {
	if err := runSync(); err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
}
