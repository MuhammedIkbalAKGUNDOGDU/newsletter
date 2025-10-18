package models

import (
	"database/sql"
	"time"
)

// Subscription model - abonelik bilgilerini tutar
type Subscription struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Status       string    `json:"status"` // active, paused, unsubscribed
	Categories   []string  `json:"categories,omitempty"`
	SubscribedAt time.Time `json:"subscribed_at"`
	PausedAt     *time.Time `json:"paused_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SubscriptionModel - abonelik veritabanı işlemlerini yönetir
type SubscriptionModel struct {
	DB *sql.DB
}

// NewSubscriptionModel - yeni subscription model instance'ı oluşturur
func NewSubscriptionModel(db *sql.DB) *SubscriptionModel {
	return &SubscriptionModel{DB: db}
}

// Create - yeni abonelik oluşturur
func (m *SubscriptionModel) Create(subscription *Subscription) error {
	query := `
		INSERT INTO subscriptions (email, status, categories, subscribed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	
	subscription.SubscribedAt = time.Now()
	subscription.UpdatedAt = time.Now()
	subscription.Status = "active"
	
	err := m.DB.QueryRow(query, subscription.Email, subscription.Status, 
		subscription.Categories, subscription.SubscribedAt, subscription.UpdatedAt).Scan(&subscription.ID)
	if err != nil {
		return err
	}
	
	return nil
}

// GetByEmail - email ile abonelik bilgisi getirir
func (m *SubscriptionModel) GetByEmail(email string) (*Subscription, error) {
	subscription := &Subscription{}
	query := `SELECT id, email, status, categories, subscribed_at, paused_at, updated_at FROM subscriptions WHERE email = $1`
	
	var categoriesStr sql.NullString
	var pausedAt sql.NullTime
	
	err := m.DB.QueryRow(query, email).Scan(
		&subscription.ID, &subscription.Email, &subscription.Status, 
		&categoriesStr, &subscription.SubscribedAt, &pausedAt, &subscription.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	// Categories string'ini array'e çevir
	if categoriesStr.Valid && categoriesStr.String != "" {
		subscription.Categories = []string{categoriesStr.String}
	}
	
	if pausedAt.Valid {
		subscription.PausedAt = &pausedAt.Time
	}
	
	return subscription, nil
}

// GetAll - tüm abonelikleri getirir
func (m *SubscriptionModel) GetAll() ([]*Subscription, error) {
	query := `SELECT id, email, status, categories, subscribed_at, paused_at, updated_at FROM subscriptions ORDER BY subscribed_at DESC`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var subscriptions []*Subscription
	for rows.Next() {
		subscription := &Subscription{}
		var categoriesStr sql.NullString
		var pausedAt sql.NullTime
		
		err := rows.Scan(
			&subscription.ID, &subscription.Email, &subscription.Status,
			&categoriesStr, &subscription.SubscribedAt, &pausedAt, &subscription.UpdatedAt)
		
		if err != nil {
			return nil, err
		}
		
		// Categories string'ini array'e çevir
		if categoriesStr.Valid && categoriesStr.String != "" {
			subscription.Categories = []string{categoriesStr.String}
		}
		
		if pausedAt.Valid {
			subscription.PausedAt = &pausedAt.Time
		}
		
		subscriptions = append(subscriptions, subscription)
	}
	
	return subscriptions, nil
}

// GetByStatus - duruma göre abonelikleri getirir
func (m *SubscriptionModel) GetByStatus(status string) ([]*Subscription, error) {
	query := `SELECT id, email, status, categories, subscribed_at, paused_at, updated_at FROM subscriptions WHERE status = $1 ORDER BY subscribed_at DESC`
	rows, err := m.DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var subscriptions []*Subscription
	for rows.Next() {
		subscription := &Subscription{}
		var categoriesStr sql.NullString
		var pausedAt sql.NullTime
		
		err := rows.Scan(
			&subscription.ID, &subscription.Email, &subscription.Status,
			&categoriesStr, &subscription.SubscribedAt, &pausedAt, &subscription.UpdatedAt)
		
		if err != nil {
			return nil, err
		}
		
		// Categories string'ini array'e çevir
		if categoriesStr.Valid && categoriesStr.String != "" {
			subscription.Categories = []string{categoriesStr.String}
		}
		
		if pausedAt.Valid {
			subscription.PausedAt = &pausedAt.Time
		}
		
		subscriptions = append(subscriptions, subscription)
	}
	
	return subscriptions, nil
}

// Pause - aboneliği durdurur
func (m *SubscriptionModel) Pause(email string) error {
	query := `
		UPDATE subscriptions 
		SET status = 'paused', paused_at = $1, updated_at = $2
		WHERE email = $3`
	
	now := time.Now()
	_, err := m.DB.Exec(query, now, now, email)
	return err
}

// Unsubscribe - abonelikten çıkarır
func (m *SubscriptionModel) Unsubscribe(email string) error {
	query := `
		UPDATE subscriptions 
		SET status = 'unsubscribed', updated_at = $1
		WHERE email = $2`
	
	_, err := m.DB.Exec(query, time.Now(), email)
	return err
}

// Reactivate - durdurulan aboneliği aktifleştirir
func (m *SubscriptionModel) Reactivate(email string) error {
	query := `
		UPDATE subscriptions 
		SET status = 'active', paused_at = NULL, updated_at = $1
		WHERE email = $2`
	
	_, err := m.DB.Exec(query, time.Now(), email)
	return err
}

// UpdateCategories - abonelik kategorilerini günceller
func (m *SubscriptionModel) UpdateCategories(email string, categories []string) error {
	query := `
		UPDATE subscriptions 
		SET categories = $1, updated_at = $2
		WHERE email = $3`
	
	_, err := m.DB.Exec(query, categories, time.Now(), email)
	return err
}
