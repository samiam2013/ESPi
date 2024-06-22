package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/samiam2013/basicauth"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load ./.env: %v", err)
	}
	creds := os.Getenv("BASIC_AUTH_CREDS")
	if strings.TrimSpace(creds) == "" {
		log.Fatalf("Need BASIC_AUTH_CREDS env var, none set")
	}
	mapCreds := make(map[string]string)
	pairs := strings.Split(creds, ",")
	for _, pair := range pairs {
		strings := strings.Split(pair, ":")
		if len(strings) != 2 {
			log.Fatalf("Basic auth pair longer than 2 after split on ':'")
		}
		mapCreds[strings[0]] = strings[1]
	}
	basicAuth, err := basicauth.Builder(mapCreds, basicauth.WithUnsafeHTTP())
	if err != nil {
		log.Fatalf("Failed to build basicauth middleware: %v", err)
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", basicAuth(getHandler))
	mux.HandleFunc("POST /", basicAuth(postHandler))

	_ = http.ListenAndServe(":8080", mux)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("get data here"))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("write data here"))
}
