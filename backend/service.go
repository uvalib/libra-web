package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
)

type eventContext struct {
	DevMode     bool
	BusName     string
	EventSource string
	Bus         uvalibrabus.UvaBus
}

// serviceContext contains common data used by all handlers
type serviceContext struct {
	Version       string
	HTTPClient    *http.Client
	EasyStore     uvaeasystore.EasyStore
	Events        eventContext
	UserService   userServiceCfg
	OrcidService  orcidServiceCfg
	JWTKey        string
	DevAuthUser   string
	Namespace     string
	UVAWhiteList  []*net.IPNet
	AuditQueryURL string
}

// RequestError contains http status code and message for a failed HTTP request
type RequestError struct {
	StatusCode int
	Message    string
}

type language struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type visibility struct {
	Value     string `json:"value"`
	Label     string `json:"label"`
	AdminOnly bool   `json:"adminOnly"`
	License   *struct {
		URL   string `json:"url"`
		Label string `json:"label"`
	} `json:"license,omitempty"`
}

type license struct {
	Value     string `json:"value"`
	URL       string `json:"url"`
	Label     string `json:"label"`
	AdminOnly bool   `json:"adminOnly"`
}

type libraNamespace struct {
	Label     string `json:"label"`
	Namespace string `json:"namespace"`
}

type etdDegree struct {
	Type   string `json:"type"` // optional or sis
	SISKey string `json:"sisKey,omitempty"`
	Degree string `json:"degree"`
}

type etdProgram struct {
	Type    string `json:"type"` // optional or sis
	SISKey  string `json:"sisKey,omitempty"`
	Program string `json:"program"`
}

type configResponse struct {
	Version    string         `json:"version"`
	Namespace  libraNamespace `json:"namespace"`
	Licenses   []license      `json:"licenses"`
	Languages  []language     `json:"languages"`
	Visibility []visibility   `json:"visibility"`
	Programs   []etdProgram   `json:"programs"`
	Degrees    []etdDegree    `json:"degrees"`
}

