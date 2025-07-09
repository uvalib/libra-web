package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type jwtClaims struct {
	*UserDetails
	jwt.StandardClaims
}

func (c *jwtClaims) isAdmin() bool {
	return c.Role == "admin"
}

func (c *jwtClaims) isRegistrar() bool {
	return c.Role == "registrar"
}

// UserDetails contains a user response from the user-ws
type UserDetails struct {
	ComputeID   string   `json:"cid"`
	UVAID       string   `json:"uva_id"`
	DisplayName string   `json:"display_name"`
	FirstName   string   `json:"first_name"`
	Initials    string   `json:"initials"`
	LastName    string   `json:"last_name"`
	Description []string `json:"description"`
	Department  []string `json:"department"`
	Title       []string `json:"title"`
	Office      []string `json:"office"`
	Phone       []string `json:"phone"`
	Affiliation []string `json:"affiliation"`
	Email       string   `json:"email"`
	Private     string   `json:"private"`
	Role        string   `json:"role"`
}

type userServiceResp struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	User    UserDetails `json:"user"`
}

type authInfo struct {
	tokenString string
	jwt         *jwtClaims
}

func (svc *serviceContext) authenticate(c *gin.Context) {
	log.Printf("Checking authentication headers...")
	log.Printf("Dump all request headers ==================================")
	for name, values := range c.Request.Header {
		for _, value := range values {
			log.Printf("%s=%s\n", name, value)
		}
	}
	log.Printf("END header dump ===========================================")

	computingID := c.GetHeader("remote_user")
	if svc.Dev.user != "" {
		computingID = svc.Dev.user
		log.Printf("INFO: using dev auth user ID: %s", computingID)
	}
	if computingID == "" {
		log.Printf("ERROR: expected auth header not present in request. Not authorized.")
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}

	log.Printf("INFO: request user info for %s", computingID)
	err := svc.checkUserServiceJWT()
	if err != nil {
		log.Printf("ERROR: unable to check user service jwt: %s", err.Error())
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}
	url := fmt.Sprintf("%s/user/%s?auth=%s", svc.UserService.URL, computingID, svc.UserService.JWT)
	resp, userErr := svc.sendGetRequest(url)
	if userErr != nil {
		log.Printf("ERROR: unable get info for user %s: %s", computingID, userErr.Message)
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}
	var jsonResp userServiceResp
	err = json.Unmarshal(resp, &jsonResp)
	if err != nil {
		log.Printf("ERROR: unable to parse user serice response: %s", err.Error())
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}

	// default to standard user role
	jsonResp.User.Role = "user"

	// if not in dev mode check for membership in libra-admins
	if svc.Dev.user == "" {
		// Membership format: cn=group_name1;cn=group_name2;...
		membershipStr := c.GetHeader("member")
		if strings.Contains(membershipStr, "libra-admins") {
			log.Printf("INFO: user %s is an admin", computingID)
			jsonResp.User.Role = "admin"
		} else if strings.Contains(membershipStr, "libra-optional-reg") {
			log.Printf("INFO: user %s can add deposit registrations", computingID)
			jsonResp.User.Role = "registrar"
		}
	} else {
		switch svc.Dev.role {
		case "admin":
			log.Printf("INFO: dev user %s is an admin", svc.Dev.user)
			jsonResp.User.Role = "admin"
		case "registrar":
			log.Printf("INFO: dev user %s is a registrar", svc.Dev.user)
			jsonResp.User.Role = "registrar"
		}
	}

	log.Printf("INFO: generate JWT for %+v", jsonResp.User)
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := jwtClaims{
		UserDetails: &jsonResp.User,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "libra-web",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, jwtErr := token.SignedString([]byte(svc.JWTKey))
	if jwtErr != nil {
		log.Printf("ERROR: unable to generate JWT for %s: %s", computingID, jwtErr.Error())
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}

	// Set auth info in a cookie the client can read and pass along in future requests
	c.SetCookie("libra3_jwt", signedStr, 10, "/", "", false, false)
	c.SetSameSite(http.SameSiteLaxMode)
	c.Redirect(http.StatusFound, "/signedin")
}

func (svc *serviceContext) checkAuthToken(c *gin.Context) {
	_, err := svc.getAuthFromHeader(c.Request.Header.Get("Authorization"))
	if err != nil {
		log.Printf("INFO: check auth failed: %s", err.Error())
		c.String(http.StatusForbidden, "invalid")
		return
	}
	c.String(http.StatusOK, "valid")
}

func (svc *serviceContext) userMiddleware(c *gin.Context) {
	jwtRequired := true
	if c.Request.Method == "GET" && (strings.Contains(c.Request.URL.Path, "/api/works/oa") || strings.Contains(c.Request.URL.Path, "/api/works/etd")) {
		log.Printf("INFO: public metadata request for %s; jwt not required", c.Request.URL.Path)
		jwtRequired = false
	} else {
		log.Printf("INFO: authorize user access to %s", c.Request.URL.Path)
	}

	auth, err := svc.getAuthFromHeader(c.Request.Header.Get("Authorization"))
	if err != nil {
		if jwtRequired {
			log.Printf("WARNING: authentication failed: %s", err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		log.Printf("INFO: optional auth is not present")
	} else {
		log.Printf("INFO: got valid bearer token: [%s] for %s", auth.tokenString, auth.jwt.ComputeID)
		c.Set("claims", auth.jwt)
	}

	c.Next()
}

func (svc *serviceContext) registrarMiddleware(c *gin.Context) {
	log.Printf("INFO: authorize registrar access to %s", c.Request.URL.Path)
	claims := getJWTClaims(c)
	if claims == nil {
		log.Printf("WARNING: registrar auth failed; no claims present")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("INFO: user %s has role %s", claims.ComputeID, claims.Role)
	if claims.isAdmin() == false && claims.isRegistrar() == false {
		log.Printf("WARNING: registrar auth failed for non-admin, non-registrar user %s", claims.ComputeID)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("INFO: registrar authorization successful")
	c.Next()
}

func (svc *serviceContext) adminMiddleware(c *gin.Context) {
	log.Printf("INFO: authorize admin access to %s", c.Request.URL.Path)
	// NOTE: this middleare happens AFTER user middleware, so the claims
	// will already be set with c.Set("claims", auth.jwt) above.
	// just need to verify here that the claims are for an admin user
	claims := getJWTClaims(c)
	if claims == nil {
		log.Printf("WARNING: admin auth failed; no claims present")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims.isAdmin() == false {
		log.Printf("WARNING: admin auth failed for non-admin user %s", claims.ComputeID)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("INFO: admin authorization successful")
	c.Next()
}

func (svc *serviceContext) getAuthFromHeader(authHeader string) (*authInfo, error) {
	log.Printf("INFO: extract auth token from authorization header")
	tokenStr, err := getBearerToken(authHeader)
	if err != nil {
		return nil, err
	}

	if tokenStr == "undefined" {
		return nil, fmt.Errorf("bearer token is undefined")
	}

	log.Printf("INFO: validating JWT auth token")
	jwtClaims := &jwtClaims{}
	_, jwtErr := jwt.ParseWithClaims(tokenStr, jwtClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(svc.JWTKey), nil
	})
	if jwtErr != nil {
		return nil, fmt.Errorf("token validation failed: %+v", jwtErr)
	}

	auth := authInfo{tokenString: tokenStr, jwt: jwtClaims}
	return &auth, nil
}

func isSignedIn(c *gin.Context) bool {
	_, signedIn := c.Get("claims")
	return signedIn
}

func getJWTClaims(c *gin.Context) *jwtClaims {
	claims, signedIn := c.Get("claims")
	if signedIn == false {
		return nil
	}
	jwtClaims, ok := claims.(*jwtClaims)
	if !ok {
		return nil
	}
	return jwtClaims
}

func getBearerToken(authorization string) (string, error) {
	// must have two components, the first of which is "Bearer", and the second a non-empty token
	components := strings.Split(strings.Join(strings.Fields(authorization), " "), " ")
	if len(components) != 2 || components[0] != "Bearer" || components[1] == "" {
		return "", fmt.Errorf("invalid authorization header: [%s]", authorization)
	}
	return components[1], nil
}
