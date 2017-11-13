package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/salmondx/wow-twitch-extension/model"
)

type CacheClient struct {
	pool *redis.Pool
}

func New(address string) *CacheClient {
	return &CacheClient{
		&redis.Pool{
			MaxIdle:     10,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", address)
			},
		},
	}
}

// 24 hours
const expirationTimeout = 24 * 60 * 60

func (cache *CacheClient) List(streamerID string) ([]*model.CharacterInfo, error) {
	if streamerID == "" {
		return nil, errors.New("StreamerID can not be empty")
	}
	conn := cache.pool.Get()
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("GET", streamerID))
	if err != nil {
		return nil, fmt.Errorf("Can't retrieve cache data: %s. Reason: %v", streamerID, err)
	}
	var characters []*model.CharacterInfo
	if data == nil {
		return characters, nil
	}
	json.Unmarshal(data, &characters)
	return characters, nil
}

func (cache *CacheClient) AddCharacters(streamerID string, characterInfos []*model.CharacterInfo) error {
	if streamerID == "" {
		return errors.New("StreamerID can not be empty")
	}
	conn := cache.pool.Get()
	defer conn.Close()

	bytes, err := json.Marshal(characterInfos)
	if err != nil {
		return fmt.Errorf("Can not serialize characters for %s. Reason: %v", streamerID, err)
	}
	conn.Send("MULTI")
	conn.Send("SET", streamerID, bytes)
	conn.Send("EXPIRE", streamerID, expirationTimeout)
	_, err = conn.Do("EXEC")
	if err != nil {
		return fmt.Errorf("Can not save characters for %s. Reason: %v", streamerID, err)
	}
	return nil
}

func (cache *CacheClient) GetProfile(streamerID, realm, name string) (*model.Character, error) {
	if streamerID == "" || realm == "" || name == "" {
		return nil, errors.New("StreamerID, realm or name can not be empty")
	}

	conn := cache.pool.Get()
	defer conn.Close()

	key := createProfileKey(streamerID, realm, name)
	bytes, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, fmt.Errorf("Can't get profile for %s. Reason: %v", streamerID, err)
	}

	var character model.Character
	err = json.Unmarshal(bytes, &character)
	if err != nil {
		return nil, fmt.Errorf("Can't serialize profile for %s. Reason: %v", streamerID, err)
	}
	return &character, nil
}

func (cache *CacheClient) AddProfile(streamerID string, character *model.Character) error {
	if streamerID == "" || character == nil {
		return errors.New("StreamerID or character can not be null or empty")
	}

	conn := cache.pool.Get()
	defer conn.Close()

	key := createProfileKey(streamerID, character.Realm, character.Name)
	data, err := json.Marshal(character)
	if err != nil {
		return fmt.Errorf("Can't serialize profile for %s. Reason: %v", streamerID, err)
	}
	conn.Send("MULTI")
	conn.Send("SET", key, data)
	conn.Send("EXPIRE", key, expirationTimeout)
	_, err = conn.Do("EXEC")
	if err != nil {
		return fmt.Errorf("Can't save profile for %s. Reason: %v", streamerID, err)
	}
	return nil
}

func (cache *CacheClient) Update(streamerID string, character *model.Character) error {
	if streamerID == "" || character == nil {
		return errors.New("StreamerID or character can not be null or empty")
	}
	characters, err := cache.List(streamerID)
	if err == nil {
		charInfo := model.CharacterInfo{
			CharIcon: character.CharIcon,
			Class:    character.Class,
			Name:     character.Name,
			Realm:    character.Realm,
		}
		characters = append(characters, &charInfo)
		err = cache.AddCharacters(streamerID, characters)
		if err != nil {
			log.Printf("Can not update characters for %s: %v", streamerID, err)
		}
	}
	err = cache.AddProfile(streamerID, character)
	if err != nil {
		log.Printf("Can not update profile for %s. Reason: %v", streamerID, err)
	}
	return nil
}

func (cache *CacheClient) ClearList(streamerID string) error {
	if streamerID == "" {
		return errors.New("StreamerID can not be empty")
	}
	conn := cache.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", streamerID)
	if err != nil {
		return fmt.Errorf("Can not delete %s list from cache. Reason: %v", streamerID, err)
	}
	return nil
}
