package service

import (
	"errors"
	"log"

	"github.com/salmondx/wow-twitch-extension/bnet"

	"github.com/salmondx/wow-twitch-extension/cache"
	"github.com/salmondx/wow-twitch-extension/model"
	"github.com/salmondx/wow-twitch-extension/storage"
)

// CharacterService allows to view currently added WoW characters,
// add new characters, delete character and get a detailed info of selected character
type CharacterService interface {
	// Get short characters info
	List(streamerID string) ([]*model.CharacterInfo, error)
	// Add new character to storage
	Add(streamerID, realm, name string) error
	// Delete character from storage
	Delete(streamerID, realm, name string) error
	// Retrieve full character profile
	Profile(streamerID, realm, name string) (*model.Character, error)
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
	if err == nil {
		log.Printf("[WARN] Can't retrive characters list from cache: %s. %v", streamerID, err)
		characters, err = s.storage.List(streamerID)
		if err != nil {
			return nil, err
		}
	}
	return characters, nil
}

func (s *CachableCharacterService) Add(streamerID, realm, name string) error {
	if streamerID == "" || realm == "" || name == "" {
		return errors.New("StreamerID, realm or name can not be empty")
	}
	bnetProfile, err := s.bnetClient.GetCharacterProfile(realm, name)
	if err != nil {
		return err
	}
	profile := Convert(bnetProfile)
	charInfo := model.CharacterInfo{
		CharIcon: profile.CharIcon,
		Class:    profile.Class,
		Name:     profile.Name,
		Realm:    profile.Realm,
	}
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

func (s *CachableCharacterService) Delete(streamerID, realm, name string) error {
	if streamerID == "" || realm == "" || name == "" {
		return errors.New("StreamerID, realm or name can not be empty")
	}
	err := s.storage.Delete(streamerID, realm, name)
	if err != nil {
		return err
	}
	return nil
}

func (s *CachableCharacterService) Profile(streamerID, realm, name string) (*model.Character, error) {
	if streamerID == "" || realm == "" || name == "" {
		return nil, errors.New("StreamerID, realm or name can not be empty")
	}
	profile, err := s.cache.GetProfile(streamerID, realm, name)
	if err != nil {
		log.Printf("[INFO] %s profile not found in cache (%s - %s). Search bnet.", streamerID, realm, name)
		bnetProfile, err := s.bnetClient.GetCharacterProfile(realm, name)
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
