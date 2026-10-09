// stubsvc is a local stand-in for the external HTTP services libra-web calls
// (user-ws, deposit-auth, orcid, audit and metrics). Dev use only; see dev/README.md
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	port := flag.Int("port", 8095, "port to listen on")
	first := flag.String("first", "Local", "first name returned for every user")
	last := flag.String("last", "Developer", "last name returned for every user")
	flag.Parse()

	writeJSON := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	}

	// user-ws: GET /user/:cid
	http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		cid := strings.TrimPrefix(r.URL.Path, "/user/")
		log.Printf("user lookup %s", cid)
		writeJSON(w, map[string]any{
			"status":  200,
			"message": "",
			"user": map[string]any{
				"cid":          cid,
				"display_name": fmt.Sprintf("%s %s", *first, *last),
				"first_name":   *first,
				"last_name":    *last,
				"email":        fmt.Sprintf("%s@localhost", cid),
				"affiliation":  []string{"student"},
				"department":   []string{},
				"description":  []string{},
				"title":        []string{},
				"office":       []string{},
				"phone":        []string{},
			},
		})
	})
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"version": "local-stub"})
	})
	http.HandleFunc("/audits", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []any{})
	})
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"views": 0, "files": []any{}})
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("no stub for %s %s", r.Method, r.URL.String())
		http.NotFound(w, r)
	})

	log.Printf("stub services listening on 127.0.0.1:%d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", *port), nil))
}
