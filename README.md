# Newsletter API

Go ile yazılmış newsletter ve email gönderim API'si. PostgreSQL veritabanı kullanır ve .env dosyası ile yapılandırılabilir.

## 🚀 Kurulum

### 1. Bağımlılıkları yükle

```bash
go mod tidy
```

### 2. PostgreSQL veritabanını hazırla

```bash
# PostgreSQL'de veritabanı oluştur
createdb newsletter
```

### 3. Environment dosyasını yapılandır

`example.env` dosyasını kopyalayın ve `.env` olarak adlandırın:

```bash
cp example.env .env
```

`.env` dosyasındaki değerleri kendi ortamınıza göre düzenleyin:

```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=newsletter

EMAIL_HOST=smtp.hostinger.com
EMAIL_PORT=465
EMAIL_SECURE=true
EMAIL_USER=your_email@domain.com
EMAIL_PASSWORD=your_password
EMAIL_FROM=your_email@domain.com
```

### 4. Uygulamayı çalıştır

```bash
go run main.go
```

## 📚 API Endpoints

### 🏠 Ana Sayfa

- **GET** `/` - API'nin çalıştığını doğrular
- **GET** `/health` - Uygulama ve veritabanı durumunu kontrol eder
- **GET** `/api/status` - Detaylı sistem durumu

### 📧 Email API'leri

#### Email Gönderimi

- **POST** `/api/email/send` - Tek email gönderir

```json
{
  "to": "user@example.com",
  "subject": "Test Email",
  "body": "Hello World!",
  "is_html": false,
  "template_id": "tech_newsletter",
  "variables": {
    "title": "Test Title",
    "content": "<p>Test content</p>",
    "unsubscribe_link": "https://example.com/unsub"
  }
}
```

- **POST** `/api/email/send-bulk` - Toplu email gönderir

```json
{
  "emails": [
    {
      "to": "user1@example.com",
      "subject": "Test Email 1",
      "body": "Hello User 1!",
      "is_html": false
    },
    {
      "to": "user2@example.com",
      "subject": "Test Email 2",
      "body": "Hello User 2!",
      "is_html": false
    }
  ]
}
```

#### Email Yönetimi

- **GET** `/api/email/test-connection` - SMTP bağlantısını test eder
- **POST** `/api/email/test/{email}` - Belirtilen emaile test email gönderir
- **GET** `/api/email/config` - Email konfigürasyonunu görüntüler (şifre gizli)

### 📬 Abonelik API'leri

#### Abonelik Yönetimi

- **POST** `/api/subscribe` - Yeni abonelik oluşturur

```json
{
  "email": "user@example.com",
  "categories": ["technology", "news"]
}
```

- **POST** `/api/unsubscribe` - Abonelikten çıkarır

```json
{
  "email": "user@example.com"
}
```

- **POST** `/api/pause` - Aboneliği durdurur

```json
{
  "email": "user@example.com"
}
```

- **POST** `/api/reactivate` - Durdurulan aboneliği aktifleştirir

```json
{
  "email": "user@example.com"
}
```

#### Abonelik Sorgulama

- **GET** `/api/subscriptions/{email}` - Email ile abonelik bilgisi getirir
- **GET** `/api/subscriptions` - Tüm abonelikleri getirir
- **GET** `/api/subscriptions/status/{status}` - Duruma göre abonelikleri getirir (active, paused, unsubscribed)
- **PUT** `/api/subscriptions/categories` - Abonelik kategorilerini günceller

```json
{
  "email": "user@example.com",
  "categories": ["technology", "sports"]
}
```

### 📰 Newsletter API'leri

#### Newsletter Yönetimi

- **POST** `/api/newsletters` - Yeni newsletter oluşturur

```json
{
  "title": "Haftalık Haberler",
  "subject": "Bu Haftanın Haberleri",
  "content": "<h1>Merhaba!</h1><p>Bu hafta önemli gelişmeler...</p>",
  "category": "technology"
}
```

- **GET** `/api/newsletters` - Tüm newsletter'ları getirir
- **GET** `/api/newsletters/{id}` - ID ile newsletter detayı getirir
- **PUT** `/api/newsletters/{id}` - Newsletter günceller
- **DELETE** `/api/newsletters/{id}` - Newsletter siler
- **GET** `/api/newsletters/status/{status}` - Duruma göre newsletter'ları getirir (draft, sent)
- **GET** `/api/newsletters/stats` - Newsletter istatistiklerini getirir

