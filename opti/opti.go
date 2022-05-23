package opti

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.com/banusca/dstock/stockparser"
)

// Optimisation is the result of a opti run
type Optimisation struct {
	InStock     stockparser.StockMatrix
	AfterStock  stockparser.StockMatrix
	Transfers   []Transfer
	Settings    Settings
	OSAIncrease float64
}

// RunOpti returns the best optimisation
func RunOpti(stock stockparser.StockMatrix, settings Settings, progress chan<- int) Optimisation {
	log.Debug("Starting optimisation")
	mystock := make(stockparser.StockMatrix)
	prioriFirst := false
	if len(settings.SecureStockIn) > 0 || len(settings.DeStockIn) > 0 {
		log.Info("Start in priori mode")
		prioriFirst = true
	}
	for k, v := range stock {
		mystock[k] = make(map[string]stockparser.StockItem)
		for kk, vv := range v {
			mystock[k][kk] = stockparser.StockItem{
				Has:   vv.Has,
				Needs: vv.Needs,
			}
		}
	}

	var transfers []Transfer

	// rule 1
	maxTo := 10
	if settings.MaxTo > 0 {
		maxTo = settings.MaxTo
	}

	// rule 3
	minProducts := 1
	if settings.MinProducts > 0 {
		minProducts = settings.MinProducts
	}

	// matrix to store if a store reached his transfer limit
	transferCounterMat := make(map[string]int)

	// matrix to store if a store reached his transfer limit
	transferAvailabilityMat := make(map[string]map[string]bool)
	for store1 := range mystock {
		transferAvailabilityMat[store1] = make(map[string]bool)
		for store2 := range mystock {
			if store1 == store2 {
				transferAvailabilityMat[store1][store2] = false
			} else {
				transferAvailabilityMat[store1][store2] = true
			}
		}
	}

	loopCounter := 0
	for loopCounter < len(mystock)*maxTo {
		// matrix to store the theoretical max exchange betwin stores
		exchangeCounterMat := make(map[string]map[string]int)
		for store1 := range mystock {
			exchangeCounterMat[store1] = make(map[string]int)
			for store2 := range mystock {
				exchangeCounterMat[store1][store2] = -1
			}
		}
		// multi thread
		var wg sync.WaitGroup
		transfersList := make(chan Transfer)

		// compute the exchange count betwin stores
		for store1 := range mystock {
			for store2 := range mystock {
				if transferAvailabilityMat[store1][store2] && transferCounterMat[store1] < maxTo {
					if (prioriFirst && (stringInSlice(store1, settings.DeStockIn) || (stringInSlice(store2, settings.SecureStockIn)))) || !prioriFirst {
						wg.Add(1)
						go func(store1 string, store2 string) {
							//log.Debugf("Store1 %s to Store2 %s", store1, store2)
							transfersList <- doExchangeItems(mystock, settings, store1, store2, false)
							defer wg.Done()
						}(store1, store2)
					}
				}
			}
		}

		optimalTransferFound := Transfer{}
		go func() {
			for transfer := range transfersList {
				if transfer.TotaItemsCount >= minProducts {
					optimalTransferFound = determineBestTransfer(optimalTransferFound, transfer, []float64{0.99, 0.01})
				}
			}
		}()

		// fix for too fast optimisation
		time.Sleep(2 * time.Microsecond)

		// wait for all rutines to finish
		wg.Wait()

		// compute the max

		if optimalTransferFound.TotaItemsCount == 0 {
			if prioriFirst {
				// we are done with prioritary stores, resume normal flow
				prioriFirst = false
				log.Info("END priori mode")
				continue
			} else {
				// no more good options in normal flow, end it
				break
			}
		}

		transferAvailabilityMat[optimalTransferFound.FromStore][optimalTransferFound.ToStore] = false
		transferCounterMat[optimalTransferFound.FromStore]++

		// recompute stock
		transfers = append(transfers, doExchangeItems(mystock, settings, optimalTransferFound.FromStore, optimalTransferFound.ToStore, true))

		log.Debugf("Transfer from %s to %s, total items count %d, osa diff +%d", optimalTransferFound.FromStore, optimalTransferFound.ToStore, optimalTransferFound.TotaItemsCount, optimalTransferFound.OsaImprouve)
		loopCounter++

		//send progress
		progress <- loopCounter * 100 / (len(mystock) * maxTo)
	}

	log.Debug("End optimisation")

	return Optimisation{
		InStock:     stock,
		AfterStock:  mystock,
		Transfers:   transfers,
		Settings:    settings,
		OSAIncrease: computeOSA(stock, mystock),
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func determineBestTransfer(old Transfer, new Transfer, order []float64) Transfer {
	oldScore := float64(old.OsaImprouve)*order[0] + float64(old.TotaItemsCount)*order[1]
	newScore := float64(new.OsaImprouve)*order[0] + float64(new.TotaItemsCount)*order[1]
	if newScore > oldScore {
		return new
	} else if newScore == oldScore {
		// if same score order by store name just to have the same opti results all the time
		if old.FromStore+"-"+old.ToStore > new.FromStore+"-"+new.ToStore {
			return new
		}
	}
	return old
}

func computeOSA(inStock stockparser.StockMatrix, outStock stockparser.StockMatrix) float64 {
	osa := 0.0
	inItemsInStock := 0
	inTotalItems := 0
	for _, s := range inStock {
		for _, p := range s {
			if p.Has > 0 {
				inItemsInStock++
			}
			inTotalItems++
		}
	}

	outItemsInStock := 0
	outTotalItems := 0
	for _, s := range outStock {
		for _, p := range s {
			if p.Has > 0 {
				outItemsInStock++
			}
			outTotalItems++
		}
	}

	if outTotalItems != 0 {
		osa = (float64(outItemsInStock) - float64(inItemsInStock)) * 100.0 / float64(outTotalItems)
	}

	return osa
}

func doExchangeItems(stock stockparser.StockMatrix, settings Settings, store1 string, store2 string, alterStock bool) Transfer {
	totalCount := 0
	osaImprouve := 0
	var exchangeList []TransferedProduct
	for product, stockInStore1 := range stock[store1] {
		stockInStore2 := stock[store2][product]
		transferItems := 0
		if stockInStore2.Has-stockInStore2.Needs < 0 && stockInStore1.Has-stockInStore1.Needs > 0 {
			transferItems = min(stockInStore1.Has-stockInStore1.Needs, stockInStore2.Needs-stockInStore2.Has)
			//log.Debugf("I could transfer %d", transferItems)
			if stockInStore2.Has == 0 && transferItems > 0 {
				osaImprouve++
			}
			if alterStock {
				stockInStore2.Has += transferItems
				stock[store2][product] = stockInStore2

				stockInStore1.Has -= transferItems
				stock[store1][product] = stockInStore1
				exchangeList = append(exchangeList, TransferedProduct{Product: product, Quant: transferItems})
			}
			totalCount += transferItems
		}
	}
	return Transfer{FromStore: store1, ToStore: store2, Products: exchangeList, TotaItemsCount: totalCount, OsaImprouve: osaImprouve}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
