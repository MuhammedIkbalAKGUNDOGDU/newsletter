package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"newsletter/services"

	"github.com/gorilla/mux"
)

// EmailController - email HTTP işlemlerini yönetir
type EmailController struct {
	EmailService *services.EmailService
}

// NewEmailController - yeni email controller oluşturur
func NewEmailController(emailService *services.EmailService) *EmailController {
	return &EmailController{EmailService: emailService}
}

type SendEmailRequest struct {
	To         string                 `json:"to"`
	Subject    string                 `json:"subject"`
	Body       string                 `json:"body"`
	IsHTML     bool                   `json:"is_html"`
	TemplateID string                 `json:"template_id,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
}

type SendBulkEmailRequest struct {
	Emails []SendEmailRequest `json:"emails"`
}

type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendEmail - tek email gönderir
func (c *EmailController) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validation
	if req.To == "" || req.Subject == "" {
		http.Error(w, "To and Subject are required", http.StatusBadRequest)
		return
	}

	// Body veya template_id gerekli
	if req.Body == "" && req.TemplateID == "" {
		http.Error(w, "Either Body or TemplateID is required", http.StatusBadRequest)
		return
	}

	// Template kullanılıyorsa render et
	body := req.Body
	if req.TemplateID != "" {
		renderedBody, err := c.renderTemplate(req.TemplateID, req.Variables)
		if err != nil {
			http.Error(w, "Failed to render template: "+err.Error(), http.StatusBadRequest)
			return
		}
		body = renderedBody
		req.IsHTML = true // Template'ler HTML'dir
	}

	// Email data oluştur
	emailData := &services.EmailData{
		To:      req.To,
		Subject: req.Subject,
		Body:    body,
		IsHTML:  req.IsHTML,
	}

	// Email gönder
	if err := c.EmailService.SendWithRetry(emailData); err != nil {
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"message":     "Email sent successfully",
		"to":          req.To,
		"template_id": req.TemplateID,
	})
}

// SendBulkEmails - toplu email gönderir
func (c *EmailController) SendBulkEmails(w http.ResponseWriter, r *http.Request) {
	var req SendBulkEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Emails) == 0 {
		http.Error(w, "At least one email is required", http.StatusBadRequest)
		return
	}

	// Email data'ları oluştur
	var emailDataList []*services.EmailData
	for _, email := range req.Emails {
		if email.To == "" || email.Subject == "" || email.Body == "" {
			http.Error(w, "All emails must have To, Subject and Body", http.StatusBadRequest)
			return
		}

		emailDataList = append(emailDataList, &services.EmailData{
			To:      email.To,
			Subject: email.Subject,
			Body:    email.Body,
			IsHTML:  email.IsHTML,
		})
	}

	// Toplu email gönder
	if err := c.EmailService.SendBulkEmails(emailDataList); err != nil {
		http.Error(w, "Failed to send bulk emails: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Bulk emails sent successfully",
		"count":   len(req.Emails),
	})
}

// TestConnection - SMTP bağlantısını test eder
func (c *EmailController) TestConnection(w http.ResponseWriter, r *http.Request) {
	err := c.EmailService.TestConnection()
	
	response := TestConnectionResponse{
		Success: err == nil,
	}
	
	if err != nil {
		response.Message = "Connection failed: " + err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		response.Message = "Connection successful"
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendTestEmail - test email gönderir
func (c *EmailController) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	// Test email data oluştur
	emailData := &services.EmailData{
		To:      email,
		Subject: "Newsletter Test Email",
		Body:    "<h1>Test Email</h1><p>This is a test email from Newsletter API.</p><p>If you received this email, your SMTP configuration is working correctly!</p>",
		IsHTML:  true,
	}

	// Test email gönder
	if err := c.EmailService.SendWithRetry(emailData); err != nil {
		http.Error(w, "Failed to send test email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Test email sent successfully",
		"to":      email,
	})
}

// GetEmailConfig - email konfigürasyonunu getirir (güvenlik için şifre hariç)
func (c *EmailController) GetEmailConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"host":     c.EmailService.Host,
		"port":     c.EmailService.Port,
		"secure":   c.EmailService.Secure,
		"from":     c.EmailService.From,
		"username": c.EmailService.Username,
		"password": "***hidden***", // Güvenlik için şifre gizlenir
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// renderTemplate - template'i render eder
func (c *EmailController) renderTemplate(templateID string, variables map[string]interface{}) (string, error) {
	// Template'leri import et (template_controller'dan)
	templates := map[string]string{
		"tech_newsletter": `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 0; background: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background: white; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; line-height: 1.6; }
        .footer { background: #333; color: white; padding: 15px; text-align: center; }
        .footer a { color: #ccc; text-decoration: none; }
        @media only screen and (max-width: 600px) {
            .container { width: 100% !important; }
            .header, .content, .footer { padding: 15px !important; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.title}}</h1>
        </div>
        <div class="content">
            {{.content}}
        </div>
        <div class="footer">
            <p>© 2024 Newsletter</p>
            <a href="{{.unsubscribe_link}}">Abonelikten Çık</a>
        </div>
    </div>
</body>
</html>`,
		"marketing_promo": `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 0; background: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background: white; }
        .header { background: #28a745; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; line-height: 1.6; }
        .cta-button { display: inline-block; background: #dc3545; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { background: #333; color: white; padding: 15px; text-align: center; }
        .footer a { color: #ccc; text-decoration: none; }
        @media only screen and (max-width: 600px) {
            .container { width: 100% !important; }
            .header, .content, .footer { padding: 15px !important; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.title}}</h1>
        </div>
        <div class="content">
            {{.content}}
            <br><br>
            <a href="{{.cta_link}}" class="cta-button">{{.cta_button}}</a>
        </div>
        <div class="footer">
            <p>© 2024 Newsletter</p>
            <a href="{{.unsubscribe_link}}">Abonelikten Çık</a>
        </div>
    </div>
</body>
</html>`,
		"simple_text": `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f9f9f9; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 20px; border-radius: 5px; }
        .footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; }
        .footer a { color: #999; }
    </style>
</head>
<body>
    <div class="container">
        <h2>{{.title}}</h2>
        <div>{{.content}}</div>
        <div class="footer">
            <a href="{{.unsubscribe_link}}">Abonelikten Çık</a>
        </div>
    </div>
</body>
</html>`,
	}

	template, exists := templates[templateID]
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateID)
	}

	// Variable'ları replace et
	html := template
	for key, value := range variables {
		placeholder := "{{." + key + "}}"
		html = strings.ReplaceAll(html, placeholder, fmt.Sprintf("%v", value))
	}

	return html, nil
}
