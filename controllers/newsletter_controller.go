package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"newsletter/models"
	"newsletter/services"

	"github.com/gorilla/mux"
)

// NewsletterController - newsletter HTTP işlemlerini yönetir
type NewsletterController struct {
	NewsletterModel   *models.NewsletterModel
	SubscriptionModel *models.SubscriptionModel
	EmailService      *services.EmailService
}

// NewNewsletterController - yeni newsletter controller oluşturur
func NewNewsletterController(newsletterModel *models.NewsletterModel, subscriptionModel *models.SubscriptionModel, emailService *services.EmailService) *NewsletterController {
	return &NewsletterController{
		NewsletterModel:   newsletterModel,
		SubscriptionModel: subscriptionModel,
		EmailService:      emailService,
	}
}

type CreateNewsletterRequest struct {
	Title    string `json:"title"`
	Subject  string `json:"subject"`
	Content  string `json:"content"`
	Category string `json:"category,omitempty"`
}

type SendNewsletterRequest struct {
	NewsletterID int      `json:"newsletter_id"`
	Emails       []string `json:"emails"`
}

type SendNewsletterToSubscribersRequest struct {
	NewsletterID int    `json:"newsletter_id"`
	Status       string `json:"status,omitempty"` // active, paused, all
	Category     string `json:"category,omitempty"`
}

