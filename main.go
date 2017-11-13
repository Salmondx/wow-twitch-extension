package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/salmondx/wow-twitch-extension/bnet"
	"github.com/salmondx/wow-twitch-extension/cache"
	"github.com/salmondx/wow-twitch-extension/model"
	"github.com/salmondx/wow-twitch-extension/storage"

	"github.com/salmondx/wow-twitch-extension/service"
)

type RequestParameters struct {
	Realm      string
	Name       string
	StreamerID string
	Role       string
}

type ErrorMessage struct {
	Code   int
	Reason string
}

// Partial commit, rewrite using DI
var (
	clientSecret      = os.Getenv("CLIENT_SECRET")
	characterNotFound = ErrorMessage{100, "No character with such name and realm pair"}
	characterLimit    = ErrorMessage{101, "Character limit reached. Delete character to add a new one"}
	unknownError      = ErrorMessage{102, "Unknown error occurred"}
	missingParameters = ErrorMessage{103, "Required parameters were not provided"}
)

func requestHandler(h func(string, RequestParameters, service.CharacterService) (interface{}, error),
	characterService service.CharacterService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		queryParams := r.URL.Query()
		realm := queryParams.Get("realm")
		name := queryParams.Get("name")

		parameters := RequestParameters{
			Realm:      realm,
			Name:       name,
			StreamerID: "123144",
			Role:       "streamer",
		}
		data, err := h(r.Method, parameters, characterService)
		if err != nil {
			errorMessage, status := handleError(err)
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(errorMessage)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}
}

func handleError(err error) (ErrorMessage, int) {
	var errorMessage ErrorMessage
	var status int
	switch err.(type) {
	case model.CharacterNotFound:
		log.Printf("[INFO] Character not found: %v", err)
		errorMessage = characterNotFound
		status = http.StatusNotFound
	case model.CharacterLimitError:
		log.Printf("[INFO] Character limit reached. %v", err)
		errorMessage = characterLimit
		status = http.StatusConflict
	case model.ParametersNotProvided:
		log.Printf("[INFO] Parameters were not provided. %v", err)
		errorMessage = missingParameters
		status = http.StatusBadRequest
	default:
		log.Printf("[ERROR] %v", err)
		errorMessage = unknownError
		status = http.StatusInternalServerError
	}
	return errorMessage, status
}

func profileHandler(method string, parameters RequestParameters, characterService service.CharacterService) (interface{}, error) {
	if parameters.StreamerID == "" || parameters.Name == "" || parameters.Realm == "" {
		return nil, model.ParametersNotProvided{"Missing parameters"}
	}
	log.Printf("[INFO] Profile for %v - %v", parameters.Realm, parameters.Name)
	profile, err := characterService.Profile(parameters.StreamerID, parameters.Realm, parameters.Name)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func main() {
	bnetClient := bnet.New(clientSecret)
	redisCache := cache.New("localhost:6379")
	dynamoStorage, _ := storage.New()
	cacheService := service.New(redisCache, dynamoStorage, bnetClient)

	// http.HandleFunc("/profile", handler)
	http.HandleFunc("/profile", requestHandler(profileHandler, cacheService))
	log.Println("Starting server")

	log.Fatal(http.ListenAndServe(":8080", nil))

}
