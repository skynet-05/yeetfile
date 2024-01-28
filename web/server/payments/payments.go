package payments

var TypeSub1Month = "sub1m"
var TypeSub1Year = "sub1y"
var Type100GB = "100gb"
var Type500GB = "500gb"
var Type1TB = "1tb"

// FIXME: Placeholder pricing -- needs updating for prod
var priceMapping = map[string]float32{
	TypeSub1Month: 0.01,
	TypeSub1Year:  0.01,
	Type100GB:     0.01,
	Type500GB:     0.01,
	Type1TB:       0.01,
}