// InitializeService sets up the service context for all API handlers
func initializeService(version string, cfg *configData) *serviceContext {
	ctx := serviceContext{Version: version,
		JWTKey:       cfg.jwtKey,
		UserService:  cfg.userService,
		OrcidService: cfg.orcidService,
		DevAuthUser:  cfg.devAuthUser,
		Namespace:    cfg.namespace,
	}

	log.Printf("INFO: initialize uva ip whitelist")
	wlBytes, err := os.ReadFile("./data/ipwhitelist.txt")
	if err != nil {
		log.Fatalf("read ipwhitelist failed: %s", err.Error())
	}
	for _, ip := range strings.Split(string(wlBytes), "\n") {
		cleanIP := strings.TrimSpace(ip)
		if cleanIP != "" {
			_, ipnet, ipErr := net.ParseCIDR(cleanIP)
			if ipErr != nil {
				log.Fatalf("unable to parse cidr %s: %s", cleanIP, ipErr.Error())
			}
			ctx.UVAWhiteList = append(ctx.UVAWhiteList, ipnet)
		}
	}

	log.Printf("INFO: create HTTP client...")
	defaultTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 600 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	ctx.HTTPClient = &http.Client{
		Transport: defaultTransport,
		Timeout:   5 * time.Second,
	}
	log.Printf("INFO: HTTP Client created")

	err = ctx.checkUserServiceJWT()
	if err != nil {
		log.Fatal(fmt.Sprintf("unable to generate user service jwt: %s", err.Error()))
	}

	log.Printf("INFO: configure easystore")
	// NOTE: easystore will disable the event bus if the bus name is bank. Do this for devBus mode
	busName := cfg.busName
	if cfg.devBus {
		log.Printf("INFO: bus is in dev mode; set blank name for easystore config")
		busName = ""
	}
	if cfg.easyStore.mode == "sqlite" {
		config := uvaeasystore.DatastoreSqliteConfig{
			DataSource: path.Join(cfg.easyStore.dbDir, cfg.easyStore.dbFile),
			// Log:        log.Default(),
			BusName:    busName,
			SourceName: cfg.eventSourceName,
		}
		es, err := uvaeasystore.NewEasyStore(config)
		if err != nil {
			log.Fatalf("create easystore failed: %s", err.Error())
		}
		ctx.EasyStore = es
	} else if cfg.easyStore.mode == "postgres" {
		config := uvaeasystore.DatastorePostgresConfig{
			DbHost:     cfg.easyStore.dbHost,
			DbPort:     cfg.easyStore.dbPort,
			DbName:     cfg.easyStore.dbName,
			DbUser:     cfg.easyStore.dbUser,
			DbPassword: cfg.easyStore.dbPass,
			DbTimeout:  cfg.easyStore.dbTimeout,
			BusName:    busName,
			SourceName: cfg.eventSourceName,
			Log:        log.Default(),
		}
		es, err := uvaeasystore.NewEasyStore(config)
		if err != nil {
			log.Fatalf("create easystore failed: %s", err.Error())
		}
		ctx.EasyStore = es
	} else if cfg.easyStore.mode == "s3" {
		config := uvaeasystore.DatastoreS3Config{
			Bucket:     cfg.easyStore.s3Bucket,
			DbHost:     cfg.easyStore.dbHost,
			DbPort:     cfg.easyStore.dbPort,
			DbName:     cfg.easyStore.dbName,
			DbUser:     cfg.easyStore.dbUser,
			DbPassword: cfg.easyStore.dbPass,
			DbTimeout:  cfg.easyStore.dbTimeout,
			BusName:    busName,
			SourceName: cfg.eventSourceName,
			Log:        log.Default(),
		}
		es, err := uvaeasystore.NewEasyStore(config)
		if err != nil {
			log.Fatalf("create easystore failed: %s", err.Error())
		}
		ctx.EasyStore = es
	} else {
		log.Fatalf("easystore mode [%s] is not supported", cfg.easyStore.mode)
	}
	log.Printf("INFO: easystore configured")

	ctx.Events.DevMode = cfg.devBus
	ctx.Events.BusName = cfg.busName
	ctx.Events.EventSource = cfg.eventSourceName
	if cfg.devBus {
		log.Printf("INFO: bus is in dev mode - log events instead of sending to bus")
	} else {
		log.Printf("INFO: configure event bus [%s] with source [%s]", cfg.busName, cfg.eventSourceName)
		busCfg := uvalibrabus.UvaBusConfig{
			Source:  cfg.eventSourceName,
			BusName: cfg.busName,
		}
		bus, err := uvalibrabus.NewUvaBus(busCfg)
		if err != nil {
			log.Fatalf("create event bus failed: %s", err.Error())
		}
		ctx.Events.Bus = bus
	}

	ctx.AuditQueryURL = cfg.auditQueryURL

	return &ctx
}

func (svc *serviceContext) publishEvent(eventName, namespace, oid string) {
	log.Printf("INFO: publish event %s for %s in namespace %s", eventName, oid, namespace)
	ev := uvalibrabus.UvaBusEvent{
		EventName:  eventName,
		Namespace:  namespace,
		Identifier: oid,
	}
	if svc.Events.DevMode {
		log.Printf("INFO: dev mode send %+v to bus [%s] with source [%s]", ev, svc.Events.BusName, svc.Events.EventSource)
	} else {
		err := svc.Events.Bus.PublishEvent(&ev)
		if err != nil {
			log.Printf("ERROR: unable to publish event %s %s - %s: %s", eventName, namespace, oid, err.Error())
		}
	}
}

func (svc *serviceContext) checkUserServiceJWT() error {
	if svc.UserService.JWT != "" {
		jwtClaims := &jwtClaims{}
		_, jwtErr := jwt.ParseWithClaims(svc.UserService.JWT, jwtClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(svc.JWTKey), nil
		})
		if jwtErr != nil {
			log.Printf("INFO: existing user ws jwt is not valid; generate another: %s", jwtErr.Error())
		}
		log.Printf("INFO: user ws jwt already exists and is valid")
		return nil
	}

	log.Printf("INFO: generate jwt for userws")
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Issuer:    "libra3",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, jwtErr := token.SignedString([]byte(svc.JWTKey))
	if jwtErr != nil {
		return jwtErr
	}
	svc.UserService.JWT = signedStr
	return nil
}

func (svc *serviceContext) getVersion(c *gin.Context) {
	vMap := svc.lookupVersion()
	c.JSON(http.StatusOK, vMap)
}

