package opti

type Settings struct {
	// a store can transfer up to MaxTo other stores
	//
	// required: false
	// min: 1
	MaxTo int `json:"maxTo"`

	// Minimal quantity of products to initiate a transfer
	//
	// required: false
	// min: 1
	MinProducts int `json:"MinProducts"`

	SecureStockIn []string `json:"secureStockIn"`
	DeStockIn     []string `json:"DeStockIn"`
}
