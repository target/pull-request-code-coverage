package main

import (
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin"
	log "github.com/sirupsen/logrus"
	"os"
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
