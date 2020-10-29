package main

import (
	"os"

	"github.com/barthv/imgsync/internal/config"
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
		log.Infof("Starting repo sync %s", source.Source.Repository)

		repoAddr := source.Source.GetRepositoryAddress()
		log.Infoln(repoAddr)
		// repoTags, err :=
		// if err != nil {
		// 	return err
		// }
	}

	return nil
}

func main() {
	if err := runSync(); err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
}
