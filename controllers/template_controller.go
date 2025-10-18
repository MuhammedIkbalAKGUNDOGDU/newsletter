package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// TemplateController - template HTTP işlemlerini yönetir
type TemplateController struct {
	// Template storage (şimdilik memory'de, sonra database'e taşınabilir)
}

// NewTemplateController - yeni template controller oluşturur
func NewTemplateController() *TemplateController {
	return &TemplateController{}
}

type EmailTemplate struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Category  string   `json:"category"`
	HTML      string   `json:"html"`
	Preview   string   `json:"preview"`
	Variables []string `json:"variables"`
}

type CreateTemplateRequest struct {
	Name      string   `json:"name"`
	Category  string   `json:"category"`
	HTML      string   `json:"html"`
	Variables []string `json:"variables"`
}

type RenderTemplateRequest struct {
	TemplateID string                 `json:"template_id"`
	Variables  map[string]interface{} `json:"variables"`
}

// Hazır template'ler
var DefaultTemplates = map[string]EmailTemplate{
	"tech_newsletter": {
		ID:   "tech_newsletter",
		Name: "Teknoloji Newsletter",
		Category: "technology",
		HTML: `<!DOCTYPE html>
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
		Variables: []string{"title", "content", "unsubscribe_link"},
	},
	"marketing_promo": {
		ID:   "marketing_promo",
		Name: "Pazarlama Promosyonu",
		Category: "marketing",
		HTML: `<!DOCTYPE html>
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
		Variables: []string{"title", "content", "cta_button", "cta_link", "unsubscribe_link"},
	},
	"simple_text": {
		ID:   "simple_text",
		Name: "Basit Metin",
		Category: "general",
		HTML: `<!DOCTYPE html>
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
		Variables: []string{"title", "content", "unsubscribe_link"},
	},
}

// GetAllTemplates - tüm template'leri getirir
func (c *TemplateController) GetAllTemplates(w http.ResponseWriter, r *http.Request) {
	var templates []EmailTemplate
	for _, template := range DefaultTemplates {
		templates = append(templates, template)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// GetTemplate - ID ile template getirir
func (c *TemplateController) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	template, exists := DefaultTemplates[templateID]
	if !exists {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// GetTemplatesByCategory - kategoriye göre template'leri getirir
func (c *TemplateController) GetTemplatesByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]

	var templates []EmailTemplate
	for _, template := range DefaultTemplates {
		if template.Category == category {
			templates = append(templates, template)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// RenderTemplate - template'i render eder
func (c *TemplateController) RenderTemplate(w http.ResponseWriter, r *http.Request) {
	var req RenderTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	template, exists := DefaultTemplates[req.TemplateID]
	if !exists {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	// Basit template rendering (Go template engine kullanılabilir)
	html := template.HTML
	for key, value := range req.Variables {
		placeholder := "{{." + key + "}}"
		html = strings.ReplaceAll(html, placeholder, fmt.Sprintf("%v", value))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"template_id": req.TemplateID,
		"html":        html,
		"variables":   req.Variables,
	})
}
