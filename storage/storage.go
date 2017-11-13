package storage

import "github.com/salmondx/wow-twitch-extension/model"

// CharacterRepository is a permanent storage of a streamer's characters
type CharacterRepository interface {
	// List retrieves characters list from database
	List(streamerID string) ([]*model.CharacterInfo, error)
	// Add adds new character to database
	Add(streamerID string, character *model.CharacterInfo) error
	// Delete deletes character from database
	Delete(streamerID, realm, name string) error
}
