package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Test database connection
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=password dbname=generalusermanagement sslmode=disable")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}

	fmt.Println("✅ Database connection successful!")

	// Test creating a simple table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS connection_test (id SERIAL PRIMARY KEY, message TEXT)`)
	if err != nil {
		log.Fatal("Error creating test table:", err)
	}

	fmt.Println("✅ Database table creation successful!")

	// Clean up test table
	_, err = db.Exec(`DROP TABLE IF EXISTS connection_test`)
	if err != nil {
		log.Printf("Warning: Could not clean up test table: %v", err)
	}

	fmt.Println("✅ All database tests passed!")
}