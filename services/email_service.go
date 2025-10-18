package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

// EmailService - email gönderim servisi
type EmailService struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Secure   bool
}

// EmailData - gönderilecek email verisi
type EmailData struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}

// NewEmailService - yeni email servisi oluşturur
func NewEmailService() *EmailService {
	host := getEnv("EMAIL_HOST", "smtp.hostinger.com")
	portStr := getEnv("EMAIL_PORT", "465")
	port, _ := strconv.Atoi(portStr)
	username := getEnv("EMAIL_USER", "")
	password := getEnv("EMAIL_PASSWORD", "")
	from := getEnv("EMAIL_FROM", "")
	secureStr := getEnv("EMAIL_SECURE", "true")
	secure := secureStr == "true"

	return &EmailService{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		Secure:   secure,
	}
}

// SendEmail - tek email gönderir
func (es *EmailService) SendEmail(emailData *EmailData) error {
	// SMTP bağlantısı oluştur
	addr := fmt.Sprintf("%s:%d", es.Host, es.Port)
	
	var auth smtp.Auth
	if es.Username != "" && es.Password != "" {
		auth = smtp.PlainAuth("", es.Username, es.Password, es.Host)
	}

	// Email içeriğini hazırla
	message := es.buildMessage(emailData)

	// Email gönder
	if es.Secure {
		// TLS ile güvenli bağlantı
		return es.sendWithTLS(addr, auth, es.From, []string{emailData.To}, []byte(message))
	} else {
		// Normal SMTP bağlantısı
		return smtp.SendMail(addr, auth, es.From, []string{emailData.To}, []byte(message))
	}
}

// SendBulkEmails - toplu email gönderir
func (es *EmailService) SendBulkEmails(emails []*EmailData) error {
	var errors []error
	
	for i, emailData := range emails {
		err := es.SendEmail(emailData)
		if err != nil {
			log.Printf("Email %d gönderilemedi: %v", i+1, err)
			errors = append(errors, err)
		} else {
			log.Printf("Email %d başarıyla gönderildi: %s", i+1, emailData.To)
		}
		
		// Rate limiting - email'ler arası bekleme
		if i < len(emails)-1 {
			time.Sleep(1 * time.Second)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("%d email gönderilemedi", len(errors))
	}
	
	return nil
}

// SendWithRetry - retry mekanizması ile email gönderir
func (es *EmailService) SendWithRetry(emailData *EmailData) error {
	maxRetries := getEnvAsInt("EMAIL_RETRY_ATTEMPTS", 3)
	retryDelay := getEnvAsInt("EMAIL_RETRY_DELAY", 5)
	
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		err := es.SendEmail(emailData)
		if err == nil {
			return nil
		}
		
		lastErr = err
		log.Printf("Email gönderimi başarısız (deneme %d/%d): %v", i+1, maxRetries, err)
		
		if i < maxRetries-1 {
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}
	}
	
	return fmt.Errorf("email %d deneme sonrası gönderilemedi: %v", maxRetries, lastErr)
}

// buildMessage - email mesajını oluşturur
func (es *EmailService) buildMessage(emailData *EmailData) string {
	headers := make(map[string]string)
	headers["From"] = es.From
	headers["To"] = emailData.To
	headers["Subject"] = emailData.Subject
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	
	if emailData.IsHTML {
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}
	
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + emailData.Body
	
	return message
}

// sendWithTLS - TLS ile güvenli email gönderir
func (es *EmailService) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// TLS bağlantısı oluştur
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         es.Host,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	// SMTP client oluştur
	client, err := smtp.NewClient(conn, es.Host)
	if err != nil {
		return err
	}
	defer client.Quit()

	// Authentication
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	// From
	if err = client.Mail(from); err != nil {
		return err
	}

	// To
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write(msg)
	return err
}

// TestConnection - SMTP bağlantısını test eder
func (es *EmailService) TestConnection() error {
	addr := fmt.Sprintf("%s:%d", es.Host, es.Port)
	
	var auth smtp.Auth
	if es.Username != "" && es.Password != "" {
		auth = smtp.PlainAuth("", es.Username, es.Password, es.Host)
	}

	if es.Secure {
		conn, err := tls.Dial("tcp", addr, &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         es.Host,
		})
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, es.Host)
		if err != nil {
			return err
		}
		defer client.Quit()

		if auth != nil {
			return client.Auth(auth)
		}
		return nil
	} else {
		client, err := smtp.Dial(addr)
		if err != nil {
			return err
		}
		defer client.Quit()

		if auth != nil {
			return client.Auth(auth)
		}
		return nil
	}
}

// Environment variable'ı al, yoksa default değer döndür
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Environment variable'ı integer olarak al
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