func (svc *serviceContext) getConfig(c *gin.Context) {
	verInfo := svc.lookupVersion()
	ver := fmt.Sprintf("v%s-%s", verInfo["version"], verInfo["build"])
	resp := configResponse{Version: ver, Namespace: libraNamespace{Label: "LibraETD", Namespace: svc.Namespace}}

	err := loadETDConfig(&resp)
	if err != nil {
		log.Printf("ERROR: unable to load config: %s", err.Error())
	}

	log.Printf("INFO: load languages")
	bytes, err := os.ReadFile("./data/languages.json")
	if err != nil {
		log.Printf("ERROR: unable to load languages: %s", err.Error())
	} else {
		err = json.Unmarshal(bytes, &resp.Languages)
		if err != nil {
			log.Printf("ERROR: unable to parse languages: %s", err.Error())
		}
	}

	// TODO all the data files need to be edited to remove OA . ETD
	log.Printf("INFO: load licenses")
	bytes, err = os.ReadFile("./data/licenses.json")
	if err != nil {
		log.Printf("ERROR: unable to load licenses: %s", err.Error())
	} else {
		err = json.Unmarshal(bytes, &resp.Licenses)
		if err != nil {
			log.Printf("ERROR: unable to parse licenses: %s", err.Error())
		}
	}

	log.Printf("INFO: load visibility")
	bytes, err = os.ReadFile("./data/visibility.json")
	if err != nil {
		log.Printf("ERROR: unable to load visibility: %s", err.Error())
	} else {
		err = json.Unmarshal(bytes, &resp.Visibility)
		if err != nil {
			log.Printf("ERROR: unable to parse visibility: %s", err.Error())
		}
	}

	c.JSON(http.StatusOK, resp)
}

