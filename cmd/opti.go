/*
Copyright © 2022 Bogdan ANUSCA <anusca.bogdan@gmail.com>

*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"gitlab.com/banusca/dstock/opti"
	"gitlab.com/banusca/dstock/stockparser"
)

// optiCmd represents the opti command
var optiCmd = &cobra.Command{
	Use:   "opti",
	Short: "Run a stock optimisation",
	Long: `Parameters:
	1. Input stock file -i
	2. Optimisation parameters -m -n --deStock --secureStock
	3. Output transfer file -o
	4. Optional params: -l`,
	Run: func(cmd *cobra.Command, args []string) {
		logLevel, _ := cmd.Flags().GetString("logLevel")
		level, _ := log.ParseLevel(logLevel)
		log.SetLevel(level)
		// read the input file
		stockPath, _ := cmd.Flags().GetString("input")
		dat, err := ioutil.ReadFile(stockPath)
		if err != nil {
			log.Error(err)
			return
		}
		stock, err := stockparser.ParseFromCSV(dat)
		if err != nil {
			log.Error(err)
			return
		}

		// run the optimisation
		maxTo, _ := cmd.Flags().GetInt("maxTo")
		minProducts, _ := cmd.Flags().GetInt("minProducts")

		resp := make(chan opti.Optimisation)
		progress := make(chan int)

		// secure|destock rule
		secureStock, _ := cmd.Flags().GetStringSlice("secureStock")
		deStock, _ := cmd.Flags().GetStringSlice("deStock")
		go func(stock stockparser.StockMatrix, settings opti.Settings, progress chan<- int) {
			resp <- opti.RunOpti(stock, settings, progress)
		}(stock, opti.Settings{MaxTo: maxTo, MinProducts: minProducts, SecureStockIn: secureStock, DeStockIn: deStock}, progress)

		go func() {
			for p := range progress {
				fmt.Printf("\rProgress: %d%%", p)
			}
		}()

		optim := <-resp
		fmt.Printf("\rProgress: 100%%\n")
		fmt.Printf("OSA increased: +%.2f%%\n", optim.OSAIncrease)

		// write output
		filePath, _ := cmd.Flags().GetString("output")
		if !strings.HasSuffix(filePath, ".csv") {
			filePath += ".csv"
		}

		f, err := os.Create(filePath)
		if err != nil {
			log.Error(err)
			return
		}
		defer f.Close()

		err = f.Truncate(0)
		if err != nil {
			log.Error(err)
			return
		}
		_, err = f.Seek(0, 0)
		if err != nil {
			log.Error(err)
			return
		}

		f.WriteString("From, To, Product, Quantity\n")

		for _, t := range optim.Transfers {
			for _, p := range t.Products {
				_, err := f.WriteString(fmt.Sprintf("%s,%s,%s,%d\n", t.FromStore, t.ToStore, p.Product, p.Quant))
				if err != nil {
					log.Error(err)
					return
				}
			}
		}
		f.Sync()
	},
}

func init() {
	rootCmd.AddCommand(optiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// optiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// optiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringP("input", "i", "", "input stock file path")
	rootCmd.PersistentFlags().StringP("output", "o", "opti-"+time.Now().Format("2006-01-02T15:04:05.999999999Z07:00")+".csv", "output transfer file path")
	rootCmd.PersistentFlags().IntP("minProducts", "m", 20, "minimum products count")
	rootCmd.PersistentFlags().IntP("maxTo", "n", 4, "max store count one store can send to")
	rootCmd.PersistentFlags().StringSlice("secureStock", []string{}, "list(comma separated) of stores to secure stock in")
	rootCmd.PersistentFlags().StringSlice("deStock", []string{}, "list(comma separated) of stores to reduce stock")
	rootCmd.PersistentFlags().StringP("logLevel", "l", "warn", "set log level: panic | fatal | error | warn | info | debug")
}
