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
		log.Infof("Starting %s repo sync", source.Source.Repository)

		sourceRepoAddr := source.Source.GetRepositoryAddress()
		sourceRepoTags, err := repo.ListRepo(sourceRepoAddr)
		if err != nil {
			return err
		}
		log.Infof("%s : %d tags discovered", sourceRepoAddr, len(sourceRepoTags))

		sourceFilteredTags, err := source.FilterTags(sourceRepoTags)
		if err != nil {
			return err
		}
		log.Infof("%s : %d tags matching provided rules", sourceRepoAddr, len(sourceFilteredTags))

		targetRepoAddr := source.GetTargetRepositoryAddress(conf.Target)
		log.Infof("%s : target is %s", sourceRepoAddr, targetRepoAddr)

		if err := conf.Target.Healthcheck(); err != nil {
			return err
		}

		targetRepoTags, _ := repo.ListRepo(targetRepoAddr)
		missingTags := config.MissingTags(sourceFilteredTags, targetRepoTags)
		allSyncTags := append(missingTags, source.MutableTags...)
		if len(missingTags) > 0 {
			log.Infof("%d tags needs to be synced", len(missingTags))
		}
		if len(source.MutableTags) > 0 {
			log.Infof("%d tags will be forced", len(missingTags))
		}

		if len(allSyncTags) == 0 {
			log.Infof("%s : target is up-to-date", sourceRepoAddr)
			continue
		}

		for _, tag := range allSyncTags {
			log.Infof("Syncing %s:%s to %s:%s", sourceRepoAddr, tag, targetRepoAddr, tag)
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
