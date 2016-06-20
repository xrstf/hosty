package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kardianos/osext"
)

var config *configuration
var store *storage
var hostyPath string

func main() {
	fmt.Print("Hosty2\n")
	fmt.Print("======\n")
	fmt.Print("\n")

	var err error

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if os.Args[1] == "hash" {
		cmdHash()
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	configFile := os.Args[2]

	// where am I?
	hostyPath, err = osext.ExecutableFolder()
	if err != nil {
		log.Fatal("Could not determine my own binary location: " + err.Error())
	}

	// load configuration
	config, err = loadConfiguration(configFile, hostyPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	// connect to database, setup storage
	db := connectToDatabase(config.DatabaseFile())
	store = NewStorage(db, config.Directories.Storage)

	switch os.Args[1] {
	case "serve":
		cmdServe(db)

	case "import":
		cmdImport(db)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	log.Println("Usage:")
	log.Println("  hosty hash                           interactively hash a password")
	log.Println("  hosty import <CONFIGFILE> <DIR>      import data from a Hosty1 installation")
	log.Println("  hosty serve <CONFIGFILE>             run the Hosty2 web application")
}
