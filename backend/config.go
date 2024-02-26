package main

import (
	"flag"
	"log"
)

type userServiceCfg struct {
	URL string
	JWT string
}

type easytStoreConfig struct {
	mode       string // none, sqlite, psql, other?
	namespace  string
	fileSystem string
	dbName     string
	DbUser     string
	dbPass     string
}

type configData struct {
	port        int
	userService userServiceCfg
	devAuthUser string
	jwtKey      string
	easyStore   easytStoreConfig
}

func getConfiguration() *configData {
	var config configData
	flag.IntVar(&config.port, "port", 8080, "Port to offer service on")
	flag.StringVar(&config.jwtKey, "jwtkey", "", "JWT signature key")
	flag.StringVar(&config.userService.URL, "userws", "", "URL for the user service")
	flag.StringVar(&config.devAuthUser, "devuser", "", "Authorized computing id for dev")

	// easystore cfg
	flag.StringVar(&config.easyStore.mode, "esmode", "sqlite", "EasyStore mode (sqlite, psql)")
	flag.StringVar(&config.easyStore.namespace, "esnamespace", "libraopen", "EasyStore namespace")
	flag.StringVar(&config.easyStore.fileSystem, "esfilesys", "/tmp", "EasyStore sqlite filesystem")
	flag.StringVar(&config.easyStore.dbName, "esdb", "", "EasyStore psql database name")
	flag.StringVar(&config.easyStore.DbUser, "esdbuser", "", "EasyStore psql user")
	flag.StringVar(&config.easyStore.dbPass, "esdbpass", "", "EasyStore psql password")

	flag.Parse()

	if config.jwtKey == "" {
		log.Fatal("Parameter jwtkey is required")
	}
	if config.userService.URL == "" {
		log.Fatal("Parameter userws is required")
	}

	log.Printf("[CONFIG] port          = [%d]", config.port)
	log.Printf("[CONFIG] userws        = [%s]", config.userService.URL)
	log.Printf("[CONFIG] esmode        = [%s]", config.easyStore.mode)
	log.Printf("[CONFIG] esnamespace   = [%s]", config.easyStore.namespace)

	if config.easyStore.mode == "sqlite" {
		log.Printf("[CONFIG] esfilesys     = [%s]", config.easyStore.namespace)
	} else {
		log.Printf("[CONFIG] esdb          = [%s]", config.easyStore.dbName)
		log.Printf("[CONFIG] esdbuser      = [%s]", config.easyStore.DbUser)
	}
	if config.devAuthUser != "" {
		log.Printf("[CONFIG] devuser       = [%s]", config.devAuthUser)
	}

	return &config
}
