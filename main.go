package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"

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
	Region     string
	Role       string
}

type ErrorMessage struct {
	Code   int
	Reason string
}

type HttpError struct {
	S    string
	Code int
}

func (e HttpError) Error() string {
	return e.S
}

// Partial commit, rewrite using DI
var (
	twitchSecret     []byte
	clientSecret     = os.Getenv("CLIENT_SECRET")
	redisAddress     = os.Getenv("REDIS_ADDRESS")
	badRequest       = HttpError{"Missing required parameters", http.StatusBadRequest}
	methodNotAllowed = HttpError{"Method not allowed", http.StatusMethodNotAllowed}
	wrongRole        = HttpError{"Only streamer is allowed to update characters list", http.StatusForbidden}

	characterNotFound = ErrorMessage{100, "No character with such name and realm pair"}
	characterLimit    = ErrorMessage{101, "Character limit reached. Delete character to add a new one"}
	unknownError      = ErrorMessage{102, "Unknown error occurred. Try again later"}
	missingParameters = ErrorMessage{103, "Required parameters were not provided"}
)

func requestHandler(h func(string, RequestParameters, service.CharacterService) (interface{}, error),
	characterService service.CharacterService,
	successCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
			w.Header().Add("Access-Control-Allow-Headers", "Authorization")
			return
		}

		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "Authorization")
		w.Header().Add("Access-Control-Allow-Origin", "*")

		rawToken := r.Header.Get("Authorization")
		log.Printf("%s\n", rawToken)
		if rawToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Unexpected signing method")
			}
			return twitchSecret, nil
		})
		if err != nil {
			log.Printf("[INFO] Unauthorized: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Printf("[INFO] Invalid token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		streamerID, ok := claims["channel_id"].(string)
		role, roleOk := claims["role"].(string)
		if !ok || !roleOk {
			log.Printf("[WARN] Can't get channel_id or role property")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		queryParams := r.URL.Query()
		realm := queryParams.Get("realm")
		name := queryParams.Get("name")
		region := queryParams.Get("region")

		parameters := RequestParameters{
			Realm:      realm,
			Name:       name,
			Region:     region,
			StreamerID: streamerID,
			Role:       role,
		}
		data, err := h(r.Method, parameters, characterService)
		if err != nil {
			errorMessage, status := handleError(err)
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(errorMessage)
			return
		}
		w.WriteHeader(successCode)
		if data != nil {
			json.NewEncoder(w).Encode(data)
		}
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
	case model.CharacterDuplicateError:
		log.Printf("[INFO] Character duplicate: %v", err)
		errorMessage = ErrorMessage{104, err.Error()}
		status = http.StatusConflict
	case HttpError:
		httpErr := err.(HttpError)
		log.Printf("[INFO] %v", httpErr.S)
		errorMessage = ErrorMessage{105, httpErr.S}
		status = httpErr.Code
	default:
		log.Printf("[ERROR] %v", err)
		errorMessage = unknownError
		status = http.StatusInternalServerError
	}
	return errorMessage, status
}

func profileHandler(method string, parameters RequestParameters, characterService service.CharacterService) (interface{}, error) {
	if method != http.MethodGet {
		return nil, methodNotAllowed
	}
	if missingRequiredParameters(parameters) {
		return nil, badRequest
	}
	log.Printf("[INFO] Profile for %v - %v", parameters.Realm, parameters.Name)
	profile, err := characterService.Profile(parameters.StreamerID, parameters.Region, parameters.Realm, parameters.Name)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func listHandler(method string, parameters RequestParameters, chacterService service.CharacterService) (interface{}, error) {
	if method != http.MethodGet {
		return nil, methodNotAllowed
	}
	if parameters.StreamerID == "" {
		return nil, badRequest
	}
	log.Printf("[INFO] Character list for %s", parameters.StreamerID)
	characters, err := chacterService.List(parameters.StreamerID)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func addCharacterHandler(method string, parameters RequestParameters, chacterService service.CharacterService) (interface{}, error) {
	if method != http.MethodPost {
		return nil, methodNotAllowed
	}
	if parameters.Role != "broadcaster" {
		return nil, wrongRole
	}
	if missingRequiredParameters(parameters) {
		return nil, badRequest
	}

	log.Printf("[INFO] Adding character for %s: %s - %s", parameters.StreamerID, parameters.Realm, parameters.Name)
	err := chacterService.Add(parameters.StreamerID, parameters.Region, parameters.Realm, parameters.Name)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func deleteCharacterHandler(method string, parameters RequestParameters, chacterService service.CharacterService) (interface{}, error) {
	if method != http.MethodDelete {
		return nil, methodNotAllowed
	}
	if parameters.Role != "broadcaster" {
		return nil, wrongRole
	}
	if missingRequiredParameters(parameters) {
		return nil, badRequest
	}

	log.Printf("[INFO] Deleting character for %s: %s - %s", parameters.StreamerID, parameters.Realm, parameters.Name)
	err := chacterService.Delete(parameters.StreamerID, parameters.Region, parameters.Realm, parameters.Name)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func missingRequiredParameters(parameters RequestParameters) bool {
	return parameters.StreamerID == "" || parameters.Realm == "" || parameters.Name == "" || parameters.Region == ""
}

func main() {
	if clientSecret == "" {
		log.Fatalln("Battle net client secret can not be null or empty! Provide it via CLIENT_SECRET environment variable")
	}
	if redisAddress == "" {
		log.Fatalln("Redis address can not be null or empty. Provide it via REDIS_ADDRESS environment variable")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalln("JWT Secret can not be null or empty. Provide it via JWT_SECRET environment variable")
	}
	s, err := base64.StdEncoding.DecodeString(jwtSecret)
	if err != nil {
		log.Fatalf("Can't decode JWT Secret from base64: %v", err)
	}
	twitchSecret = []byte(s)

	bnetClient := bnet.New(clientSecret)
	redisCache := cache.New(redisAddress)
	dynamoStorage, _ := storage.New()
	cacheService := service.New(redisCache, dynamoStorage, bnetClient)

	http.HandleFunc("/profile", requestHandler(profileHandler, cacheService, http.StatusOK))
	http.HandleFunc("/list", requestHandler(listHandler, cacheService, http.StatusOK))
	http.HandleFunc("/list/add", requestHandler(addCharacterHandler, cacheService, http.StatusCreated))
	http.HandleFunc("/list/delete", requestHandler(deleteCharacterHandler, cacheService, http.StatusNoContent))
	log.Println("Starting server")

	log.Fatal(http.ListenAndServe(":8080", nil))

}