func loadETDConfig(cfg *configResponse) error {
	log.Printf("INFO: load opt programs")
	var optPrograms []string
	bytes, err := os.ReadFile("./data/opt_programs.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &optPrograms)
	if err != nil {
		return err
	}
	for _, p := range optPrograms {
		cfg.Programs = append(cfg.Programs, etdProgram{Type: "optional", Program: p})
	}

	log.Printf("INFO: load sis programs")
	var sisPrograms []etdProgram
	bytes, err = os.ReadFile("./data/sis_programs.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &sisPrograms)
	if err != nil {
		return err
	}
	for _, p := range sisPrograms {
		cfg.Programs = append(cfg.Programs, etdProgram{Type: "sis", Program: p.Program, SISKey: p.SISKey})
	}

	log.Printf("INFO: load opt degrees")
	var optDegrees []string
	bytes, err = os.ReadFile("./data/opt_degrees.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &optDegrees)
	if err != nil {
		return err
	}
	for _, d := range optDegrees {
		cfg.Degrees = append(cfg.Degrees, etdDegree{Type: "optional", Degree: d})
	}

	log.Printf("INFO: load sis degrees")
	var sisDegrees []etdDegree
	bytes, err = os.ReadFile("./data/sis_degrees.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &sisDegrees)
	if err != nil {
		return err
	}
	for _, d := range sisDegrees {
		cfg.Degrees = append(cfg.Degrees, etdDegree{Type: "sis", Degree: d.Degree, SISKey: d.SISKey})
	}

	return nil
}

func (svc *serviceContext) lookupVersion() map[string]string {
	build := "unknown"
	// working directory is the bin directory, and build tag is in the root
	files, _ := filepath.Glob("../buildtag.*")
	if len(files) == 1 {
		build = strings.Replace(files[0], "../buildtag.", "", 1)
	}

	vMap := make(map[string]string)
	vMap["version"] = svc.Version
	vMap["build"] = build
	return vMap
}

func (svc *serviceContext) healthCheck(c *gin.Context) {
	type hcResp struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message,omitempty"`
	}
	hcMap := make(map[string]hcResp)
	hcMap["libra3"] = hcResp{Healthy: true}

	_, err := svc.sendGetRequest(fmt.Sprintf("%s/version", svc.UserService.URL))
	if err != nil {
		hcMap["user-ws"] = hcResp{Healthy: false, Message: err.Message}
	} else {
		hcMap["user-ws"] = hcResp{Healthy: true}
	}

	c.JSON(http.StatusOK, hcMap)
}

func (svc *serviceContext) lookupComputeID(c *gin.Context) {
	computeID := c.Param("cid")
	log.Printf("INFO: lookup compute id [%s]", computeID)
	err := svc.checkUserServiceJWT()
	if err != nil {
		log.Printf("ERROR: unable to check user service jwt: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	url := fmt.Sprintf("%s/user/%s?auth=%s", svc.UserService.URL, computeID, svc.UserService.JWT)
	resp, userErr := svc.sendGetRequest(url)
	if userErr != nil {
		log.Printf("INFO: lookup info user [%s] failed: %s", computeID, userErr.Message)
		c.String(http.StatusNotFound, fmt.Sprintf("%s not found", computeID))
		return
	}

	log.Printf("GOT: %s", resp)
	var jsonResp userServiceResp
	err = json.Unmarshal(resp, &jsonResp)
	if err != nil {
		log.Printf("ERROR: unable to parse user serice responce: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, jsonResp.User)
}

// OrcidDetailsResponse is the response from the Orcid service
type OrcidDetailsResponse struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Details []OrcidDetails `json:"results"`
}

// OrcidDetails holds details from the Orcid service
type OrcidDetails struct {
	Orcid string `json:"orcid,omitempty"`
	URI   string `json:"uri,omitempty"`
}

func (svc *serviceContext) lookupOrcidID(c *gin.Context) {
	computeID := c.Param("cid")
	log.Printf("INFO: lookup ORCID for compute id [%s]", computeID)
	err := svc.checkUserServiceJWT()
	if err != nil {
		log.Printf("ERROR: unable to check user service jwt: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	url := svc.OrcidService.GetURL
	// substitute values into url
	url = strings.Replace(url, "{:id}", computeID, 1)
	url = strings.Replace(url, "{:auth}", svc.UserService.JWT, 1)
	payload, userErr := svc.sendGetRequest(url)
	if userErr != nil {
		log.Printf("INFO: lookup ORCID user [%s] failed: %s", computeID, userErr.Message)
		c.String(http.StatusNotFound, fmt.Sprintf("%s not found", computeID))
		return
	}
	resp := OrcidDetailsResponse{}
	err = json.Unmarshal(payload, &resp)
	if err != nil {
		fmt.Printf("ERROR: json unmarshal of OrcidDetailsResponse (%s)\n", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Printf("INFO: located ORCID [%v] for cid [%s]\n", resp, computeID)

	// if we have details, return them
	if len(resp.Details) != 0 {
		c.JSON(http.StatusOK, resp.Details[0])
		return
	}

	fmt.Printf("ERROR: No ORCID details found. API may have changed.\n")
	c.String(http.StatusNotFound, "No ORCID details found")

}

func (svc *serviceContext) sendGetRequest(url string) ([]byte, *RequestError) {
	return svc.sendRequest("GET", url, nil)
}

func (svc *serviceContext) sendRequest(verb string, url string, payload *url.Values) ([]byte, *RequestError) {
	log.Printf("INFO: %s request: %s", verb, url)
	startTime := time.Now()

	var req *http.Request
	if verb == "POST" && payload != nil {
		req, _ = http.NewRequest("POST", url, strings.NewReader(payload.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, _ = http.NewRequest(verb, url, nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	}

	rawResp, rawErr := svc.HTTPClient.Do(req)
	resp, err := handleAPIResponse(url, rawResp, rawErr)
	elapsedNanoSec := time.Since(startTime)
	elapsedMS := int64(elapsedNanoSec / time.Millisecond)

	if err != nil {
		log.Printf("ERROR: Failed response from %s %s - %d:%s. Elapsed Time: %d (ms)",
			verb, url, err.StatusCode, err.Message, elapsedMS)
	} else {
		log.Printf("INFO: Successful response from %s %s. Elapsed Time: %d (ms)", verb, url, elapsedMS)
	}
	return resp, err
}

func handleAPIResponse(logURL string, resp *http.Response, err error) ([]byte, *RequestError) {
	if err != nil {
		status := http.StatusBadRequest
		errMsg := err.Error()
		if strings.Contains(err.Error(), "Timeout") {
			status = http.StatusRequestTimeout
			errMsg = fmt.Sprintf("%s timed out", logURL)
		} else if strings.Contains(err.Error(), "connection refused") {
			status = http.StatusServiceUnavailable
			errMsg = fmt.Sprintf("%s refused connection", logURL)
		}
		return nil, &RequestError{StatusCode: status, Message: errMsg}
	} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		status := resp.StatusCode
		errMsg := string(bodyBytes)
		return nil, &RequestError{StatusCode: status, Message: errMsg}
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	return bodyBytes, nil
}
