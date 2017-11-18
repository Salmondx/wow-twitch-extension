package service

import (
	"testing"

	"github.com/salmondx/wow-twitch-extension/bnet"
)

func TestConverter(t *testing.T) {
	bnetProfile := &bnet.CharacterProfile{
		Class:     2,
		Realm:     "Soulflayer",
		Name:      "Salmond",
		Region:    "eu",
		Thumbnail: "soulflayer/51/64174899-avatar.jpg",
		Items: bnet.Items{
			Head: bnet.Item{
				ID:   142982,
				Name: "Fearless Combatant's Plate Helm of the Quickblade",
				Icon: "inv_helm_plate_legionhonor_d_01",
			},
			Neck: bnet.Item{
				ID:   133767,
				Name: "Pendant of the Stormforger",
				Icon: "inv_7_0raid_necklace_14a",
				Enchantments: bnet.Enchantments{
					Gem0:        1235,
					Gem1:        12445,
					Enchantment: 123567,
				},
			},
		},
	}

	actual := Convert(bnetProfile)
	if actual.Name != "Salmond" {
		t.Errorf("Name not equals")
	}

	if actual.Realm != "Soulflayer" {
		t.Errorf("Realm not equals")
	}

	if actual.Region != "eu" {
		t.Errorf("Region not equals")
	}

	if actual.Class != "Paladin" {
		t.Errorf("Class not equals")
	}

	if len(actual.Items) != 2 {
		t.Errorf("Not enought items")
	}

	head := actual.Items[0]
	if head.Name != "Fearless Combatant's Plate Helm of the Quickblade" {
		t.Error()
	}
}

func TestItemConverter(t *testing.T) {
	region := "eu"
	var tests = []struct {
		item         bnet.Item
		itemType     string
		wowhead      string
		enchantments string
		icon         string
	}{
		{item: bnet.Item{
			ID:        133767,
			ItemLevel: 865,
			Name:      "Pendant of the Stormforger",
			Icon:      "inv_7_0raid_necklace_14a",

			Enchantments: bnet.Enchantments{
				Gem0:        1235,
				Gem1:        12445,
				Enchantment: 123567,
			},
		},
			wowhead:  "item=133767&gems=1235:12445&ench=123567",
			itemType: "Neck",
			icon:     "https://render-eu.worldofwarcraft.com/icons/36/inv_7_0raid_necklace_14a.jpg",
		},
		{item: bnet.Item{
			ID:        133767,
			ItemLevel: 14,
			Name:      "t6 shoulders",
			Icon:      "inv_7_0raid_necklace_14a",
		},
			wowhead:  "item=133767",
			itemType: "Shoulders",
			icon:     "https://render-eu.worldofwarcraft.com/icons/36/inv_7_0raid_necklace_14a.jpg",
		},
	}

	for _, tt := range tests {
		item := convItem(tt.item, tt.itemType, region)
		if item.Type != tt.itemType {
			t.Errorf("Wrong type")
		}
		if item.Name != tt.item.Name {
			t.Error("Names not equal")
		}
		if item.IconURL != tt.icon {
			t.Errorf("IconURL not equal: %s", item.IconURL)
		}
		if item.ItemLvl != tt.item.ItemLevel {
			t.Error("Wront item lvl")
		}
		if item.DescriptionURL != tt.wowhead {
			t.Error("Wrong wowhead link")
		}
	}
}

func TestClassConverter(t *testing.T) {
	var tests = []struct {
		in  int
		out string
	}{
		{1, "Warrior"},
		{2, "Paladin"},
		{3, "Hunter"},
		{4, "Rogue"},
		{5, "Priest"},
		{6, "Death Knight"},
		{7, "Shaman"},
		{8, "Mage"},
		{9, "Warlock"},
		{10, "Monk"},
		{11, "Druid"},
		{12, "Demon Hunter"},
	}

	for _, tt := range tests {
		className := classByIndex(tt.in)
		if className != tt.out {
			t.Errorf("%s not equals %s by index %d", className, tt.out, tt.in)
		}
	}
}
