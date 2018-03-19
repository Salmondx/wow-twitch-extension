package service

import (
	"testing"

	"github.com/salmondx/wow-twitch-extension/bnet"
)

var specTalents = []bnet.SpecTalents{
	bnet.SpecTalents{
		Selected: true,
		Talents: []bnet.Talents{
			bnet.Talents{Tier: 0, Spell: bnet.Spell{ID: 107570, Name: "Storm Bolt", Icon: "warrior_talent_icon_stormbolt", Description: "Hurls your weapon at an enemy, causing 14,348 Physical damage and stunning for 4 sec."}, Spec: bnet.Spec{Name: "Arms", Icon: "ability_warrior_savageblow", Order: 0}},
			bnet.Talents{Tier: 0, Spell: bnet.Spell{ID: 203179, Name: "Opportunity Strikes", Icon: "ability_backstab", Description: "Your melee abilities have up to a 60% chance, based on the target's missing health, to trigger an extra attack that deals 24,080 Physical damage and generates 5 Rage."}, Spec: bnet.Spec{Name: "Arms", Icon: "ability_warrior_savageblow", Order: 0}},
		},
	},
}

func TestConverter(t *testing.T) {
	bnetProfile := &bnet.CharacterProfile{
		Class:     2,
		Realm:     "Soulflayer",
		Name:      "Salmond",
		Region:    "eu",
		Thumbnail: "soulflayer/51/64174899-avatar.jpg",
		Guild:     bnet.Guild{"Test"},
		Items: bnet.Items{
			AverageItemLevelEquipped: 942,
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
		Talents: specTalents,
		ArenaRating: bnet.ArenaRating{
			bnet.Brackets{
				TwoPlayers:   bnet.ArenaStats{1950, 0, 0, 0},
				ThreePlayers: bnet.ArenaStats{0, 0, 0, 0},
				RBG:          bnet.ArenaStats{0, 0, 0, 0},
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

	if actual.Guild != "Test" {
		t.Errorf("Guild not equals")
	}
	if actual.ItemLvl != 942 {
		t.Errorf("Wrong item lvl")
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

	if len(actual.Specs) != 1 {
		t.Fatalf("Spec doesn't converted")
	}

	if len(actual.Specs[0].Talents) != 2 {
		t.Fatalf("Talants doesn't converted")
	}
}

func TestTalentConverter(t *testing.T) {
	converted := getSpecs(specTalents, "eu")
	if len(converted) == 0 {
		t.Fatalf("Failed to convert")
	}
	spec := converted[0]
	if spec.Name != "Arms" {
		t.Errorf("Wrong spec name")
	}

	if !spec.Selected {
		t.Errorf("Spec is not selected")
	}

	if spec.IconURL == "" {
		t.Errorf("Icon is empty")
	}

	if len(spec.Talents) == 0 {
		t.Fatalf("Talents not converted")
	}

	talents := spec.Talents

	if talents[0].Spell.Name != "Storm Bolt" || talents[1].Spell.Name != "Opportunity Strikes" {
		t.Errorf("Wrong spell name")
	}

	if talents[0].Spell.IconURL == "" || talents[1].Spell.IconURL == "" {
		t.Errorf("Spell icon is empty")
	}

	if talents[0].Spell.ID == 0 || talents[1].Spell.ID == 0 {
		t.Errorf("Spell id is empty")
	}
}

func TestArenaConverter(t *testing.T) {
	bnetArena := bnet.ArenaRating{
		bnet.Brackets{
			TwoPlayers:   bnet.ArenaStats{0, 0, 0, 0},
			ThreePlayers: bnet.ArenaStats{0, 0, 0, 0},
			RBG:          bnet.ArenaStats{0, 0, 0, 0},
		},
	}

	rating := getArenaRating(bnetArena)
	if len(rating) != 3 {
		t.Errorf("Failed to convert rating. Length is not 3")
	}

	if rating[0].Type != "2v2" {
		t.Errorf("2v2 not found")
	}

	if rating[1].Type != "3v3" {
		t.Errorf("3v3 not found")
	}

	if rating[2].Type != "RBG" {
		t.Errorf("RBG not found")
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
