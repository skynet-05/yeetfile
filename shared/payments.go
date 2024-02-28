package shared

import "time"

var TypeSub3Months = "sub3m"
var TypeSub1Year = "sub1y"
var Type100GB = "100gb"
var Type500GB = "500gb"
var Type1TB = "1tb"

// PriceMapping FIXME: Placeholder pricing -- needs updating for prod
var PriceMapping = map[string]float32{
	TypeSub3Months: 0.50,
	TypeSub1Year:   0.10,
	Type100GB:      0.10,
	Type500GB:      0.10,
	Type1TB:        0.10,
}

// DescriptionMap maps product type tags to a description of what the product
// actually is
var DescriptionMap = map[string]string{
	TypeSub3Months: "3 Month YeetFile Membership",
	TypeSub1Year:   "1 Year YeetFile Membership",
	Type100GB:      "YeetFile 100GB Transfer Upgrade",
	Type500GB:      "YeetFile 500GB Transfer Upgrade",
	Type1TB:        "YeetFile 1TB Transfer Upgrade",
}

// MembershipDurationFunctionMap maps membership order types to a function for
// generating the correct end date for the user's membership
var MembershipDurationFunctionMap = map[string]func() time.Time{
	TypeSub3Months: func() time.Time { return AddDate(0, 3) },
	TypeSub1Year:   func() time.Time { return AddDate(1, 0) },
}

// UpgradeMap maps upgrade order types to an amount of storage that should be
// added to a user's account
var UpgradeMap = map[string]int{
	Type100GB: 107_374_182_400,   // 100GB
	Type500GB: 536_870_912_000,   // 500GB
	Type1TB:   1_073_741_824_000, // 1TB
}
