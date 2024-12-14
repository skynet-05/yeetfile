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
	"yeetfile/shared/constants"
)

var upgrades []*shared.Upgrade

// upgradeDurationFnMap maps account upgrade types to a function for
// generating the correct end date for the user's upgrade
var upgradeDurationFnMap = map[constants.UpgradeDuration]func(quantity int) time.Time{
	constants.DurationMonth: func(quantity int) time.Time { return AddDate(0, 1*quantity) },
	constants.DurationYear:  func(quantity int) time.Time { return AddDate(1*quantity, 0) },
}

func GetLoadedUpgrades() []*shared.Upgrade {
	return upgrades
}

// GetUpgradeExpiration returns a point in time in the future that the user's
// selected upgrade should expire.
func GetUpgradeExpiration(duration constants.UpgradeDuration, quantity int) (time.Time, error) {
	expFn, ok := upgradeDurationFnMap[duration]
	if !ok {
		return time.Time{}, errors.New("invalid sub duration")
	}

	return expFn(quantity), nil
}

func GetUpgrades(durationFilter constants.UpgradeDuration, upgrades []*shared.Upgrade) []shared.Upgrade {
	var result []shared.Upgrade
	for _, upgrade := range upgrades {
		if len(durationFilter) > 0 && upgrade.Duration != durationFilter {
			continue
		}

		result = append(result, *upgrade)
	}

	return result
}

func GetUpgradeByTag(tag string, upgrades []*shared.Upgrade) (shared.Upgrade, error) {
	for _, upgrade := range upgrades {
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
			panic(err)
		}
	}

	for _, product := range upgrades {
		product.SendGBReal = int64(product.SendGB * 1000 * 1000 * 1000)
		product.StorageGBReal = int64(product.StorageGB * 1000 * 1000 * 1000)
	}

	sort.Slice(upgrades, func(i, j int) bool {
		return upgrades[i].Price < upgrades[j].Price
	})

	if len(upgrades) > 0 {
		log.Println("Loaded upgrades:")
		for _, upgrade := range upgrades {
			utils.LogStruct(upgrade)
		}
	}
}
