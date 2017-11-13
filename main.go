package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/salmondx/wow-twitch-extension/bnet"
	"github.com/salmondx/wow-twitch-extension/cache"
	"github.com/salmondx/wow-twitch-extension/service"
	"github.com/salmondx/wow-twitch-extension/storage"
)

// Partial commit, rewrite using DI
var (
	clientSecret     = os.Getenv("CLIENT_SECRET")
	bnetClient       = bnet.New(clientSecret)
	redisCache       = cache.New("localhost:6379")
	dynamoStorage, _ = storage.New()
	cacheService     = service.New(redisCache, dynamoStorage, bnetClient)
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parameters := r.URL.Query()
	server := parameters.Get("server")
	name := parameters.Get("name")

	if server == "" || name == "" {
		log.Println("[WARN] Missing server or name query parameter")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing server or name query parameter"))
		return
	}
	log.Printf("[INFO] Profile for: %v - %v\n", server, name)
	profile, _ := cacheService.Profile("12314", server, name)
	json.NewEncoder(w).Encode(&profile)
}

func main() {
	http.HandleFunc("/profile", handler)
	log.Println("Starting server")

	log.Fatal(http.ListenAndServe(":8080", nil))

}