#### Newsletter Gönderimi

- **POST** `/api/newsletters/send-to-emails` - Seçili emaillere newsletter gönderir

```json
{
  "newsletter_id": 1,
  "emails": ["user1@example.com", "user2@example.com", "user3@example.com"]
}
```

- **POST** `/api/newsletters/send-to-subscribers` - Abonelere newsletter gönderir

```json
{
  "newsletter_id": 1,
  "status": "active",
  "category": "technology"
}
```

### 🎨 Template API'leri

#### Template Yönetimi

- **GET** `/api/templates` - Tüm template'leri getirir
- **GET** `/api/templates/{id}` - ID ile template getirir
- **GET** `/api/templates/category/{category}` - Kategoriye göre template'leri getirir (technology, marketing, general)
- **POST** `/api/templates/render` - Template'i render eder

```json
{
  "template_id": "tech_newsletter",
  "variables": {
    "title": "Haftalık Teknoloji Haberleri",
    "content": "<p>Bu hafta önemli gelişmeler...</p>",
    "unsubscribe_link": "https://example.com/unsub"
  }
}
```

## 🎯 Mevcut Template'ler

### 1. `tech_newsletter` - Teknoloji Newsletter

- **Kategori**: technology
- **Tema**: Mavi
- **Variables**: title, content, unsubscribe_link

### 2. `marketing_promo` - Pazarlama Promosyonu

- **Kategori**: marketing
- **Tema**: Yeşil + CTA butonu
- **Variables**: title, content, cta_button, cta_link, unsubscribe_link

### 3. `simple_text` - Basit Metin

- **Kategori**: general
- **Tema**: Minimal
- **Variables**: title, content, unsubscribe_link

## 🗄️ Veritabanı Şeması

### Subscriptions

- `id` - Primary key
- `email` - Unique email adresi
- `status` - Durum (active, paused, unsubscribed)
- `categories` - Kategoriler (array)
- `subscribed_at` - Abonelik tarihi
- `paused_at` - Durdurma tarihi
- `updated_at` - Güncelleme tarihi

### Newsletters

- `id` - Primary key
- `title` - Newsletter başlığı
- `subject` - Email konusu
- `content` - Newsletter içeriği
- `status` - Durum (draft, sent)
- `category` - Kategori
- `created_at` - Oluşturulma tarihi
- `updated_at` - Güncelleme tarihi
- `sent_at` - Gönderilme tarihi

## 🔧 Özellikler

- ✅ **MVC mimarisi**
- ✅ **PostgreSQL** veritabanı
- ✅ **SMTP** email gönderimi
- ✅ **Template** sistemi
- ✅ **Abonelik** yönetimi
- ✅ **Newsletter** oluşturma
- ✅ **Toplu email** gönderimi
- ✅ **Responsive** email tasarımı
- ✅ **Kategori** sistemi
- ✅ **Retry** mekanizması
- ✅ **Rate limiting**
- ✅ **Environment** yapılandırması

## 📝 Kullanım Örnekleri

### Test Email Gönder

```bash
curl -X POST http://localhost:8080/api/email/test/user@example.com
```

### Abone Ol

```bash
curl -X POST http://localhost:8080/api/subscribe \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "categories": ["technology"]}'
```

### Newsletter Oluştur ve Gönder

```bash
# Newsletter oluştur
curl -X POST http://localhost:8080/api/newsletters \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Haftalık Haberler",
    "subject": "Bu Haftanın Haberleri",
    "content": "<h1>Merhaba!</h1><p>Önemli gelişmeler...</p>",
    "category": "technology"
  }'

# Newsletter'ı gönder
curl -X POST http://localhost:8080/api/newsletters/send-to-emails \
  -H "Content-Type: application/json" \
  -d '{
    "newsletter_id": 1,
    "emails": ["user@example.com"]
  }'
```

## 🛠️ Geliştirme

### Yeni Template Ekleme

1. `controllers/email_controller.go` dosyasındaki `renderTemplate` fonksiyonuna yeni template ekleyin
2. `controllers/template_controller.go` dosyasındaki `DefaultTemplates` map'ine ekleyin

### Yeni API Endpoint Ekleme

1. İlgili controller'a yeni fonksiyon ekleyin
2. `main.go` dosyasındaki `initializeRoutes` fonksiyonuna route ekleyin

## 📄 Lisans

MIT License
