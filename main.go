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

	for _, source := range conf.Sources {
		log.Infof("Starting repo sync for %s", source.Source.Repository)

		repoAddr := source.Source.GetRepositoryAddress()
		repoTags, err := repo.ListTags(repoAddr)
		if err != nil {
			return err
		}

		filteredTags, err := source.FilterTags(repoTags)
		if err != nil {
			return err
		}
		log.Infoln(filteredTags)

	}

	return nil
}

func main() {
	if err := runSync(); err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
}
