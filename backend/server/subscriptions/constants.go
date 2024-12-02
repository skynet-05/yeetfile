package subscriptions

import (
	"fmt"
)

const monthly = "monthly"
const yearly = "yearly"

const TypeNovice = "novice"
const TypeRegular = "regular"
const TypeAdvanced = "advanced"

// GB
const NoviceStorage = 50
const RegularStorage = 250
const AdvancedStorage = 500

// GB
const NoviceSend = 10
const RegularSend = 25
const AdvancedSend = 50

var MonthlyNovice = fmt.Sprintf("%s-%s", monthly, TypeNovice)
var MonthlyRegular = fmt.Sprintf("%s-%s", monthly, TypeRegular)
var MonthlyAdvanced = fmt.Sprintf("%s-%s", monthly, TypeAdvanced)
var YearlyNovice = fmt.Sprintf("%s-%s", yearly, TypeNovice)
var YearlyRegular = fmt.Sprintf("%s-%s", yearly, TypeRegular)
var YearlyAdvanced = fmt.Sprintf("%s-%s", yearly, TypeAdvanced)

type SubscriptionTemplateValues struct {
	MonthlyNovice   string
	YearlyNovice    string
	MonthlyRegular  string
	YearlyRegular   string
	MonthlyAdvanced string
	YearlyAdvanced  string

	NoviceStorage   int
	NoviceSend      int
	RegularStorage  int
	RegularSend     int
	AdvancedStorage int
	AdvancedSend    int

	MonthlyNovicePrice   int
	YearlyNovicePrice    int
	MonthlyRegularPrice  int
	YearlyRegularPrice   int
	MonthlyAdvancedPrice int
	YearlyAdvancedPrice  int
}

var TemplateValues = SubscriptionTemplateValues{
	MonthlyNovice:        MonthlyNovice,
	YearlyNovice:         YearlyNovice,
	MonthlyRegular:       MonthlyRegular,
	YearlyRegular:        YearlyRegular,
	MonthlyAdvanced:      MonthlyAdvanced,
	YearlyAdvanced:       YearlyAdvanced,
	NoviceStorage:        NoviceStorage,
	NoviceSend:           NoviceSend,
	RegularStorage:       RegularStorage,
	RegularSend:          RegularSend,
	AdvancedStorage:      AdvancedStorage,
	AdvancedSend:         AdvancedSend,
	MonthlyNovicePrice:   PriceMapping[MonthlyNovice],
	YearlyNovicePrice:    PriceMapping[YearlyNovice],
	MonthlyRegularPrice:  PriceMapping[MonthlyRegular],
	YearlyRegularPrice:   PriceMapping[YearlyRegular],
	MonthlyAdvancedPrice: PriceMapping[MonthlyAdvanced],
	YearlyAdvancedPrice:  PriceMapping[YearlyAdvanced],
}

// ValidSubscriptionTags defines the valid subscription tags that can be used
// when signing up for a subscription
var ValidSubscriptionTags = []string{
	MonthlyNovice,
	MonthlyRegular,
	MonthlyAdvanced,
	YearlyNovice,
	YearlyRegular,
	YearlyAdvanced,
}

// PriceMapping maps full subscription strings to their price (in USD)
var PriceMapping = map[string]int{
	MonthlyNovice:   3,
	MonthlyRegular:  6,
	MonthlyAdvanced: 9,
	YearlyNovice:    30,
	YearlyRegular:   60,
	YearlyAdvanced:  90,
}

var NameMap = map[string]string{
	TypeNovice:   "Novice Yeeter",
	TypeRegular:  "Regular Yeeter",
	TypeAdvanced: "Advanced Yeeter",
}

// StorageAmountMap maps subscription types to an amount of vault storage that
// should be provided to the user
var StorageAmountMap = map[string]int64{
	TypeNovice:   NoviceStorage * 1000 * 1000 * 1000,
	TypeRegular:  RegularStorage * 1000 * 1000 * 1000,
	TypeAdvanced: AdvancedStorage * 1000 * 1000 * 1000,
}

// SendAmountMap maps subscription types to an amount of file sending that
// should be provided to the user
var SendAmountMap = map[string]int64{
	TypeNovice:   NoviceSend * 1000 * 1000 * 1000,
	TypeRegular:  RegularSend * 1000 * 1000 * 1000,
	TypeAdvanced: AdvancedSend * 1000 * 1000 * 1000,
}

func GetSubTagName(subType string, isYearly bool) string {
	if isYearly {
		return fmt.Sprintf("%s-%s", yearly, subType)
	}

	return fmt.Sprintf("%s-%s", monthly, subType)
}
