package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
)

type eventContext struct {
	DevMode     bool
	BusName     string
	EventSource string
	Bus         uvalibrabus.UvaBus
}

// services that require jwt authorization and the necessary token
type protectedServices struct {
	DepositAuthURL string
	ORCID          orcidConfig
	UserServiceURL string
	JWT            string
}

func (ps *protectedServices) refreshJWT(key string) error {
	if ps.JWT != "" {
		userClaims := jwtClaims{}
		_, jwtErr := jwt.ParseWithClaims(ps.JWT, &userClaims, func(token *jwt.Token) (any, error) {
			return []byte(key), nil
		})
		if jwtErr != nil {
			log.Printf("INFO: existing protected services jwt is not valid; generate another: %s", jwtErr.Error())
		} else {
			log.Printf("INFO: protected services jwt already exists and is valid")
			return nil
		}
	}

	log.Printf("INFO: generate jwt for protected services")
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		Issuer:    "libra-web",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, jwtErr := token.SignedString([]byte(key))
	if jwtErr != nil {
		return jwtErr
	}
	ps.JWT = signedStr
	return nil
}

// serviceContext contains common data used by all handlers
type serviceContext struct {
	Version         string
	EtdURL          string
	TimeFormat      string
	HTTPClient      *http.Client
	EasyStore       uvaeasystore.EasyStore
	Events          eventContext
	AuditQueryURL   string
	IndexURL        string
	MetricsQueryURL string
	Protected       protectedServices
	JWTKey          string
	Namespace       string
	UVAWhiteList    []*net.IPNet
	Dev             devConfig
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
	Version        string         `json:"version"`
	Namespace      libraNamespace `json:"namespace"`
	Licenses       []license      `json:"licenses"`
	Languages      []language     `json:"languages"`
	Visibility     []visibility   `json:"visibility"`
	Programs       []etdProgram   `json:"programs"`
	Degrees        []etdDegree    `json:"degrees"`
	ORCIDClientURL string         `json:"orcid"`
}

