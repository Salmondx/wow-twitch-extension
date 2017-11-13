package cache

import "github.com/salmondx/wow-twitch-extension/model"

// Cache is an interface for caching characters list and full character profiles
type Cache interface {
	List(streamerID string) ([]*model.CharacterInfo, error)
	AddCharacters(streamerID string, characterInfos []*model.CharacterInfo) error
	GetProfile(streamerID, realm, name string) (*model.Character, error)
	AddProfile(streamerID string, character *model.Character) error
	Update(streamerID string, character *model.Character) error
	ClearList(streamerID string) error
}

func createProfileKey(streamerID, realm, name string) string {
	return streamerID + ":" + realm + ":" + name
}
