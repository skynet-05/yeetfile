package shared

import "time"

var TypeSub1Month = "sub1m"
var TypeSub1Year = "sub1y"
var Type100GB = "100gb"
var Type500GB = "500gb"
var Type1TB = "1tb"

// PriceMapping FIXME: Placeholder pricing -- needs updating for prod
var PriceMapping = map[string]float32{
	TypeSub1Month: 0.50,
	TypeSub1Year:  0.10,
	Type100GB:     0.10,
	Type500GB:     0.10,
	Type1TB:       0.10,
}

// DescriptionMap maps product type tags to a description of what the product
// actually is
var DescriptionMap = map[string]string{
	TypeSub1Month: "1 Month YeetFile Membership",
	TypeSub1Year:  "1 Year YeetFile Membership",
	Type100GB:     "YeetFile 100GB Transfer Upgrade",
	Type500GB:     "YeetFile 500GB Transfer Upgrade",
	Type1TB:       "YeetFile 1TB Transfer Upgrade",
}

// MembershipMap maps membership order types to a function for generating the
// correct end date for the user's membership
var MembershipMap = map[string]func() time.Time{
	TypeSub1Month: func() time.Time { return time.Now().AddDate(0, 1, 0) },
	TypeSub1Year:  func() time.Time { return time.Now().AddDate(1, 0, 0) },
}

// UpgradeMap maps upgrade order types to an amount of storage that should be
// added to a user's account
var UpgradeMap = map[string]int{
	Type100GB: 107_374_182_400,   // 100GB
	Type500GB: 536_870_912_000,   // 500GB
	Type1TB:   1_073_741_824_000, // 1TB
}
