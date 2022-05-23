package stockparser

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type StockItem struct {
	Has   int
	Needs int
}

type StockMatrix map[string]map[string]StockItem

// ParseFromCSV reads the stock from a CSV file with header: Store, Product, Stock, Max
func ParseFromCSV(inputBytes []byte) (StockMatrix, error) {
	csvReader := csv.NewReader(bufio.NewReader(bytes.NewReader(inputBytes)))
	log.Debug("Enter ParseFromCSV")
	stock := make(StockMatrix)

	for {
		record, err := csvReader.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}

		store := strings.TrimSpace(record[0])
		product := strings.TrimSpace(record[1])
		value := strings.TrimSpace(record[2])
		max := strings.TrimSpace(record[3])

		if store == "Store" && product == "Product" {
			// this is the header: check stuff
			continue
		}

		if stock[store] == nil {
			stock[store] = make(map[string]StockItem)
		}
		istock, errParseStock := strconv.Atoi(value)
		if errParseStock != nil {
			log.Error(errParseStock)
			return nil, errParseStock
		}

		imax, errParseMax := strconv.Atoi(max)
		if errParseMax != nil {
			log.Error(errParseMax)
			return nil, errParseMax
		}

		stock[store][product] = StockItem{
			Has:   istock,
			Needs: imax,
		}

	}

	return stock, nil
}
