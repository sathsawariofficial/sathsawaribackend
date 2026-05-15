package configuration

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"sync"
)

var once sync.Once
var ConfigurationData Configuration

func init() {
	once.Do(func() {
		configFile := flag.String("config", "", "Path to the configuration file")
		flag.Parse()
		// Check if the configuration file is provided
		if *configFile == "" {
			log.Println("Usage: go run main.go -config /path/to/configuration.json")
			os.Exit(1)
		}

		err := LoadConfiguration(*configFile)
		if err != nil {
			log.Fatal("Error loading config file")
		}
	})
}

func LoadConfiguration(confFile string) error {
	log.Println("Trying to get configuration from ", confFile)

	// Read JSON file
	jsonData, err := os.ReadFile(confFile)
	if err != nil {
		log.Println("Failed to load conf file due to ", err.Error())
		return err
	}

	// Unmarshal JSON data into the struct
	err = json.Unmarshal(jsonData, &ConfigurationData)
	if err != nil {
		log.Println("Failed to unmarshal conf file due to ", err.Error())
		return err
	}

	if ConfigurationData.Timeout == 0 {
		ConfigurationData.Timeout = DEFAULT_TIMEOUT
	}

	return nil
}
