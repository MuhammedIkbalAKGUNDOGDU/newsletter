package migrations

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) {
	log.Println("Running database migrations...")
	
	// Subscriptions table
	createSubscriptionsTable(db)
	
	// Newsletters table
	createNewslettersTable(db)
	
	// Indexes
	createIndexes(db)
	
	log.Println("Database migrations completed successfully")
}

func createUsersTable(db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}
	log.Println("Users table created/verified")
}

func createNewslettersTable(db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS newsletters (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			subject VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			status VARCHAR(50) DEFAULT 'draft',
			category VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			sent_at TIMESTAMP NULL
		)`
	
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create newsletters table:", err)
	}
	log.Println("Newsletters table created/verified")
}

func createSubscriptionsTable(db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS subscriptions (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			status VARCHAR(50) DEFAULT 'active',
			categories TEXT[],
			subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			paused_at TIMESTAMP NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create subscriptions table:", err)
	}
	log.Println("Subscriptions table created/verified")
}

func createIndexes(db *sql.DB) {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_subscriptions_email ON subscriptions(email)",
		"CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status)",
		"CREATE INDEX IF NOT EXISTS idx_newsletters_status ON newsletters(status)",
		"CREATE INDEX IF NOT EXISTS idx_newsletters_category ON newsletters(category)",
		"CREATE INDEX IF NOT EXISTS idx_newsletters_created_at ON newsletters(created_at)",
	}
	
	for _, index := range indexes {
		_, err := db.Exec(index)
		if err != nil {
			log.Fatal("Failed to create index:", err)
		}
	}
	log.Println("Database indexes created/verified")
}
