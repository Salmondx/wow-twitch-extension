package model

// Item is a full description of a currently equipped item by type
type Item struct {
	Type           string
	Name           string
	ItemLvl        int
	IconURL        string
	DescriptionURL string `json:",omitempty"`
}

type Spell struct {
	ID             int
	Name           string
	IconURL        string
	Description    string
	DescriptionURL string
}

type Talent struct {
	Tier  int
	Spell Spell
}

type Spec struct {
	Selected bool
	Name     string
	IconURL  string
	Order    int
	Talents  []Talent
}

type ArenaRating struct {
	Type         string
	Rating       int
	SeasonPlayed int
	SeasonWon    int
	SeasonLost   int
}

// Character is a full description of a WoW character with items
type Character struct {
	Name        string
	Realm       string
	Class       string
	Region      string
	CharIcon    string
	ItemLvl     int
	Guild       string
	Items       []Item
	Specs       []Spec
	ArenaRating []ArenaRating
}

// CharacterInfo is a short description of a WoW character, without items
type CharacterInfo struct {
	Name     string
	Realm    string
	Region   string
	Class    string
	CharIcon string
	Guild    string
	ItemLvl  int
}
