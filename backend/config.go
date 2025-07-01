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

type devConfig struct {
	user    string
	role    string
	fakeBus bool
}

type orcidConfig struct {
	serviceURL string
	clientURL  string
}

type configData struct {
	port            int
	edtURL          string
	userService     userServiceCfg
	orcid           orcidConfig
	auditQueryURL   string
	jwtKey          string
	easyStoreProxy  string
	namespace       string
	busName         string
	eventSourceName string
	indexURL        string
	dev             devConfig
}

func getConfiguration() *configData {
	var config configData
	flag.IntVar(&config.port, "port", 8080, "Port to offer service on")
	flag.StringVar(&config.edtURL, "etdurl", "https://libra-web-dev.internal.lib.virginia.edu", "URL for the LibraETD service")
	flag.StringVar(&config.jwtKey, "jwtkey", "", "JWT signature key")
	flag.StringVar(&config.userService.URL, "userws", "", "URL for the user service")
	flag.StringVar(&config.auditQueryURL, "auditqueryurl", "", "Query URL for the audit service")

	// ORCID ID:
	// * getorcidurl is backed request to get a users ORCID ID
	// * orcidurl is the url the ORCID client used to manage ORCID connection
	flag.StringVar(&config.orcid.serviceURL, "getorcidurl", "", "GET orcid for user service")
	flag.StringVar(&config.orcid.clientURL, "orcidurl", "", "GET orcid for user service")

	// dev mode
	flag.StringVar(&config.dev.user, "devuser", "", "Authorized computing id for dev")
	flag.StringVar(&config.dev.role, "devrole", "user", "Role for dev user")
	flag.BoolVar(&config.dev.fakeBus, "devbus", false, "bus dev mode (no events sent out)")

	// easystore cfg
	flag.StringVar(&config.easyStoreProxy, "esproxy", "", "EasyStore proxy")

	// namespace
	flag.StringVar(&config.namespace, "namespace", "libraetd", "Namespace for work processing")

	// search index
	flag.StringVar(&config.indexURL, "index", "", "Search index URL")

	// event bus
	flag.StringVar(&config.busName, "busname", "", "Event bus name")
	flag.StringVar(&config.eventSourceName, "eventsrc", "", "Event source name")

	flag.Parse()

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
	if config.easyStoreProxy == "" {
		log.Fatal("Parameter esproxy is required")
	}
	if config.indexURL == "" {
		log.Fatal("Parameter index is required")
	}

	log.Printf("[CONFIG] port          = [%d]", config.port)
	log.Printf("[CONFIG] edturl        = [%s]", config.edtURL)
	log.Printf("[CONFIG] userws        = [%s]", config.userService.URL)
	log.Printf("[CONFIG] getorcidurl   = [%s]", config.orcid.serviceURL)
	log.Printf("[CONFIG] orcidurl      = [%s]", config.orcid.clientURL)
	log.Printf("[CONFIG] auditqueryurl = [%s]", config.auditQueryURL)
	log.Printf("[CONFIG] namespace     = [%s]", config.namespace)
	log.Printf("[CONFIG] eventsrc      = [%s]", config.eventSourceName)
	log.Printf("[CONFIG] busname       = [%s]", config.busName)
	log.Printf("[CONFIG] index         = [%s]", config.indexURL)
	log.Printf("[CONFIG] esproxy       = [%s]", config.easyStoreProxy)

	if config.dev.user != "" {
		log.Printf("[CONFIG] devuser       = [%s]", config.dev.user)
		log.Printf("[CONFIG] devrole       = [%s]", config.dev.role)
	}
	if config.dev.fakeBus {
		log.Printf("[CONFIG] ** dev mode bus - event publishing is disabled **")
	}

	return &config
}
