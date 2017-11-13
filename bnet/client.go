package bnet

import "net/http"
import "fmt"
import "encoding/json"
import "github.com/salmondx/wow-twitch-extension/model"

type Item struct {
	ID           int
	Name         string
	ItemLevel    int
	Icon         string
	Enchantments Enchantments `json:"tooltipParams"`
}

type Enchantments struct {
	Gem0        int
	Gem1        int
	Gem2        int
	Gem3        int
	Gem4        int
	Enchantment int
}

type Items struct {
	Head     Item
	Neck     Item
	Shoulder Item
	Back     Item
	Chest    Item
	Tabard   Item
	Wrist    Item
	Hands    Item
	Waist    Item
	Legs     Item
	Feet     Item
	Finger1  Item
	Finger2  Item
	Trinket1 Item
	Trinket2 Item
	MainHand Item
	OffHand  Item
}

type CharacterProfile struct {
	Name      string
	Realm     string
	Region    string
	Class     int
	Thumbnail string
	Items     Items
}

type Client struct {
	secret string
}

const battleNetURL = "https://eu.api.battle.net/wow/character/%s/%s?fields=items&locale=en_GB&apikey=%s"

// New creates a new Battle.Net client
func New(secret string) *Client {
	return &Client{secret}
}

// GetCharacterProfile retrieves character profile from Battle.Net API by character name and realm
// If not found, then error is thrown
func (c *Client) GetCharacterProfile(realm, name string) (*CharacterProfile, error) {
	resp, err := http.Get(fmt.Sprintf(battleNetURL, realm, name, c.secret))
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve profile for %s - %s. Reason: %v", realm, name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, model.CharacterNotFound{fmt.Sprintf("Character not found: %s - %s", realm, name)}
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid return code: %d", resp.StatusCode)
	}
	var characterProfile CharacterProfile

	err = json.NewDecoder(resp.Body).Decode(&characterProfile)
	if err != nil {
		return nil, fmt.Errorf("Can't deserialize response: %v", err)
	}
	return &characterProfile, nil
}
