package model

// Item is a full description of a currently equipped item by type
type Item struct {
	Type            string
	Name            string
	ItemLvl         int
	IconURL         string
	DescriptionURL  string `json:",omitempty"`
	EnchantmentsURL string `json:",omitempty"`
}

// Character is a full description of a WoW character with items
type Character struct {
	Name     string
	Realm    string
	Class    string
	CharIcon string
	Items    []Item
}

// CharacterInfo is a short description of a WoW character, without items
type CharacterInfo struct {
	Name     string
	Realm    string
	Class    string
	CharIcon string
}
