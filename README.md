# Golang REST API Starter Project

Project REST API Golang modular dengan arsitektur **Controller - Service - Repository** (mirip konsep Spring Boot).

---

## 📁 Struktur Folder

```text
test-golang-project/
├── cmd/
│   ├── api/
│   │   └── main.go                     # Entry point & perakitan dependency injection
│   └── seeder/
│       └── main.go                     # Entry point database seeder
├── config/
│   └── config.go                       # Konfigurasi dari file .env
├── database/
│   ├── migrations/                     # File skema database untuk Goose (.sql)
│   │   └── 00001_create_users_table.sql
│   └── seeder/                         # Logika pengisian data awal (Admin, dummy user)
│       ├── seeder.go
│       └── user_seeder.go
├── internal/
│   ├── controller/                     # Layer Controller HTTP Gin (pengganti handler)
│   │   └── user_controller.go
│   ├── entity/                         # Struct Entity & DTO (Request/Response)
│   │   └── user.go
│   ├── middleware/                     # Middleware JWT & Auth Guard
│   │   └── auth_middleware.go
│   ├── repository/                     # Layer Akses Database MySQL (GORM)
│   │   └── user_repository.go
│   └── service/                        # Layer Business Logic
│       ├── user_service.go
│       └── user_service_test.go        # Unit Test menggunakan Testify Mock
├── pkg/
│   ├── database/                       # Inisialisasi koneksi & connection pool MySQL
│   │   └── mysql.go
│   ├── logger/                         # Daily rotating log (lumberjack + slog + Gin writer)
│   │   └── logger.go
│   └── response/                       # Standard JSON response helper
│       └── response.go
├── storage/
│   └── logs/                           # Folder penyimpanan file log harian (app-YYYY-MM-DD.log)
├── .env.example
├── .env
├── .gitignore
├── go.mod
└── go.sum
```

---

## 🚀 Cara Menjalankan

### 1. Konfigurasi Environment (.env)
Salin `.env.example` ke `.env` dan sesuaikan nilainya:
```env
APP_PORT=8080
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=golang_db
JWT_SECRET=rahasia-jwt-anda

# Pengaturan Logging Harian
LOG_DIR=storage/logs
LOG_LEVEL=info
LOG_MAX_SIZE=10
LOG_MAX_BACKUPS=14
LOG_MAX_AGE=14
LOG_COMPRESS=true
```

### 2. Jalankan Migrasi Database (Goose)
Pastikan tool Goose sudah terinstall:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Jalankan perintah migrasi ke database MySQL:
```bash
goose -dir database/migrations mysql "root:password@tcp(127.0.0.1:3306)/test_db?parseTime=true" up
```
Untuk rollback migrasi:
```bash
goose -dir database/migrations mysql "root:password@tcp(127.0.0.1:3306)/test_db?parseTime=true" down
```

### 3. Jalankan Database Seeder (Opsional)
Mengisi data awal pengguna (Admin & Test User dengan password bcrypt):
```bash
go run ./cmd/seeder
# Atau via Makefile:
make seed
```

### 4. Jalankan Aplikasi
```bash
go run ./cmd/api
# Atau via Makefile:
make run
```

### 5. Menjalankan Unit Test
```bash
go test -v ./...
# Atau via Makefile:
make test
```

---

## 📝 Logging Harian (Daily Log)

Log otomatis ditulis ke terminal (Console) dan ke file log harian di folder `storage/logs/app-YYYY-MM-DD.log`.
Untuk mencatat log di service/controller:
```go
import "log/slog"

slog.Info("Informasi penting", "user_id", 123)
slog.Warn("Peringatan", "path", "/api/v1/auth")
slog.Error("Terjadi error", "error", err)
```
Rotasi log dan retensi otomatis dikelola oleh library `lumberjack` (default retensi: 14 hari dengan kompresi gzip).

---

## 📌 Endpoint API

| Method | Endpoint | Keterangan | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | Health Check status server & DB | Publik |
| `POST` | `/api/v1/auth/register` | Registrasi akun baru | Publik |
| `POST` | `/api/v1/auth/login` | Login & mendapatkan JWT token | Publik |
| `GET` | `/api/v1/profile` | Melihat profil pengguna saat ini | Bearer Token |
