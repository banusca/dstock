/*
Copyright © 2022 Bogdan ANUSCA <anusca.bogdan@gmail.com>

*/
package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"gitlab.com/banusca/dstock/cmd"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)
}

func main() {
	cmd.Execute()
}
