package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"newsletter/models"

	"github.com/gorilla/mux"
)

// SubscriptionController - abonelik HTTP işlemlerini yönetir
type SubscriptionController struct {
	SubscriptionModel *models.SubscriptionModel
}

// NewSubscriptionController - yeni controller instance'ı oluşturur
func NewSubscriptionController(subscriptionModel *models.SubscriptionModel) *SubscriptionController {
	return &SubscriptionController{SubscriptionModel: subscriptionModel}
}

type SubscribeRequest struct {
	Email      string   `json:"email"`
	Categories []string `json:"categories,omitempty"`
}

type UnsubscribeRequest struct {
	Email string `json:"email"`
}

type PauseRequest struct {
	Email string `json:"email"`
}

type ReactivateRequest struct {
	Email string `json:"email"`
}

type UpdateCategoriesRequest struct {
	Email      string   `json:"email"`
	Categories []string `json:"categories"`
}

// Subscribe - yeni abonelik oluşturur
func (c *SubscriptionController) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Email validation
	if req.Email == "" || !isValidEmail(req.Email) {
		http.Error(w, "Valid email is required", http.StatusBadRequest)
		return
	}

	// Check if already subscribed
	existing, err := c.SubscriptionModel.GetByEmail(req.Email)
	if err == nil && existing != nil {
		if existing.Status == "active" {
			http.Error(w, "Email already subscribed", http.StatusConflict)
			return
		} else if existing.Status == "paused" {
			// Reactivate paused subscription
			if err := c.SubscriptionModel.Reactivate(req.Email); err != nil {
				http.Error(w, "Failed to reactivate subscription", http.StatusInternalServerError)
				return
			}
			existing.Status = "active"
			existing.Categories = req.Categories
			c.SubscriptionModel.UpdateCategories(req.Email, req.Categories)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(existing)
			return
		}
	}

	// Create new subscription
	subscription := &models.Subscription{
		Email:      req.Email,
		Categories: req.Categories,
	}

	if err := c.SubscriptionModel.Create(subscription); err != nil {
		http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subscription)
}

// Unsubscribe - abonelikten çıkarır
func (c *SubscriptionController) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	var req UnsubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Check if subscription exists
	subscription, err := c.SubscriptionModel.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	if err := c.SubscriptionModel.Unsubscribe(req.Email); err != nil {
		http.Error(w, "Failed to unsubscribe", http.StatusInternalServerError)
		return
	}

	subscription.Status = "unsubscribed"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// Pause - aboneliği durdurur
func (c *SubscriptionController) Pause(w http.ResponseWriter, r *http.Request) {
	var req PauseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Check if subscription exists
	subscription, err := c.SubscriptionModel.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	if subscription.Status != "active" {
		http.Error(w, "Only active subscriptions can be paused", http.StatusBadRequest)
		return
	}

	if err := c.SubscriptionModel.Pause(req.Email); err != nil {
		http.Error(w, "Failed to pause subscription", http.StatusInternalServerError)
		return
	}

	subscription.Status = "paused"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// Reactivate - durdurulan aboneliği aktifleştirir
func (c *SubscriptionController) Reactivate(w http.ResponseWriter, r *http.Request) {
	var req ReactivateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Check if subscription exists
	subscription, err := c.SubscriptionModel.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	if subscription.Status != "paused" {
		http.Error(w, "Only paused subscriptions can be reactivated", http.StatusBadRequest)
		return
	}

	if err := c.SubscriptionModel.Reactivate(req.Email); err != nil {
		http.Error(w, "Failed to reactivate subscription", http.StatusInternalServerError)
		return
	}

	subscription.Status = "active"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// GetSubscription - email ile abonelik bilgisi getirir
func (c *SubscriptionController) GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	subscription, err := c.SubscriptionModel.GetByEmail(email)
	if err != nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// GetAllSubscriptions - tüm abonelikleri getirir
func (c *SubscriptionController) GetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := c.SubscriptionModel.GetAll()
	if err != nil {
		http.Error(w, "Failed to fetch subscriptions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscriptions)
}

// GetSubscriptionsByStatus - duruma göre abonelikleri getirir
func (c *SubscriptionController) GetSubscriptionsByStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	status := vars["status"]

	subscriptions, err := c.SubscriptionModel.GetByStatus(status)
	if err != nil {
		http.Error(w, "Failed to fetch subscriptions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscriptions)
}

// UpdateCategories - abonelik kategorilerini günceller
func (c *SubscriptionController) UpdateCategories(w http.ResponseWriter, r *http.Request) {
	var req UpdateCategoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Check if subscription exists
	subscription, err := c.SubscriptionModel.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	if err := c.SubscriptionModel.UpdateCategories(req.Email, req.Categories); err != nil {
		http.Error(w, "Failed to update categories", http.StatusInternalServerError)
		return
	}

	subscription.Categories = req.Categories
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// isValidEmail - basit email doğrulama yapar
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
