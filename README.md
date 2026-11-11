# Golang REST API Starter Project

Project REST API Golang modular dengan arsitektur **Controller - Service - Repository**.

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
│   └── seeder/                         # Logika pengisian data awal (User, Category, Todo)
│       ├── seeder.go
│       ├── user_seeder.go
│       ├── category_seeder.go
│       └── todo_seeder.go
├── internal/
│   ├── controller/                     # Layer Controller HTTP Gin (pengganti handler)
│   │   ├── auth_controller.go
│   │   └── user_controller.go
│   ├── dto/                            # Kontrak Data Transfer Object (Request & Response)
│   │   ├── auth_request.go
│   │   ├── auth_response.go
│   │   ├── user_request.go
│   │   └── user_response.go
│   ├── entity/                         # Struct Entity Model Database (MySQL)
│   │   └── user.go
│   ├── middleware/                     # Middleware JWT & Auth Guard
│   │   └── auth_middleware.go
│   ├── repository/                     # Layer Akses Database MySQL (GORM)
│   │   └── user_repository.go
│   ├── routes/                         # Pendaftaran Rute, Endpoint Grouping, & Middleware
│   │   └── router.go
│   └── service/                        # Layer Business Logic
│       ├── auth_service.go
│       └── user_service.go
├── pkg/
│   ├── database/                       # Inisialisasi koneksi & connection pool MySQL
│   │   └── mysql.go
│   ├── logger/                         # Daily rotating log (lumberjack + slog + Gin writer)
│   │   └── logger.go
│   └── response/                       # Standard JSON response helper
│       └── response.go
├── storage/
│   └── logs/                           # Folder penyimpanan file log harian (app-YYYY-MM-DD.log)
├── test/                               # Folder khusus untuk pengujian Endpoint / Integration test
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
Mengisi data awal untuk 4 User, masing-masing 3 Category, dan setiap Category memiliki 2 Todo (total 24 Todos):
```bash
go run ./cmd/seeder
```

### 4. Jalankan Aplikasi
```bash
go run ./cmd/api
```

### 5. Menjalankan Unit Test
```bash
go test -v ./...
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
| **Auth** | | | |
| `POST` | `/api/auth/register` | Registrasi akun baru | Publik |
| `POST` | `/api/auth/login` | Login & mendapatkan JWT token | Publik |
| `POST` | `/api/auth/logout` | Logout sesi pengguna | Bearer Token |
| **Profile** | | | |
| `GET` | `/api/profile` | Melihat profil pengguna saat ini | Bearer Token |
| `PATCH` | `/api/profile` | Memperbarui profil akun sendiri | Bearer Token |
| **User Management** | | *(Khusus Role Admin)* | |
| `GET` | `/api/user` | Daftar semua pengguna (paginasi & search) | Admin Bearer Token |
| `POST` | `/api/user` | Menambahkan pengguna baru | Admin Bearer Token |
| `GET` | `/api/user/deleted` | Daftar pengguna yang telah di-soft-delete | Admin Bearer Token |
| `GET` | `/api/user/:id` | Detail pengguna berdasarkan ID | Admin Bearer Token |
| `PATCH` | `/api/user/:id` | Update data pengguna | Admin Bearer Token |
| `DELETE` | `/api/user/:id` | Soft delete pengguna | Admin Bearer Token |
| `PATCH` | `/api/user/:id/restore` | Restore pengguna terhapus | Admin Bearer Token |
| `DELETE` | `/api/user/:id/permanent` | Hapus permanen pengguna | Admin Bearer Token |
| **Category Management** | | *(Creator Scoped - 404 jika bukan miliknya)* | |
| `GET` | `/api/category` | Daftar category milik user login | Bearer Token |
| `POST` | `/api/category` | Buat category baru | Bearer Token |
| `GET` | `/api/category/deleted` | Daftar category terhapus (trash) milik user | Bearer Token |
| `GET` | `/api/category/:id` | Detail category | Bearer Token |
| `PATCH` | `/api/category/:id` | Update category | Bearer Token |
| `DELETE` | `/api/category/:id` | Soft delete category | Bearer Token |
| `PATCH` | `/api/category/:id/restore` | Restore category terhapus | Bearer Token |
| `DELETE` | `/api/category/:id/permanent` | Hapus permanen category | Bearer Token |
| **Todo Management** | | *(Creator Scoped - 404 jika bukan miliknya)* | |
| `GET` | `/api/todo` | Daftar todo milik user (filter category_id, is_completed, search) | Bearer Token |
| `POST` | `/api/todo` | Buat todo baru (opsional category_id) | Bearer Token |
| `GET` | `/api/todo/deleted` | Daftar todo terhapus (trash) milik user | Bearer Token |
| `GET` | `/api/todo/:id` | Detail todo | Bearer Token |
| `PATCH` | `/api/todo/:id` | Update todo (title wajib diisi, description & category_id nullable) | Bearer Token |
| `PATCH` | `/api/todo/:id/complete` | Toggle status selesai/belum selesai (`is_completed`) | Bearer Token |
| `DELETE` | `/api/todo/:id` | Soft delete todo | Bearer Token |
| `PATCH` | `/api/todo/:id/restore` | Restore todo terhapus | Bearer Token |
| `DELETE` | `/api/todo/:id/permanent` | Hapus permanen todo | Bearer Token |
