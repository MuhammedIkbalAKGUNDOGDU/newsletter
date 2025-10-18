package migrations

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) {
	log.Println("Running database migrations...")
	
	// Users table
	createUsersTable(db)
	
	// Newsletters table
	createNewslettersTable(db)
	
	// Subscriptions table
	createSubscriptionsTable(db)
	
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
			content TEXT NOT NULL,
			author_id INTEGER REFERENCES users(id),
			status VARCHAR(50) DEFAULT 'draft',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
			user_id INTEGER REFERENCES users(id),
			newsletter_id INTEGER REFERENCES newsletters(id),
			subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, newsletter_id)
		)`
	
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create subscriptions table:", err)
	}
	log.Println("Subscriptions table created/verified")
}

func createIndexes(db *sql.DB) {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_newsletters_author ON newsletters(author_id)",
		"CREATE INDEX IF NOT EXISTS idx_newsletters_status ON newsletters(status)",
		"CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_subscriptions_newsletter ON subscriptions(newsletter_id)",
	}
	
	for _, index := range indexes {
		_, err := db.Exec(index)
		if err != nil {
			log.Fatal("Failed to create index:", err)
		}
	}
	log.Println("Database indexes created/verified")
}
