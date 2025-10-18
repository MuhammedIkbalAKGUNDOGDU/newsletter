package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"newsletter/controllers"
	"newsletter/migrations"
	"newsletter/models"
	"newsletter/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type App struct {
	DB     *sql.DB
	Router *mux.Router
}

func (a *App) Initialize() {
	// .env dosyasını yükle
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database bağlantısını başlat
	a.initializeDB()

	// Migration'ları çalıştır
	migrations.RunMigrations(a.DB)

	// Router'ı başlat
	a.initializeRoutes()
}

func (a *App) initializeDB() {
	// .env dosyasından database bilgilerini al
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "newsletter")

	// PostgreSQL connection string oluştur
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Database'e bağlan
	var err error
	a.DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Bağlantıyı test et
	err = a.DB.Ping()
	if err != nil {
		log.Fatal("Database ping failed:", err)
	}

	log.Println("Database connection established successfully")
}

func (a *App) initializeRoutes() {
	a.Router = mux.NewRouter()

	// Models oluştur
	subscriptionModel := models.NewSubscriptionModel(a.DB)
	newsletterModel := models.NewNewsletterModel(a.DB)

	// Services oluştur
	emailService := services.NewEmailService()

	// Controllers oluştur
	subscriptionController := controllers.NewSubscriptionController(subscriptionModel)
	emailController := controllers.NewEmailController(emailService)
	newsletterController := controllers.NewNewsletterController(newsletterModel, subscriptionModel, emailService)
	templateController := controllers.NewTemplateController()

	// API routes
	a.Router.HandleFunc("/", a.homeHandler).Methods("GET")
	a.Router.HandleFunc("/health", a.healthHandler).Methods("GET")
	a.Router.HandleFunc("/api/status", a.statusHandler).Methods("GET")

	// Subscription routes
	a.Router.HandleFunc("/api/subscribe", subscriptionController.Subscribe).Methods("POST")
	a.Router.HandleFunc("/api/unsubscribe", subscriptionController.Unsubscribe).Methods("POST")
	a.Router.HandleFunc("/api/pause", subscriptionController.Pause).Methods("POST")
	a.Router.HandleFunc("/api/reactivate", subscriptionController.Reactivate).Methods("POST")
	a.Router.HandleFunc("/api/subscriptions/{email}", subscriptionController.GetSubscription).Methods("GET")
	a.Router.HandleFunc("/api/subscriptions", subscriptionController.GetAllSubscriptions).Methods("GET")
	a.Router.HandleFunc("/api/subscriptions/status/{status}", subscriptionController.GetSubscriptionsByStatus).Methods("GET")
	a.Router.HandleFunc("/api/subscriptions/categories", subscriptionController.UpdateCategories).Methods("PUT")

	// Email routes
	a.Router.HandleFunc("/api/email/send", emailController.SendEmail).Methods("POST")
	a.Router.HandleFunc("/api/email/send-bulk", emailController.SendBulkEmails).Methods("POST")
	a.Router.HandleFunc("/api/email/test-connection", emailController.TestConnection).Methods("GET")
	a.Router.HandleFunc("/api/email/test/{email}", emailController.SendTestEmail).Methods("POST")
	a.Router.HandleFunc("/api/email/config", emailController.GetEmailConfig).Methods("GET")

	// Newsletter routes
	a.Router.HandleFunc("/api/newsletters", newsletterController.CreateNewsletter).Methods("POST")
	a.Router.HandleFunc("/api/newsletters", newsletterController.GetAllNewsletters).Methods("GET")
	a.Router.HandleFunc("/api/newsletters/{id}", newsletterController.GetNewsletter).Methods("GET")
	a.Router.HandleFunc("/api/newsletters/{id}", newsletterController.UpdateNewsletter).Methods("PUT")
	a.Router.HandleFunc("/api/newsletters/{id}", newsletterController.DeleteNewsletter).Methods("DELETE")
	a.Router.HandleFunc("/api/newsletters/status/{status}", newsletterController.GetNewslettersByStatus).Methods("GET")
	a.Router.HandleFunc("/api/newsletters/send-to-emails", newsletterController.SendNewsletterToEmails).Methods("POST")
	a.Router.HandleFunc("/api/newsletters/send-to-subscribers", newsletterController.SendNewsletterToSubscribers).Methods("POST")
	a.Router.HandleFunc("/api/newsletters/stats", newsletterController.GetNewsletterStats).Methods("GET")

	// Template routes
	a.Router.HandleFunc("/api/templates", templateController.GetAllTemplates).Methods("GET")
	a.Router.HandleFunc("/api/templates/{id}", templateController.GetTemplate).Methods("GET")
	a.Router.HandleFunc("/api/templates/category/{category}", templateController.GetTemplatesByCategory).Methods("GET")
	a.Router.HandleFunc("/api/templates/render", templateController.RenderTemplate).Methods("POST")
}

func (a *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Newsletter API is running!", "status": "success"}`)
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	// Database bağlantısını kontrol et
	err := a.DB.Ping()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status": "unhealthy", "database": "disconnected"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "database": "connected"}`)
}

func (a *App) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{
		"service": "Newsletter API",
		"version": "1.0.0",
		"status": "running",
		"database": "connected"
	}`)
}

func (a *App) Run() {
	// Port'u .env dosyasından al, yoksa 8080 kullan
	port := getEnv("PORT", "8080")
	
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, a.Router))
}

// Environment variable'ı al, yoksa default değer döndür
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	app := App{}
	app.Initialize()
	app.Run()
}