// CreateNewsletter - yeni newsletter oluşturur
func (c *NewsletterController) CreateNewsletter(w http.ResponseWriter, r *http.Request) {
	var req CreateNewsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Title == "" || req.Subject == "" || req.Content == "" {
		http.Error(w, "Title, Subject and Content are required", http.StatusBadRequest)
		return
	}

	// Newsletter oluştur
	newsletter := &models.Newsletter{
		Title:    req.Title,
		Subject:  req.Subject,
		Content:  req.Content,
		Category: req.Category,
	}

	if err := c.NewsletterModel.Create(newsletter); err != nil {
		http.Error(w, "Failed to create newsletter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newsletter)
}

// GetNewsletter - ID ile newsletter getirir
func (c *NewsletterController) GetNewsletter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid newsletter ID", http.StatusBadRequest)
		return
	}

	newsletter, err := c.NewsletterModel.GetByID(id)
	if err != nil {
		http.Error(w, "Newsletter not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newsletter)
}

// GetAllNewsletters - tüm newsletter'ları getirir
func (c *NewsletterController) GetAllNewsletters(w http.ResponseWriter, r *http.Request) {
	newsletters, err := c.NewsletterModel.GetAll()
	if err != nil {
		http.Error(w, "Failed to fetch newsletters", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newsletters)
}

// GetNewslettersByStatus - duruma göre newsletter'ları getirir
func (c *NewsletterController) GetNewslettersByStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	status := vars["status"]

	newsletters, err := c.NewsletterModel.GetByStatus(status)
	if err != nil {
		http.Error(w, "Failed to fetch newsletters", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newsletters)
}

// UpdateNewsletter - newsletter günceller
func (c *NewsletterController) UpdateNewsletter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid newsletter ID", http.StatusBadRequest)
		return
	}

	var req CreateNewsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Newsletter'ı güncelle
	newsletter := &models.Newsletter{
		ID:       id,
		Title:    req.Title,
		Subject:  req.Subject,
		Content:  req.Content,
		Category: req.Category,
	}

	if err := c.NewsletterModel.Update(newsletter); err != nil {
		http.Error(w, "Failed to update newsletter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Güncellenmiş newsletter'ı getir
	updatedNewsletter, err := c.NewsletterModel.GetByID(id)
	if err != nil {
		http.Error(w, "Failed to fetch updated newsletter", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedNewsletter)
}

// SendNewsletterToEmails - seçili emaillere newsletter gönderir
func (c *NewsletterController) SendNewsletterToEmails(w http.ResponseWriter, r *http.Request) {
	var req SendNewsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validation
	if req.NewsletterID == 0 || len(req.Emails) == 0 {
		http.Error(w, "Newsletter ID and emails list are required", http.StatusBadRequest)
		return
	}

	// Newsletter'ı getir
	newsletter, err := c.NewsletterModel.GetByID(req.NewsletterID)
	if err != nil {
		http.Error(w, "Newsletter not found", http.StatusNotFound)
		return
	}

	// Email'leri hazırla
	var emailDataList []*services.EmailData
	for _, email := range req.Emails {
		emailDataList = append(emailDataList, &services.EmailData{
			To:      email,
			Subject: newsletter.Subject,
			Body:    newsletter.Content,
			IsHTML:  true,
		})
	}

	// Email'leri gönder
	if err := c.EmailService.SendBulkEmails(emailDataList); err != nil {
		http.Error(w, "Failed to send newsletter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Newsletter'ı gönderildi olarak işaretle
	if err := c.NewsletterModel.MarkAsSent(req.NewsletterID); err != nil {
		log.Printf("Warning: Failed to mark newsletter as sent: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "Newsletter sent successfully",
		"newsletter_id": req.NewsletterID,
		"emails_sent":  len(req.Emails),
		"emails":       req.Emails,
	})
}

// SendNewsletterToSubscribers - abonelere newsletter gönderir
func (c *NewsletterController) SendNewsletterToSubscribers(w http.ResponseWriter, r *http.Request) {
	var req SendNewsletterToSubscribersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validation
	if req.NewsletterID == 0 {
		http.Error(w, "Newsletter ID is required", http.StatusBadRequest)
		return
	}

	// Newsletter'ı getir
	newsletter, err := c.NewsletterModel.GetByID(req.NewsletterID)
	if err != nil {
		http.Error(w, "Newsletter not found", http.StatusNotFound)
		return
	}

	// Aboneleri getir
	var subscribers []*models.Subscription
	if req.Status != "" {
		subscribers, err = c.SubscriptionModel.GetByStatus(req.Status)
	} else {
		subscribers, err = c.SubscriptionModel.GetAll()
	}
	
	if err != nil {
		http.Error(w, "Failed to fetch subscribers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kategori filtresi uygula
	var filteredSubscribers []*models.Subscription
	if req.Category != "" {
		for _, subscriber := range subscribers {
			for _, category := range subscriber.Categories {
				if category == req.Category {
					filteredSubscribers = append(filteredSubscribers, subscriber)
					break
				}
			}
		}
	} else {
		filteredSubscribers = subscribers
	}

	// Email'leri hazırla
	var emailDataList []*services.EmailData
	for _, subscriber := range filteredSubscribers {
		emailDataList = append(emailDataList, &services.EmailData{
			To:      subscriber.Email,
			Subject: newsletter.Subject,
			Body:    newsletter.Content,
			IsHTML:  true,
		})
	}

	// Email'leri gönder
	if err := c.EmailService.SendBulkEmails(emailDataList); err != nil {
		http.Error(w, "Failed to send newsletter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Newsletter'ı gönderildi olarak işaretle
	if err := c.NewsletterModel.MarkAsSent(req.NewsletterID); err != nil {
		log.Printf("Warning: Failed to mark newsletter as sent: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"message":         "Newsletter sent to subscribers successfully",
		"newsletter_id":   req.NewsletterID,
		"subscribers_sent": len(filteredSubscribers),
		"status_filter":   req.Status,
		"category_filter": req.Category,
	})
}

// DeleteNewsletter - newsletter siler
func (c *NewsletterController) DeleteNewsletter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid newsletter ID", http.StatusBadRequest)
		return
	}

	if err := c.NewsletterModel.Delete(id); err != nil {
		http.Error(w, "Failed to delete newsletter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetNewsletterStats - newsletter istatistiklerini getirir
func (c *NewsletterController) GetNewsletterStats(w http.ResponseWriter, r *http.Request) {
	stats, err := c.NewsletterModel.GetStats()
	if err != nil {
		http.Error(w, "Failed to fetch newsletter stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
