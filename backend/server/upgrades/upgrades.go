package upgrades

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"time"
	"yeetfile/backend/utils"
)

type UpgradeDuration string

const (
	DurationMonth UpgradeDuration = "month"
	DurationYear  UpgradeDuration = "year"
)

type Upgrade struct {
	Tag         string          `json:"tag"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       int64           `json:"price"`
	Duration    UpgradeDuration `json:"duration"`

	SendGB     int `json:"send_gb"`
	SendGBReal int64

	StorageGB     int `json:"storage_gb"`
	StorageGBReal int64

	BTCPayLink string `json:"btcpay_link"`
}

var Upgrades []*Upgrade

// upgradeDurationFnMap maps account upgrade types to a function for
// generating the correct end date for the user's upgrade
var upgradeDurationFnMap = map[UpgradeDuration]func(quantity int) time.Time{
	DurationMonth: func(quantity int) time.Time { return AddDate(0, 1*quantity) },
	DurationYear:  func(quantity int) time.Time { return AddDate(1*quantity, 0) },
}

// GetUpgradeExpiration returns a point in time in the future that the user's
// selected upgrade should expire.
func GetUpgradeExpiration(duration UpgradeDuration, quantity int) (time.Time, error) {
	expFn, ok := upgradeDurationFnMap[duration]
	if !ok {
		return time.Time{}, errors.New("invalid sub duration")
	}

	return expFn(quantity), nil
}

func GetUpgrades(durationFilter UpgradeDuration) []Upgrade {
	var result []Upgrade
	for _, upgrade := range Upgrades {
		if len(durationFilter) > 0 && upgrade.Duration != durationFilter {
			continue
		}

		result = append(result, *upgrade)
	}

	return result
}

func GetUpgradeByTag(tag string) (Upgrade, error) {
	for _, upgrade := range Upgrades {
		if upgrade.Tag == tag {
			return *upgrade, nil
		}
	}

	return Upgrade{}, errors.New("upgrade not found")
}

func init() {
	upgradesJson := os.Getenv("YEETFILE_UPGRADES_JSON")
	if len(upgradesJson) > 0 {
		err := json.Unmarshal([]byte(upgradesJson), &Upgrades)
		if err != nil {
			panic(err)
		}
	}

	for _, product := range Upgrades {
		product.SendGBReal = int64(product.SendGB * 1000 * 1000 * 1000)
		product.StorageGBReal = int64(product.StorageGB * 1000 * 1000 * 1000)
	}

	sort.Slice(Upgrades, func(i, j int) bool {
		return Upgrades[i].Price < Upgrades[j].Price
	})

	if len(Upgrades) > 0 {
		log.Println("Loaded upgrades:")
		for _, upgrade := range Upgrades {
			utils.LogStruct(upgrade)
		}
	}
}
