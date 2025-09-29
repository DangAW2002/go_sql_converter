package main

import (
	"log"
	"os"
	"time"

	"go_sql_converter/internal"

	"gopkg.in/yaml.v3"
)

func main() {
	log.Println("Starting MQTT Subscriber...")

	// Read config
	log.Println("Loading configuration from config/config.yaml...")
	var config internal.Config
	for {
		configFile, err := os.Open("config/config.yaml")
		if err != nil {
			log.Printf("Error opening config file: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer configFile.Close()

		decoder := yaml.NewDecoder(configFile)
		err = decoder.Decode(&config)
		if err != nil {
			log.Printf("Error decoding config: %v. Retrying in 5 seconds...", err)
			configFile.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println("Config loaded successfully.")
		break
	}

	// Setup database
	db := internal.SetupDatabase(config)
	defer db.Close()

	// Setup MQTT and subscribe
	internal.SetupMQTT(config, db)

	// Keep running
	for {
		time.Sleep(1 * time.Second)
	}
}
