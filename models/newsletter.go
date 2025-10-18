package models

import (
	"database/sql"
	"time"
)

// Newsletter model - newsletter bilgilerini tutar
type Newsletter struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Subject   string    `json:"subject"`
	Content   string    `json:"content"`
	Status    string    `json:"status"` // draft, scheduled, sent
	Category  string    `json:"category,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
}

// NewsletterModel - newsletter veritabanı işlemlerini yönetir
type NewsletterModel struct {
	DB *sql.DB
}

// NewNewsletterModel - yeni newsletter model oluşturur
func NewNewsletterModel(db *sql.DB) *NewsletterModel {
	return &NewsletterModel{DB: db}
}

// Create - yeni newsletter oluşturur
func (m *NewsletterModel) Create(newsletter *Newsletter) error {
	query := `
		INSERT INTO newsletters (title, subject, content, status, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	
	now := time.Now()
	newsletter.CreatedAt = now
	newsletter.UpdatedAt = now
	newsletter.Status = "draft"
	
	err := m.DB.QueryRow(query, newsletter.Title, newsletter.Subject, 
		newsletter.Content, newsletter.Status, newsletter.Category, 
		newsletter.CreatedAt, newsletter.UpdatedAt).Scan(&newsletter.ID)
	if err != nil {
		return err
	}
	
	return nil
}

// GetByID - ID ile newsletter getirir
func (m *NewsletterModel) GetByID(id int) (*Newsletter, error) {
	newsletter := &Newsletter{}
	query := `SELECT id, title, subject, content, status, category, created_at, updated_at, sent_at FROM newsletters WHERE id = $1`
	
	var sentAt sql.NullTime
	
	err := m.DB.QueryRow(query, id).Scan(
		&newsletter.ID, &newsletter.Title, &newsletter.Subject, 
		&newsletter.Content, &newsletter.Status, &newsletter.Category,
		&newsletter.CreatedAt, &newsletter.UpdatedAt, &sentAt)
	
	if err != nil {
		return nil, err
	}
	
	if sentAt.Valid {
		newsletter.SentAt = &sentAt.Time
	}
	
	return newsletter, nil
}

// GetAll - tüm newsletter'ları getirir
func (m *NewsletterModel) GetAll() ([]*Newsletter, error) {
	query := `SELECT id, title, subject, content, status, category, created_at, updated_at, sent_at FROM newsletters ORDER BY created_at DESC`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var newsletters []*Newsletter
	for rows.Next() {
		newsletter := &Newsletter{}
		var sentAt sql.NullTime
		
		err := rows.Scan(
			&newsletter.ID, &newsletter.Title, &newsletter.Subject,
			&newsletter.Content, &newsletter.Status, &newsletter.Category,
			&newsletter.CreatedAt, &newsletter.UpdatedAt, &sentAt)
		
		if err != nil {
			return nil, err
		}
		
		if sentAt.Valid {
			newsletter.SentAt = &sentAt.Time
		}
		
		newsletters = append(newsletters, newsletter)
	}
	
	return newsletters, nil
}

// GetByStatus - duruma göre newsletter'ları getirir
func (m *NewsletterModel) GetByStatus(status string) ([]*Newsletter, error) {
	query := `SELECT id, title, subject, content, status, category, created_at, updated_at, sent_at FROM newsletters WHERE status = $1 ORDER BY created_at DESC`
	rows, err := m.DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var newsletters []*Newsletter
	for rows.Next() {
		newsletter := &Newsletter{}
		var sentAt sql.NullTime
		
		err := rows.Scan(
			&newsletter.ID, &newsletter.Title, &newsletter.Subject,
			&newsletter.Content, &newsletter.Status, &newsletter.Category,
			&newsletter.CreatedAt, &newsletter.UpdatedAt, &sentAt)
		
		if err != nil {
			return nil, err
		}
		
		if sentAt.Valid {
			newsletter.SentAt = &sentAt.Time
		}
		
		newsletters = append(newsletters, newsletter)
	}
	
	return newsletters, nil
}

// GetByCategory - kategoriye göre newsletter'ları getirir
func (m *NewsletterModel) GetByCategory(category string) ([]*Newsletter, error) {
	query := `SELECT id, title, subject, content, status, category, created_at, updated_at, sent_at FROM newsletters WHERE category = $1 ORDER BY created_at DESC`
	rows, err := m.DB.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var newsletters []*Newsletter
	for rows.Next() {
		newsletter := &Newsletter{}
		var sentAt sql.NullTime
		
		err := rows.Scan(
			&newsletter.ID, &newsletter.Title, &newsletter.Subject,
			&newsletter.Content, &newsletter.Status, &newsletter.Category,
			&newsletter.CreatedAt, &newsletter.UpdatedAt, &sentAt)
		
		if err != nil {
			return nil, err
		}
		
		if sentAt.Valid {
			newsletter.SentAt = &sentAt.Time
		}
		
		newsletters = append(newsletters, newsletter)
	}
	
	return newsletters, nil
}

// Update - newsletter günceller
func (m *NewsletterModel) Update(newsletter *Newsletter) error {
	query := `
		UPDATE newsletters 
		SET title = $1, subject = $2, content = $3, status = $4, category = $5, updated_at = $6
		WHERE id = $7`
	
	newsletter.UpdatedAt = time.Now()
	_, err := m.DB.Exec(query, newsletter.Title, newsletter.Subject, 
		newsletter.Content, newsletter.Status, newsletter.Category, 
		newsletter.UpdatedAt, newsletter.ID)
	return err
}

// MarkAsSent - newsletter'ı gönderildi olarak işaretler
func (m *NewsletterModel) MarkAsSent(id int) error {
	query := `
		UPDATE newsletters 
		SET status = 'sent', sent_at = $1, updated_at = $2
		WHERE id = $3`
	
	now := time.Now()
	_, err := m.DB.Exec(query, now, now, id)
	return err
}

// Delete - newsletter siler
func (m *NewsletterModel) Delete(id int) error {
	query := `DELETE FROM newsletters WHERE id = $1`
	_, err := m.DB.Exec(query, id)
	return err
}

// GetStats - newsletter istatistiklerini getirir
func (m *NewsletterModel) GetStats() (map[string]int, error) {
	stats := make(map[string]int)
	
	// Toplam newsletter sayısı
	var total int
	err := m.DB.QueryRow("SELECT COUNT(*) FROM newsletters").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total
	
	// Draft sayısı
	var draft int
	err = m.DB.QueryRow("SELECT COUNT(*) FROM newsletters WHERE status = 'draft'").Scan(&draft)
	if err != nil {
		return nil, err
	}
	stats["draft"] = draft
	
	// Sent sayısı
	var sent int
	err = m.DB.QueryRow("SELECT COUNT(*) FROM newsletters WHERE status = 'sent'").Scan(&sent)
	if err != nil {
		return nil, err
	}
	stats["sent"] = sent
	
	// Scheduled sayısı
	var scheduled int
	err = m.DB.QueryRow("SELECT COUNT(*) FROM newsletters WHERE status = 'scheduled'").Scan(&scheduled)
	if err != nil {
		return nil, err
	}
	stats["scheduled"] = scheduled
	
	return stats, nil
}
