package service

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/salmondx/wow-twitch-extension/bnet"
	"github.com/salmondx/wow-twitch-extension/model"
)

const iconPlaceholderURL = "https://render-%s.worldofwarcraft.com/icons/36/%s.jpg"
const charIconPlaceholderURL = "https://render-%s.worldofwarcraft.com/character/%s"

const wowheadURL = "item=%d"
const wowheadSpellURL = "spell=%d"

// Convert converts profile from Battle.Net API to a required object
func Convert(bnetProfile *bnet.CharacterProfile) *model.Character {
	extensionProfile := model.Character{}
	extensionProfile.Name = bnetProfile.Name
	extensionProfile.Realm = bnetProfile.Realm
	extensionProfile.Region = bnetProfile.Region
	extensionProfile.Class = classByIndex(bnetProfile.Class)
	extensionProfile.CharIcon = fmt.Sprintf(charIconPlaceholderURL, bnetProfile.Region, bnetProfile.Thumbnail)
	extensionProfile.Items = getItems(bnetProfile.Items, bnetProfile.Region)
	extensionProfile.Specs = getSpecs(bnetProfile.Talents, bnetProfile.Region)
	extensionProfile.ArenaRating = getArenaRating(bnetProfile.ArenaRating)
	return &extensionProfile
}

func getSpecs(bnetTalents []bnet.SpecTalents, region string) []model.Spec {
	specs := make([]model.Spec, 0)
	for _, bnetSpec := range bnetTalents {
		if len(bnetSpec.Talents) == 0 {
			continue
		}
		spec := model.Spec{}
		spec.Selected = bnetSpec.Selected

		bnetSpecInfo := getSpecInfo(bnetSpec.Talents)
		spec.Name = bnetSpecInfo.Name
		spec.Order = bnetSpecInfo.Order
		spec.IconURL = fmt.Sprintf(iconPlaceholderURL, region, bnetSpecInfo.Icon)

		talents := make([]model.Talent, 0)

		for _, bnetTalent := range bnetSpec.Talents {
			talent := model.Talent{}
			talent.Tier = bnetTalent.Tier

			spell := model.Spell{}
			spell.Description = bnetTalent.Spell.Description
			spell.ID = bnetTalent.Spell.ID
			spell.IconURL = fmt.Sprintf(iconPlaceholderURL, region, bnetTalent.Spell.Icon)
			spell.Name = bnetTalent.Spell.Name
			spell.DescriptionURL = fmt.Sprintf(wowheadSpellURL, bnetTalent.Spell.ID)

			talent.Spell = spell
			talents = append(talents, talent)
		}
		spec.Talents = talents
		specs = append(specs, spec)
	}
	return specs
}

func getSpecInfo(bnetTalents []bnet.Talents) bnet.Spec {
	for _, bnetTalent := range bnetTalents {
		spec := bnetTalent.Spec
		if spec.Name != "" {
			return spec
		}
	}
	return bnet.Spec{}
}

func getArenaRating(bnetArena bnet.ArenaRating) []model.ArenaRating {
	arenaRating := make([]model.ArenaRating, 3)

	brackets := bnetArena.Brackets

	arenaRating[0] = model.ArenaRating{
		Type:         "2v2",
		Rating:       brackets.TwoPlayers.Rating,
		SeasonPlayed: brackets.TwoPlayers.SeasonPlayed,
		SeasonWon:    brackets.TwoPlayers.SeasonWon,
		SeasonLost:   brackets.TwoPlayers.SeasonLost,
	}

	arenaRating[1] = model.ArenaRating{
		Type:         "3v3",
		Rating:       brackets.ThreePlayers.Rating,
		SeasonPlayed: brackets.ThreePlayers.SeasonPlayed,
		SeasonWon:    brackets.ThreePlayers.SeasonWon,
		SeasonLost:   brackets.ThreePlayers.SeasonLost,
	}

	arenaRating[2] = model.ArenaRating{
		Type:         "RBG",
		Rating:       brackets.RBG.Rating,
		SeasonPlayed: brackets.RBG.SeasonPlayed,
		SeasonWon:    brackets.RBG.SeasonWon,
		SeasonLost:   brackets.RBG.SeasonLost,
	}
	return arenaRating
}

func getItems(bnetItems bnet.Items, region string) []model.Item {
	items := make([]model.Item, 0)

	reflectValue := reflect.ValueOf(bnetItems)

	for i := 0; i < reflectValue.NumField(); i++ {
		name := reflectValue.Type().Field(i).Name
		item := reflectValue.Field(i).Interface().(bnet.Item)
		if item.Name == "" {
			continue
		}

		items = append(items, convItem(item, name, region))
	}
	return items
}

func convItem(bnetItem bnet.Item, itemType string, region string) model.Item {
	item := model.Item{}
	item.Type = itemType
	item.Name = bnetItem.Name
	item.ItemLvl = bnetItem.ItemLevel
	item.IconURL = fmt.Sprintf(iconPlaceholderURL, region, bnetItem.Icon)
	wowheadURL := fmt.Sprintf(wowheadURL, bnetItem.ID)
	enchantmentsURL := genEnchURL(bnetItem)
	if enchantmentsURL != "" {
		wowheadURL = wowheadURL + "&" + enchantmentsURL
	}
	item.DescriptionURL = wowheadURL
	return item
}

func genEnchURL(bnetItem bnet.Item) string {
	var buffer bytes.Buffer
	notSeen := true
	if bnetItem.Enchantments.Gem0 != 0 {
		buffer.WriteString(addGem(bnetItem.Enchantments.Gem0, notSeen))
		notSeen = false
	}
	if bnetItem.Enchantments.Gem1 != 0 {
		buffer.WriteString(addGem(bnetItem.Enchantments.Gem1, notSeen))
		notSeen = false
	}
	if bnetItem.Enchantments.Gem2 != 0 {
		buffer.WriteString(addGem(bnetItem.Enchantments.Gem2, notSeen))
		notSeen = false
	}
	if bnetItem.Enchantments.Gem3 != 0 {
		buffer.WriteString(addGem(bnetItem.Enchantments.Gem3, notSeen))
		notSeen = false
	}
	if bnetItem.Enchantments.Gem4 != 0 {
		buffer.WriteString(addGem(bnetItem.Enchantments.Gem4, notSeen))
		notSeen = false
	}
	if bnetItem.Enchantments.Enchantment != 0 {
		if notSeen {
			buffer.WriteString(fmt.Sprintf("ench=%d", bnetItem.Enchantments.Enchantment))
		} else {
			buffer.WriteString(fmt.Sprintf("&ench=%d", bnetItem.Enchantments.Enchantment))
		}
	}
	return buffer.String()
}

func addGem(id int, first bool) string {
	if first {
		return fmt.Sprintf("gems=%d", id)
	}
	return fmt.Sprintf(":%d", id)
}

func classByIndex(idx int) string {
	var class string
	switch idx {
	case 1:
		class = "Warrior"
	case 2:
		class = "Paladin"
	case 3:
		class = "Hunter"
	case 4:
		class = "Rogue"
	case 5:
		class = "Priest"
	case 6:
		class = "Death Knight"
	case 7:
		class = "Shaman"
	case 8:
		class = "Mage"
	case 9:
		class = "Warlock"
	case 10:
		class = "Monk"
	case 11:
		class = "Druid"
	case 12:
		class = "Demon Hunter"
	}
	return class
}
