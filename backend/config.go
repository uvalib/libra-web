package main

import (
	"flag"
	"log"
)

type configData struct {
	port        int
	devAuthUser string
	jwtKey      string
}

func getConfiguration() *configData {
	var config configData
	flag.IntVar(&config.port, "port", 8080, "Port to offer service on (default 8085)")
	flag.StringVar(&config.jwtKey, "jwtkey", "", "JWT signature key")

	// dev user
	flag.StringVar(&config.devAuthUser, "devuser", "", "Authorized computing id for dev")

	flag.Parse()

	if config.jwtKey == "" {
		log.Fatal("Parameter jwtkey is required")
	}

	log.Printf("[CONFIG] port          = [%d]", config.port)
	if config.devAuthUser != "" {
		log.Printf("[CONFIG] devuser       = [%s]", config.devAuthUser)
	}

	return &config
}
