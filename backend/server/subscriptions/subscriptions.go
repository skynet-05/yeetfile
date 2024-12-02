package subscriptions

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"time"
	"yeetfile/backend/utils"
)

type SubDuration string

const (
	SubMonth SubDuration = "month"
	SubYear  SubDuration = "year"
)

type Product struct {
	Tag         string      `json:"tag"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Price       int64       `json:"price"`
	Duration    SubDuration `json:"duration"`

	SendGB     int `json:"send_gb"`
	SendGBReal int64

	StorageGB     int `json:"storage_gb"`
	StorageGBReal int64

	BTCPayLink string `json:"btcpay_link"`
}

var Products []*Product

// membershipDurationFunctionMap maps membership order types to a function for
// generating the correct end date for the user's membership
var membershipDurationFunctionMap = map[SubDuration]func(quantity int) time.Time{
	SubMonth: func(quantity int) time.Time { return AddDate(0, 1*quantity) },
	SubYear:  func(quantity int) time.Time { return AddDate(1*quantity, 0) },
}

// GetSubscriptionExpiration returns a point in time in the future that the
// user's selected subscription should expire.
func GetSubscriptionExpiration(duration SubDuration, quantity int) (time.Time, error) {
	expFn, ok := membershipDurationFunctionMap[duration]
	if !ok {
		return time.Time{}, errors.New("invalid sub duration")
	}

	return expFn(quantity), nil
}

func GetProducts(durationFilter SubDuration) []Product {
	var result []Product
	for _, product := range Products {
		if len(durationFilter) > 0 && product.Duration != durationFilter {
			continue
		}

		result = append(result, *product)
	}

	return result
}

func GetProductByTag(tag string) (Product, error) {
	for _, product := range Products {
		if product.Tag == tag {
			return *product, nil
		}
	}

	return Product{}, errors.New("product not found")
}

func init() {
	productsJson := os.Getenv("YEETFILE_PRODUCTS_JSON")
	if len(productsJson) > 0 {
		err := json.Unmarshal([]byte(productsJson), &Products)
		if err != nil {
			panic(err)
		}
	}

	for _, product := range Products {
		product.SendGBReal = int64(product.SendGB * 1000 * 1000 * 1000)
		product.StorageGBReal = int64(product.StorageGB * 1000 * 1000 * 1000)
	}

	sort.Slice(Products, func(i, j int) bool {
		return Products[i].Price < Products[j].Price
	})

	if len(Products) > 0 {
		log.Println("Loaded products:")
		for _, product := range Products {
			utils.LogStruct(product)
		}
	}
}
