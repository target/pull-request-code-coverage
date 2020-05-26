package main

import (
	"os"

	"git.target.com/searchoss/pull-request-code-coverage/internal/plugin"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := plugin.NewRunner().Run(os.LookupEnv, os.Stdin, os.Stdout)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Info("An unexpected error occurred")

		os.Exit(1)
	}
}
