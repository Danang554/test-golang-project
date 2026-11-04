package main

import (
	"log"

	"test-golang-project/config"
	"test-golang-project/database/seeder"
	"test-golang-project/pkg/database"
)

func main() {
	log.Println("[INFO] Starting database seeder...")

	// 1. Load application configuration (.env)
	cfg := config.LoadConfig()

	// 2. Connect to MySQL database
	db, err := database.NewMySQLDatabase(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to MySQL database: %v\n", err)
	}

	// 3. Execute all registered seeders
	if err := seeder.RunAll(db); err != nil {
		log.Fatalf("[FATAL] Seeding failed with error: %v\n", err)
	}

	log.Println("[INFO] Database seeding process finished.")
}
