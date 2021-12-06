package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/target/pull-request-code-coverage/internal/plugin"
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
