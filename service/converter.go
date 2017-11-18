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

// Convert converts profile from Battle.Net API to a required object
func Convert(bnetProfile *bnet.CharacterProfile) *model.Character {
	extensionProfile := model.Character{}
	extensionProfile.Name = bnetProfile.Name
	extensionProfile.Realm = bnetProfile.Realm
	extensionProfile.Region = bnetProfile.Region
	extensionProfile.Class = classByIndex(bnetProfile.Class)
	extensionProfile.CharIcon = fmt.Sprintf(charIconPlaceholderURL, bnetProfile.Region, bnetProfile.Thumbnail)
	extensionProfile.Items = getItems(bnetProfile.Items, bnetProfile.Region)
	return &extensionProfile
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
