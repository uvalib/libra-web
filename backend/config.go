package main

import (
	"flag"
	"log"
)

type userServiceCfg struct {
	URL string
	JWT string
}

type orcidServiceCfg struct {
	GetURL string
}

type easyStoreConfig struct {
	mode      string // sqlite, postgres, s3
	dbDir     string
	dbFile    string
	dbHost    string
	dbPort    int
	dbName    string
	dbUser    string
	dbPass    string
	dbTimeout int
	s3Bucket  string
}

type devConfig struct {
	user    string
	role    string
	fakeBus bool
}

type configData struct {
	port            int
	userService     userServiceCfg
	orcidService    orcidServiceCfg
	auditQueryURL   string
	jwtKey          string
	easyStore       easyStoreConfig
	namespace       string
	busName         string
	eventSourceName string
	indexURL        string
	dev             devConfig
}

func getConfiguration() *configData {
	var config configData
	flag.IntVar(&config.port, "port", 8080, "Port to offer service on")
	flag.StringVar(&config.jwtKey, "jwtkey", "", "JWT signature key")
	flag.StringVar(&config.userService.URL, "userws", "", "URL for the user service")
	flag.StringVar(&config.auditQueryURL, "auditqueryurl", "", "Query URL for the audit service")
	flag.StringVar(&config.orcidService.GetURL, "getorcidurl", "", "GET orcid for user service")

	// dev mode
	flag.StringVar(&config.dev.user, "devuser", "", "Authorized computing id for dev")
	flag.StringVar(&config.dev.role, "devrole", "user", "Role for dev user")
	flag.BoolVar(&config.dev.fakeBus, "devbus", false, "bus dev mode (no events sent out)")

	// easystore cfg
	flag.StringVar(&config.easyStore.mode, "esmode", "none", "EasyStore mode (sqlite, psql)")
	flag.StringVar(&config.easyStore.dbDir, "esdbdir", "/tmp", "EasyStore sqlite base directory")
	flag.StringVar(&config.easyStore.dbFile, "esdbfile", "sqlite.db", "EasyStore sqlite file")
	flag.StringVar(&config.easyStore.dbHost, "esdbhost", "", "EasyStore psql host")
	flag.IntVar(&config.easyStore.dbPort, "esdbport", 0, "EasyStore psql port")
	flag.StringVar(&config.easyStore.dbName, "esdb", "", "EasyStore psql database name")
	flag.StringVar(&config.easyStore.dbUser, "esdbuser", "", "EasyStore psql user")
	flag.StringVar(&config.easyStore.dbPass, "esdbpass", "", "EasyStore psql password")
	flag.IntVar(&config.easyStore.dbTimeout, "esdbtimeout", 30, "EasyStore psql password")
	flag.StringVar(&config.easyStore.s3Bucket, "esbucket", "", "EasyStore S3 bucket name for file storage")

	// namespace
	flag.StringVar(&config.namespace, "namespace", "libraetd", "Namespace for work processing")

	// search index
	flag.StringVar(&config.indexURL, "index", "", "Search index URL")

	// event bus
	flag.StringVar(&config.busName, "busname", "", "Event bus name")
	flag.StringVar(&config.eventSourceName, "eventsrc", "", "Event source name")

	flag.Parse()

	if config.easyStore.mode != "sqlite" && config.easyStore.mode != "postgres" && config.easyStore.mode != "s3" {
		log.Fatal("Parameter esmode must be either sqlite, postgres or s3")
	}
	if config.jwtKey == "" {
		log.Fatal("Parameter jwtkey is required")
	}
	if config.userService.URL == "" {
		log.Fatal("Parameter userws is required")
	}
	if config.auditQueryURL == "" {
		log.Fatal("Parameter auditqueryurl is required")
	}
	if config.busName == "" {
		log.Fatal("Parameter busname is required")
	}
	if config.eventSourceName == "" {
		log.Fatal("Parameter eventsrc is required")
	}
	if config.easyStore.mode == "s3" && config.easyStore.s3Bucket == "" {
		log.Fatal("Parameter esbucket is required for easystore s3 mode")
	}
	if config.indexURL == "" {
		log.Fatal("Parameter index is required")
	}

	log.Printf("[CONFIG] port          = [%d]", config.port)
	log.Printf("[CONFIG] userws        = [%s]", config.userService.URL)
	log.Printf("[CONFIG] getorcidurl   = [%s]", config.orcidService.GetURL)
	log.Printf("[CONFIG] auditqueryurl = [%s]", config.auditQueryURL)
	log.Printf("[CONFIG] esmode        = [%s]", config.easyStore.mode)
	log.Printf("[CONFIG] namespace     = [%s]", config.namespace)
	log.Printf("[CONFIG] eventsrc      = [%s]", config.eventSourceName)
	log.Printf("[CONFIG] busname       = [%s]", config.busName)
	log.Printf("[CONFIG] index         = [%s]", config.indexURL)

	if config.easyStore.mode == "sqlite" {
		log.Printf("[CONFIG] esdbdir       = [%s]", config.easyStore.dbDir)
		log.Printf("[CONFIG] esdbfile      = [%s]", config.easyStore.dbFile)
	} else {
		log.Printf("[CONFIG] esdbhost      = [%s]", config.easyStore.dbHost)
		log.Printf("[CONFIG] esdbport      = [%d]", config.easyStore.dbPort)
		log.Printf("[CONFIG] esdb          = [%s]", config.easyStore.dbName)
		log.Printf("[CONFIG] esdbuser      = [%s]", config.easyStore.dbUser)
		log.Printf("[CONFIG] esdbtimeout   = [%d]", config.easyStore.dbTimeout)
		if config.easyStore.mode == "s3" {
			log.Printf("[CONFIG] esbucket      = [%s]", config.easyStore.s3Bucket)
		}
	}
	if config.dev.user != "" {
		log.Printf("[CONFIG] devuser       = [%s]", config.dev.user)
		log.Printf("[CONFIG] devrole       = [%s]", config.dev.role)
	}
	if config.dev.fakeBus {
		log.Printf("[CONFIG] ** dev mode bus - event publishing is disabled **")
	}

	return &config
}
