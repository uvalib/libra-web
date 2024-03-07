package main

import (
	"flag"
	"log"
)

type userServiceCfg struct {
	URL string
	JWT string
}

type easyStoreConfig struct {
	mode   string // none, sqlite, postgres
	dbDir  string
	dbFile string
	dbHost string
	dbPort int
	dbName string
	dbUser string
	dbPass string
}

type namespaceConfig struct {
	oa  string
	etd string
}

type configData struct {
	port            int
	userService     userServiceCfg
	devAuthUser     string
	jwtKey          string
	easyStore       easyStoreConfig
	namespace       namespaceConfig
	busName         string
	eventSourceName string
}

func getConfiguration() *configData {
	var config configData
	flag.IntVar(&config.port, "port", 8080, "Port to offer service on")
	flag.StringVar(&config.jwtKey, "jwtkey", "", "JWT signature key")
	flag.StringVar(&config.userService.URL, "userws", "", "URL for the user service")
	flag.StringVar(&config.devAuthUser, "devuser", "", "Authorized computing id for dev")

	// easystore cfg
	flag.StringVar(&config.easyStore.mode, "esmode", "none", "EasyStore mode (sqlite, psql)")
	flag.StringVar(&config.easyStore.dbDir, "esdbdir", "/tmp", "EasyStore sqlite base directory")
	flag.StringVar(&config.easyStore.dbFile, "esdbfile", "sqlite.db", "EasyStore sqlite file")
	flag.StringVar(&config.easyStore.dbHost, "esdbhost", "", "EasyStore psql host")
	flag.IntVar(&config.easyStore.dbPort, "esdbport", 0, "EasyStore psql port")
	flag.StringVar(&config.easyStore.dbName, "esdb", "", "EasyStore psql database name")
	flag.StringVar(&config.easyStore.dbUser, "esdbuser", "", "EasyStore psql user")
	flag.StringVar(&config.easyStore.dbPass, "esdbpass", "", "EasyStore psql password")

	// namespaces
	flag.StringVar(&config.namespace.oa, "oanamespace", "oa", "Namespace for OA processing")
	flag.StringVar(&config.namespace.etd, "etdnamespace", "etd", "Namespace for ETD processing")

	// event bus
	flag.StringVar(&config.busName, "busname", "", "Event bus name")
	flag.StringVar(&config.eventSourceName, "eventsrc", "", "Event source name")

	flag.Parse()

	if config.easyStore.mode != "sqlite" && config.easyStore.mode != "postgres" {
		log.Fatal("Parameter esmode must be either sqlite or postgres")
	}
	if config.jwtKey == "" {
		log.Fatal("Parameter jwtkey is required")
	}
	if config.userService.URL == "" {
		log.Fatal("Parameter userws is required")
	}

	log.Printf("[CONFIG] port          = [%d]", config.port)
	log.Printf("[CONFIG] userws        = [%s]", config.userService.URL)
	log.Printf("[CONFIG] esmode        = [%s]", config.easyStore.mode)
	log.Printf("[CONFIG] oanamespace   = [%s]", config.namespace.oa)
	log.Printf("[CONFIG] etdnamespace  = [%s]", config.namespace.etd)
	log.Printf("[CONFIG] busname       = [%s]", config.busName)
	log.Printf("[CONFIG] eventsrc      = [%s]", config.eventSourceName)

	if config.easyStore.mode == "sqlite" {
		log.Printf("[CONFIG] esdbdir       = [%s]", config.easyStore.dbDir)
		log.Printf("[CONFIG] esdbfile      = [%s]", config.easyStore.dbFile)
	} else {
		log.Printf("[CONFIG] esdbhost      = [%s]", config.easyStore.dbHost)
		log.Printf("[CONFIG] esdbport      = [%d]", config.easyStore.dbPort)
		log.Printf("[CONFIG] esdb          = [%s]", config.easyStore.dbName)
		log.Printf("[CONFIG] esdbuser      = [%s]", config.easyStore.dbUser)
	}
	if config.devAuthUser != "" {
		log.Printf("[CONFIG] devuser       = [%s]", config.devAuthUser)
	}

	return &config
}
