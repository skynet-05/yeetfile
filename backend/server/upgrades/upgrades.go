package upgrades

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"time"
	"yeetfile/backend/utils"
	"yeetfile/shared"
)

var upgrades *shared.Upgrades

func GetAllUpgrades() *shared.Upgrades {
	return upgrades
}

// GetUpgradeExpiration returns a point in time in the future that the user's
// selected vault upgrade should expire.
func GetUpgradeExpiration(upgrade shared.Upgrade, quantity int) (time.Time, error) {
	if !upgrade.IsVaultUpgrade {
		return time.Time{}, errors.New("requested upgrade exp for non-vault upgrade")
	}

	if upgrade.Annual {
		// Years * quantity purchased
		return AddDate(1*quantity, 0), nil
	} else {
		// Months * quantity purchased
		return AddDate(0, 1*quantity), nil
	}
}

func GetVaultUpgrades(annual bool, upgrades []*shared.Upgrade) []*shared.Upgrade {
	var result []*shared.Upgrade
	for _, upgrade := range upgrades {
		if upgrade.Annual != annual {
			continue
		}

		result = append(result, upgrade)
	}

	return result
}

func GetUpgradeByTag(tag string, upgrades *shared.Upgrades) (shared.Upgrade, error) {
	for _, upgrade := range upgrades.SendUpgrades {
		if upgrade.Tag == tag {
			return *upgrade, nil
		}
	}

	for _, upgrade := range upgrades.VaultUpgrades {
		if upgrade.Tag == tag {
			return *upgrade, nil
		}
	}

	return shared.Upgrade{}, errors.New("upgrade not found")
}

func init() {
	upgradesJson := os.Getenv("YEETFILE_UPGRADES_JSON")
	if len(upgradesJson) > 0 {
		err := json.Unmarshal([]byte(upgradesJson), &upgrades)
		if err != nil {
			log.Fatalln("Error reading upgrades json:", err)
		}
	} else {
		upgrades = &shared.Upgrades{
			SendUpgrades:  []*shared.Upgrade{},
			VaultUpgrades: []*shared.Upgrade{},
		}
		return
	}

	finalizeUpgrades := func(subUpgrades []*shared.Upgrade, isVaultUpgrade bool) {
		for _, upgrade := range subUpgrades {
			if len(upgrade.Tag) == 0 {
				utils.LogStruct(upgrade)
				log.Fatalln("Missing upgrade tag")
			}

			upgrade.ReadableBytes = shared.ReadableFileSize(upgrade.Bytes)
			upgrade.IsVaultUpgrade = isVaultUpgrade
		}

		sort.Slice(subUpgrades, func(i, j int) bool {
			return subUpgrades[i].Price < subUpgrades[j].Price
		})
	}

	finalizeUpgrades(upgrades.SendUpgrades, false)
	finalizeUpgrades(upgrades.VaultUpgrades, true)

	if len(upgrades.SendUpgrades) > 0 || len(upgrades.VaultUpgrades) > 0 {
		log.Printf("-- Loaded %d send upgrades and %d vault upgrades\n",
			len(upgrades.SendUpgrades),
			len(upgrades.VaultUpgrades))
	}
}
