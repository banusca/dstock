package opti

import (
	"testing"

	"gitlab.com/banusca/dstock/stockparser"
)

func TestOptiSuccess(t *testing.T) {
	// init test

	stock := stockparser.StockMatrix{
		"Paris": {
			"Apples": {Has: 7, Needs: 9},
			"Corn":   {Has: 0, Needs: 4},
			"Sugar":  {Has: 4, Needs: 4},
			"Gems":   {Has: 2, Needs: 0},
			"Roses":  {Has: 6, Needs: 4},
		},
		"Bucharest": {
			"Apples": {Has: 17, Needs: 3},
			"Corn":   {Has: 6, Needs: 2},
			"Sugar":  {Has: 4, Needs: 4},
			"Gems":   {Has: 3, Needs: 0},
			"Roses":  {Has: 5, Needs: 2},
		},
	}
	settings := Settings{MaxTo: 1, MinProducts: 1}
	resp := make(chan Optimisation)
	progress := make(chan int)

	shouldGet := Optimisation{
		OSAIncrease: 10,
		Transfers: []Transfer{
			{
				FromStore: "Bucharest",
				ToStore:   "Paris",
				Products: []TransferedProduct{
					{Product: "Apples", Quant: 2},
					{Product: "Corn", Quant: 4},
				},
			},
		},
	}

	// run the optimisation
	go func(stock stockparser.StockMatrix, settings Settings, progress chan<- int) {
		resp <- RunOpti(stock, settings, progress)
	}(stock, settings, progress)

	go func() {
		for range progress {
			// we could test the progress here
		}
	}()

	// get the result
	got := <-resp

	// check resp
	if got.OSAIncrease != shouldGet.OSAIncrease {
		t.Errorf("invalid OSA increase, got %f, want %f", got.OSAIncrease, shouldGet.OSAIncrease)
	}

	if len(got.Transfers) != len(shouldGet.Transfers) {
		t.Errorf("invalid transfer, got size %d, want  size %d", len(got.Transfers), len(shouldGet.Transfers))
	}

}
