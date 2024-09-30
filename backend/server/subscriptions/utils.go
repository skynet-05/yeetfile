package subscriptions

import (
	"errors"
	"strings"
	"time"
)

// GetSubscriptionDuration takes a subscription tag (i.e. "yearly-novice") and
// returns just the duration portion (i.e. "yearly")
func GetSubscriptionDuration(tag string) (string, error) {
	subDuration := tag
	if strings.Contains(subDuration, "-") {
		subDuration = strings.Split(tag, "-")[0]
	}

	if subDuration != monthly && subDuration != yearly {
		return "", errors.New("invalid subscription duration")
	}

	return subDuration, nil
}

// GetSubscriptionType takes a subscription tag (i.e. "monthly-advanced") and
// returns just the type of subscription (i.e. "advanced")
func GetSubscriptionType(tag string) (string, error) {
	subType := tag
	if strings.Contains(subType, "-") {
		subType = strings.Split(tag, "-")[1]
	}

	if subType != TypeNovice && subType != TypeRegular && subType != TypeAdvanced {
		return "", errors.New("invalid subscription type")
	}

	return subType, nil
}

// GetSubscriptionExpiration returns a point in time in the future that the
// user's selected subscription should expire.
func GetSubscriptionExpiration(tag string) (time.Time, error) {
	duration, err := GetSubscriptionDuration(tag)
	if err != nil {
		return time.Time{}, err
	}

	expFn, ok := membershipDurationFunctionMap[duration]
	if !ok {
		return time.Time{}, err
	}

	return expFn(), nil
}

// GetSubscriptionStorage returns the amount of storage a user should receive
// for their selected subscription.
func GetSubscriptionStorage(tag string) (int64, error) {
	subType, err := GetSubscriptionType(tag)
	if err != nil {
		return 0, err
	}

	storage, ok := StorageAmountMap[subType]
	if !ok {
		return 0, errors.New("invalid subscription tag")
	}

	return storage, nil
}

// GetSubscriptionSend returns the amount of sending space a user should receive
// for their selected subscription.
func GetSubscriptionSend(tag string) (int64, error) {
	subType, err := GetSubscriptionType(tag)
	if err != nil {
		return 0, err
	}

	send, ok := SendAmountMap[subType]
	if !ok {
		return 0, errors.New("invalid subscription tag")
	}

	return send, nil
}

// IsValidSubscriptionTag checks to see if a provided subscription tag is valid
func IsValidSubscriptionTag(tag string) bool {
	for _, subTag := range ValidSubscriptionTags {
		if tag == subTag {
			return true
		}
	}

	return false
}
