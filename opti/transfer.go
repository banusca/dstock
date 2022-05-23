package opti

// Transfer storse all info regarding a transfer
type Transfer struct {
	FromStore      string
	ToStore        string
	Products       []TransferedProduct
	TotaItemsCount int
	OsaImprouve    int
}

// TransferedProduct is a item transfered
type TransferedProduct struct {
	Product string
	Quant   int
}
