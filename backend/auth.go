package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type jwtClaims struct {
	ComputeID string `json:"computeID"`
	jwt.StandardClaims
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
	if svc.DevAuthUser != "" {
		computingID = svc.DevAuthUser
		log.Printf("INFO: using dev auth user ID: %s", computingID)
	}
	if computingID == "" {
		log.Printf("ERROR: expected auth header not present in request. Not authorized.")
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}

	log.Printf("INFO: generate JWT for %s", computingID)
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := jwtClaims{
		ComputeID: computingID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "libra3",
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
	c.SetCookie("ts2_jwt", signedStr, 10, "/", "", false, false)
	c.SetSameSite(http.SameSiteLaxMode)
	c.Redirect(http.StatusFound, "/granted")
}

// AuthMiddleware is middleware that checks for a user auth token in the
// Authorization header. For now, it does nothing but ensure token presence.
func (svc *serviceContext) authMiddleware(c *gin.Context) {
	log.Printf("Authorize access to %s", c.Request.URL)
	tokenStr, err := getBearerToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		log.Printf("WARNING: authentication failed: [%s]", err.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if tokenStr == "undefined" {
		log.Printf("WARNING: authentication failed; bearer token is undefined")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("INFO: validating JWT auth token...")
	jwtClaims := &jwtClaims{}
	_, jwtErr := jwt.ParseWithClaims(tokenStr, jwtClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(svc.JWTKey), nil
	})
	if jwtErr != nil {
		log.Printf("WARNING: authentication failed; token validation failed: %+v", jwtErr)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("INFO: got valid bearer token: [%s] for %s", tokenStr, jwtClaims.ComputeID)
	c.Set("jwt", tokenStr)
	c.Set("claims", jwtClaims)
	c.Next()
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
	components := strings.Split(strings.Join(strings.Fields(authorization), " "), " ")

	// must have two components, the first of which is "Bearer", and the second a non-empty token
	if len(components) != 2 || components[0] != "Bearer" || components[1] == "" {
		return "", fmt.Errorf("Invalid Authorization header: [%s]", authorization)
	}

	return components[1], nil
}
