package subscriptions

import (
	"fmt"
	"time"
	"yeetfile/shared"
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

const NonMemberPassMax = 100

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

// DescriptionMap maps subscription tags to a description of what the subscription
// actually provides
var DescriptionMap = map[string]string{
	MonthlyNovice: "1 Month YeetFile Novice Membership " +
		fmt.Sprintf("(%dGB vault storage + %dGB/month send)",
			NoviceStorage, NoviceSend),
	MonthlyRegular: "1 Month YeetFile Regular Membership " +
		fmt.Sprintf("(%dGB vault storage + %dGB/month send)",
			RegularStorage, RegularSend),
	MonthlyAdvanced: "1 Month YeetFile Advanced Membership " +
		fmt.Sprintf("(%dGB vault storage + %dGB/month send)",
			AdvancedStorage, AdvancedSend),
	YearlyNovice: "1 Year YeetFile Novice Membership " +
		fmt.Sprintf("(%dGB vault storage + %dGB/month send)",
			NoviceStorage, NoviceSend),
	YearlyRegular: "1 Year YeetFile Regular Membership " +
		fmt.Sprintf("(%dGB vault storage + %dGB/month send)",
			RegularStorage, RegularSend),
	YearlyAdvanced: "1 Year YeetFile Advanced Membership " +
		fmt.Sprintf("(%dGB vault storage + %dGB/month send)",
			AdvancedStorage, AdvancedSend),
}

// membershipDurationFunctionMap maps membership order types to a function for
// generating the correct end date for the user's membership
var membershipDurationFunctionMap = map[string]func() time.Time{
	monthly: func() time.Time { return shared.AddDate(0, 1) },
	yearly:  func() time.Time { return shared.AddDate(1, 0) },
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