// InitializeService sets up the service context for all API handlers
func initializeService(version string, cfg *configData) *serviceContext {
	ctx := serviceContext{
		Version:         version,
		TimeFormat:      "2006-01-02T15:04:05Z",
		Dev:             cfg.dev,
		JWTKey:          cfg.jwtKey,
		Namespace:       cfg.namespace,
		EtdURL:          cfg.etdURL,
		AuditQueryURL:   cfg.auditQueryURL,
		IndexURL:        cfg.indexURL,
		MetricsQueryURL: cfg.metricsQueryURL,
	}

	// URLs to external service that just JWT protection
	ctx.Protected.DepositAuthURL = cfg.depositAuthURL
	ctx.Protected.ORCID = cfg.orcid
	ctx.Protected.UserServiceURL = cfg.userServiceURL

	log.Printf("INFO: initialize uva ip whitelist")
	wlBytes, err := os.ReadFile("./data/ipwhitelist.txt")
	if err != nil {
		log.Fatalf("read ipwhitelist failed: %s", err.Error())
	}
	for ip := range strings.SplitSeq(string(wlBytes), "\n") {
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

	log.Printf("INFO: init jwt for protected services")
	if err := ctx.Protected.refreshJWT(ctx.JWTKey); err != nil {
		log.Fatalf("unable to generate protected services jwt: %s", err.Error())
	}

	log.Printf("INFO: configure easystore")
	config := uvaeasystore.ProxyConfigImpl{
		ServiceEndpoint: cfg.easyStoreProxy,
		Log:             log.Default(),
	}
	es, err := uvaeasystore.NewEasyStoreProxy(config)
	if err != nil {
		log.Fatalf("create easystore failed: %s", err.Error())
	}
	ctx.EasyStore = es
	log.Printf("INFO: easystore configured")

	ctx.Events.DevMode = cfg.dev.fakeBus
	ctx.Events.BusName = cfg.busName
	ctx.Events.EventSource = cfg.eventSourceName
	if cfg.dev.fakeBus == false {
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

func (svc *serviceContext) getVersion(c *gin.Context) {
	vMap := svc.lookupVersion()
	c.JSON(http.StatusOK, vMap)
}

func (svc *serviceContext) getConfig(c *gin.Context) {
	verInfo := svc.lookupVersion()
	ver := fmt.Sprintf("v%s-%s", verInfo["version"], verInfo["build"])
	resp := configResponse{
		Version: ver,
		Namespace: libraNamespace{
			Label:     "LibraETD",
			Namespace: svc.Namespace},
		ORCIDClientURL: svc.Protected.ORCID.clientURL,
	}

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

	if esErr := svc.EasyStore.Check(); esErr != nil {
		hcMap["easystore"] = hcResp{Healthy: false, Message: esErr.Error()}
	} else {
		hcMap["easystore"] = hcResp{Healthy: true}
	}

	if _, err := svc.sendGetRequest(fmt.Sprintf("%s/version", svc.Protected.UserServiceURL)); err != nil {
		hcMap["user-ws"] = hcResp{Healthy: false, Message: err.Message}
	} else {
		hcMap["user-ws"] = hcResp{Healthy: true}
	}

	if _, err := svc.sendGetRequest(fmt.Sprintf("%s/version", svc.Protected.ORCID.serviceURL)); err != nil {
		hcMap["orcid"] = hcResp{Healthy: false, Message: err.Message}
	} else {
		hcMap["orcid"] = hcResp{Healthy: true}
	}

	if _, err := svc.sendGetRequest(fmt.Sprintf("%s/version", svc.Protected.DepositAuthURL)); err != nil {
		hcMap["deposit-auth"] = hcResp{Healthy: false, Message: err.Message}
	} else {
		hcMap["deposit-auth"] = hcResp{Healthy: true}
	}

	c.JSON(http.StatusOK, hcMap)
}

func (svc *serviceContext) mintUserJWT(user *UserDetails) (string, error) {
	log.Printf("INFO: generate JWT for %s", user.ComputeID)
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := jwtClaims{
		UserDetails: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "libra-web",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, jwtErr := token.SignedString([]byte(svc.JWTKey))
	if jwtErr != nil {
		return "", jwtErr
	}
	return signedStr, nil
}

func (svc *serviceContext) lookupComputeID(c *gin.Context) {
	computeID := c.Param("cid")
	log.Printf("INFO: lookup compute id [%s]", computeID)
	if err := svc.Protected.refreshJWT(svc.JWTKey); err != nil {
		log.Printf("ERROR: unable to refresh protected services jwt: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	url := fmt.Sprintf("%s/user/%s?auth=%s", svc.Protected.UserServiceURL, computeID, svc.Protected.JWT)
	resp, userErr := svc.sendGetRequest(url)
	if userErr != nil {
		log.Printf("INFO: lookup info user [%s] failed: %s", computeID, userErr.Message)
		c.String(http.StatusNotFound, fmt.Sprintf("%s not found", computeID))
		return
	}

	var jsonResp userServiceResp
	if err := json.Unmarshal(resp, &jsonResp); err != nil {
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
	log.Printf("INFO: lookup orcid for compute id [%s]", computeID)
	orcidDetail, err := svc.doOrcidLookup(computeID)
	if err != nil {
		log.Printf("ERROR: orcid lookup failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// no err and nil data means there is no orcid for this user
	if orcidDetail == nil {
		log.Printf("INFO: no orcid for user %s", computeID)
		c.String(http.StatusNotFound, fmt.Sprintf("%s not found", computeID))
		return
	}

	c.JSON(http.StatusOK, *orcidDetail)
}

func (svc *serviceContext) doOrcidLookup(computeID string) (*OrcidDetails, error) {
	if err := svc.Protected.refreshJWT(svc.JWTKey); err != nil {
		return nil, fmt.Errorf("unable to refresh protected services jwt: %s", err.Error())
	}
	url := fmt.Sprintf("%s/cid/%s?auth=%s", svc.Protected.ORCID.serviceURL, computeID, svc.Protected.JWT)
	payload, userErr := svc.sendGetRequest(url)
	if userErr != nil {
		if userErr.StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("%s", userErr.Message)
	}

	resp := OrcidDetailsResponse{}
	if err := json.Unmarshal(payload, &resp); err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}

	log.Printf("INFO: parsed orcid response [%+v] for cid [%s]", resp, computeID)
	if len(resp.Details) == 0 {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	return &resp.Details[0], nil
}

func (svc *serviceContext) sendGetRequest(url string) ([]byte, *RequestError) {
	return svc.sendRequest("GET", url, nil)
}

func (svc *serviceContext) sendPostRequest(url string, payload any) ([]byte, *RequestError) {
	return svc.sendRequest("POST", url, payload)
}

func (svc *serviceContext) sendRequest(verb string, url string, payload any) ([]byte, *RequestError) {
	log.Printf("INFO: %s request: %s", verb, url)
	startTime := time.Now()

	var req *http.Request
	if verb == "POST" && payload != nil {
		b, _ := json.Marshal(payload)
		req, _ = http.NewRequest("POST", url, bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	} else {
		req, _ = http.NewRequest(verb, url, nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36")
	}

	rawResp, rawErr := svc.HTTPClient.Do(req)
	resp, err := handleAPIResponse(url, rawResp, rawErr)
	elapsedNanoSec := time.Since(startTime)
	elapsedMS := int64(elapsedNanoSec / time.Millisecond)

	if err != nil {
		if err.StatusCode == 404 {
			log.Printf("INFO: 404 response from %s %s. Elapsed Time: %d (ms)", verb, url, elapsedMS)
		} else {
			log.Printf("ERROR: Failed response from %s %s - %d:%s. Elapsed Time: %d (ms)", verb, url, err.StatusCode, err.Message, elapsedMS)
		}
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

func parseDate(dateStr string) time.Time {
	date, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return date
	}
	date, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("WARNING: unable to parse date %s: %s", dateStr, err.Error())
	}
	return date
}
