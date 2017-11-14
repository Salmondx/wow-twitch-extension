package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/salmondx/wow-twitch-extension/bnet"

	"github.com/salmondx/wow-twitch-extension/cache"
	"github.com/salmondx/wow-twitch-extension/model"
	"github.com/salmondx/wow-twitch-extension/storage"
)

// CharacterService allows to view currently added WoW characters,
// add new characters, delete character and get a detailed info of selected character
type CharacterService interface {
	// Get short characters info. Returns empty slice if no characters
	List(streamerID string) ([]*model.CharacterInfo, error)
	// Add new character to storage. If character exists with such realm - name pair, error is thrown
	Add(streamerID, region, realm, name string) error
	// Delete character from storage
	Delete(streamerID, region, realm, name string) error
	// Retrieve full character profile
	Profile(streamerID, region, realm, name string) (*model.Character, error)
}

// CachableCharacterService implements CharacterService interface
// It caches and stores data in db. If not found, searches data in Bnet.API
type CachableCharacterService struct {
	cache      cache.Cache
	storage    storage.CharacterRepository
	bnetClient *bnet.Client
}

func New(cache cache.Cache, storage storage.CharacterRepository, bnetClient *bnet.Client) *CachableCharacterService {
	return &CachableCharacterService{
		cache:      cache,
		storage:    storage,
		bnetClient: bnetClient,
	}
}

func (s *CachableCharacterService) List(streamerID string) ([]*model.CharacterInfo, error) {
	if streamerID == "" {
		return nil, errors.New("StreamerID can not be empty")
	}
	characters, err := s.cache.List(streamerID)
	if err != nil {
		log.Printf("[WARN] Can't retrive characters list from cache: %s. %v", streamerID, err)
		characters, err = s.storage.List(streamerID)
		if err != nil {
			return nil, err
		}
		err = s.cache.AddCharacters(streamerID, characters)
		if err != nil {
			log.Printf("[ERROR] %v", err)
		}
	}
	if characters == nil {
		characters = make([]*model.CharacterInfo, 0)
	}
	return characters, nil
}

func (s *CachableCharacterService) Add(streamerID, region, realm, name string) error {
	if missingRequiredParameters(streamerID, region, realm, name) {
		return errors.New("StreamerID, realm or name can not be empty")
	}
	bnetProfile, err := s.bnetClient.GetCharacterProfile(region, realm, name)
	if err != nil {
		return err
	}
	profile := Convert(bnetProfile)
	charInfo := model.CharacterInfo{
		CharIcon: profile.CharIcon,
		Class:    profile.Class,
		Name:     profile.Name,
		Realm:    profile.Realm,
		Region:   profile.Region,
	}

	// Trying to search characters for duplications
	characters, err := s.List(streamerID)
	if err != nil {
		return err
	}
	if characters != nil {
		for _, character := range characters {
			if character.Realm == realm && character.Name == name {
				return model.CharacterDuplicateError{fmt.Sprintf("Character with name %s on realm %s already exists", name, realm)}
			}
		}
	}
	// Add new character
	err = s.storage.Add(streamerID, &charInfo)
	if err != nil {
		return err
	}
	// update characters in cache
	err = s.cache.Update(streamerID, profile)
	if err != nil {
		log.Printf("[ERROR] Can not save profile in cache %s. %v", streamerID, err)
	}
	return nil
}

func (s *CachableCharacterService) Delete(streamerID, region, realm, name string) error {
	if missingRequiredParameters(streamerID, region, realm, name) {
		return errors.New("StreamerID, realm or name can not be empty")
	}
	err := s.storage.Delete(streamerID, region, realm, name)
	if err != nil {
		return err
	}
	err = s.cache.ClearList(streamerID)
	if err != nil {
		log.Printf("[ERROR] Can not delete character from cache: %s, %s - %s. %v", streamerID, realm, name, err)
	}
	return nil
}

func (s *CachableCharacterService) Profile(streamerID, region, realm, name string) (*model.Character, error) {
	if missingRequiredParameters(streamerID, region, realm, name) {
		return nil, errors.New("StreamerID, realm or name can not be empty")
	}
	profile, err := s.cache.GetProfile(streamerID, region, realm, name)
	if err != nil {
		log.Printf("[INFO] %s profile not found in cache (%s - %s). Search bnet.", streamerID, realm, name)
		bnetProfile, err := s.bnetClient.GetCharacterProfile(region, realm, name)
		if err != nil {
			return nil, err
		}
		profile = Convert(bnetProfile)
		err = s.cache.AddProfile(streamerID, profile)
		if err != nil {
			log.Printf("Can not update cache for %s. %v", streamerID, err)
		}
	}
	return profile, nil
}

func missingRequiredParameters(streamerID, region, realm, name string) bool {
	return streamerID == "" || realm == "" || name == "" || region == ""
}
