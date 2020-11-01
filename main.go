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
		log.Infof("Starting repo sync for %s", source.Source.Repository)

		sourceRepoAddr := source.Source.GetRepositoryAddress()
		sourceRepoTags, err := repo.ListTags(sourceRepoAddr)
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
		log.Infof("Repository %s will be synced to %s", sourceRepoAddr, targetRepoAddr)

		if err := conf.Target.Healthcheck(); err != nil {
			return err
		}

		// targetRepoTags, _ := repo.ListTags(targetRepoAddr)
		// log.Infoln(targetRepoTags)

		for _, tag := range sourceFilteredTags {

			log.Infof("Syncing %s:%s to %s:%s", sourceRepoAddr, tag, targetRepoAddr, tag)
			err = repo.SyncTags(tag, sourceRepoAddr, targetRepoAddr)
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
