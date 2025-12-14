# Go-Middle

Go-Middle adalah library middleware untuk Gin Framework yang menyediakan verifikasi JWT token dengan remote public key. Library ini memungkinkan aplikasi Go untuk memverifikasi JWT token menggunakan public key yang diambil dari URL remote secara otomatis.

## Daftar Isi

- [Fitur](#-fitur)
- [Instalasi](#-instalasi)
- [Konfigurasi](#-konfigurasi)
- [Penggunaan](#-penggunaan)
- [Struktur Proyek](#-struktur-proyek)
- [API Documentation](#-api-documentation)
- [Contoh Penggunaan](#-contoh-penggunaan)
- [Lisensi](#-lisensi)
- [Kontribusi](#-kontribusi)

## Fitur

- **JWT Token Verification**: Memverifikasi JWT token menggunakan RSA public key
- **Remote Public Key**: Mengambil public key dari URL remote secara otomatis
- **Auto Refresh**: Public key di-refresh secara berkala untuk memastikan keamanan
- **Thread Safe**: Implementasi yang aman untuk penggunaan concurrent
- **Environment Variable Support**: Konfigurasi melalui environment variables
- **Gin Middleware**: Terintegrasi langsung dengan Gin framework

## Instalasi

```bash
go get github.com/digitcodestudiotech/go-middle
```

### Dependencies

Library ini menggunakan beberapa dependencies utama:

- `github.com/gin-gonic/gin v1.11.0` - Web framework
- `github.com/golang-jwt/jwt/v5 v5.3.0` - JWT library
- `github.com/joho/godotenv v1.5.1` - Environment variable loader

## Konfigurasi

### Environment Variables

Buat file `.env` berdasarkan `.env.example`:

```bash
cp .env.example .env
```

Kemudian konfigurasi variable berikut:

```env
PUBLIC_KEY_URL=http://localhost:3000/keys/public.pem
```

### Environment Variables yang Tersedia

| Variable | Deskripsi | Required | Default |
|----------|-----------|----------|---------|
| `PUBLIC_KEY_URL` | URL untuk mengambil RSA public key dalam format PEM | ✅ | - |

## Penggunaan

### Basic Usage

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/digitcodestudiotech/go-middle/middleware"
)

func main() {
    r := gin.Default()
    
    // Gunakan middleware untuk memverifikasi JWT token
    r.Use(middleware.VerifyToken())
    
    r.GET("/protected", func(c *gin.Context) {
        // Ambil claims dari context
        claims, exists := c.Get("claims")
        if !exists {
            c.JSON(401, gin.H{"error": "No claims found"})
            return
        }
        
        c.JSON(200, gin.H{
            "message": "Access granted",
            "claims": claims,
        })
    })
    
    r.Run(":8080")
}
```

### Menggunakan pada Route Tertentu

```go
// Hanya pada route tertentu
protected := r.Group("/api/protected")
protected.Use(middleware.VerifyToken())
{
    protected.GET("/user", getUserHandler)
    protected.POST("/data", postDataHandler)
}
```

### Mengakses Claims

Setelah token berhasil diverifikasi, claims JWT akan tersedia dalam Gin context:

```go
func protectedHandler(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(401, gin.H{"error": "No claims found"})
        return
    }
    
    // Cast ke jwt.MapClaims
    jwtClaims := claims.(jwt.MapClaims)
    
    userID := jwtClaims["user_id"].(string)
    email := jwtClaims["email"].(string)
    
    c.JSON(200, gin.H{
        "user_id": userID,
        "email": email,
    })
}
```

## Struktur Proyek

```
go-middle/
├── .env.example          # Contoh konfigurasi environment
├── .gitignore           # Git ignore file
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── LICENSE              # Lisensi GPL v3 (Bahasa Indonesia)
├── crypto/              # Package untuk cryptography
│   └── key.go          # Remote public key management
├── middleware/          # Package middleware Gin
│   └── verify.go       # JWT verification middleware
└── utils/              # Package utilities
    └── env.go          # Environment variable utilities
```

### Penjelasan File

#### `/crypto/key.go`
File ini mengimplementasikan `RemotePublicKey` struct yang bertugas:
- Mengambil RSA public key dari URL remote
- Melakukan auto-refresh key secara berkala (default: 5 menit)
- Thread-safe access menggunakan RWMutex
- Parsing PEM format ke RSA public key

#### `/middleware/verify.go`
Middleware utama yang menyediakan:
- JWT token verification menggunakan Bearer token format
- Integration dengan Gin framework
- Error handling untuk berbagai skenario (missing token, invalid format, expired token)
- Menyimpan JWT claims ke Gin context

#### `/utils/env.go`
Utility functions untuk:
- Loading environment variables dari file `.env`
- Fallback ke system environment jika `.env` tidak ditemukan
- Warning logging untuk environment variables yang tidak diset

## API Documentation

### `middleware.VerifyToken()`

Fungsi utama yang mengembalikan Gin middleware handler.

**Return**: `gin.HandlerFunc`

**Behavior**:
1. Load environment variables
2. Inisialisasi remote public key dari `PUBLIC_KEY_URL`
3. Return middleware function yang memverifikasi setiap request

### `crypto.NewRemotePublicKey(url, refreshEvery)`

Membuat instance baru dari `RemotePublicKey`.

**Parameters**:
- `url` (string): URL untuk mengambil public key
- `refreshEvery` (time.Duration): Interval refresh key

**Return**: 
- `*RemotePublicKey`: Instance remote public key
- `error`: Error jika terjadi kesalahan

### HTTP Response Codes

| Code | Deskripsi |
|------|-----------|
| `401` | Token tidak valid, expired, atau format authorization header salah |
| `200` | Token valid, request dilanjutkan ke handler berikutnya |

### Error Response Format

```json
{
  "error": "error message description"
}
```

**Possible Error Messages**:
- `"missing authorization header"` - Header Authorization tidak ada
- `"invalid authorization format"` - Format bukan "Bearer <token>"
- `"invalid or expired token"` - Token tidak valid atau sudah expired

## Contoh Penggunaan

### 1. Server dengan Public Key Endpoint

```go
// Server yang menyediakan public key
func setupKeyServer() {
    r := gin.Default()
    
    r.GET("/keys/public.pem", func(c *gin.Context) {
        // Return public key dalam format PEM
        c.Header("Content-Type", "application/x-pem-file")
        c.String(200, publicKeyPEM)
    })
    
    r.Run(":3000")
}
```

### 2. Client Request dengan JWT Token

```bash
# Request tanpa token (akan gagal)
curl http://localhost:8080/protected

# Request dengan token
curl -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
     http://localhost:8080/protected
```

### 3. Multiple Middleware

```go
r.Use(
    gin.Logger(),
    gin.Recovery(),
    middleware.VerifyToken(), // JWT verification
    // middleware lainnya...
)
```

## Keamanan

### Best Practices

1. **HTTPS Only**: Selalu gunakan HTTPS untuk `PUBLIC_KEY_URL` di production
2. **Key Rotation**: Implementasikan key rotation pada auth server
3. **Token Expiration**: Set expiration time yang sesuai pada JWT token
4. **Secure Storage**: Jangan simpan private key di repository

### Auto Refresh Mechanism

Public key akan di-refresh secara otomatis setiap 5 menit untuk memastikan:
- Key rotation yang smooth
- Security yang up-to-date
- Minimal downtime saat key berubah

## Troubleshooting

### Error: "PUBLIC_KEY_URL is required in .env"
**Solusi**: Pastikan file `.env` ada dan berisi `PUBLIC_KEY_URL`

### Error: "failed loading remote public key"
**Penyebab Umum**:
- URL tidak dapat diakses
- Format PEM tidak valid
- Key bukan RSA public key

**Solusi**: 
- Periksa konektivitas ke URL
- Validasi format PEM key
- Pastikan menggunakan RSA public key

### Error: "invalid or expired token"
**Penyebab Umum**:
- Token sudah expired
- Token di-sign dengan private key yang berbeda
- Token format tidak valid

## Lisensi

Proyek ini dilisensikan di bawah [GNU General Public License v3.0](LICENSE).

Lisensi ini memungkinkan Anda untuk:
- ✅ Menggunakan secara komersial
- ✅ Memodifikasi kode
- ✅ Mendistribusikan
- ✅ Menggunakan secara pribadi

Dengan kewajiban:
- 📋 Menyertakan lisensi dan copyright
- 📋 Menyatakan perubahan
- 📋 Menggunakan lisensi yang sama untuk derivative works
- 📋 Menyediakan source code

## Kontribusi

Kami menyambut kontribusi! Silakan ikuti langkah berikut:

1. Fork repository ini
2. Buat feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit perubahan (`git commit -m 'Add some AmazingFeature'`)
4. Push ke branch (`git push origin feature/AmazingFeature`)
5. Buat Pull Request

### Development Setup

```bash
# Clone repository
git clone https://github.com/digitcodestudiotech/go-middle.git
cd go-middle

# Install dependencies
go mod tidy

# Copy environment file
cp .env.example .env

# Edit .env sesuai kebutuhan
# Jalankan tests (jika ada)
go test ./...
```

## Support

Jika Anda menemukan bug atau membutuhkan bantuan:

1. Cek [Issues](https://github.com/digitcodestudiotech/go-middle/issues) yang sudah ada
2. Buat issue baru jika belum ada
3. Sertakan informasi lengkap (Go version, error message, dll.)

## Versioning

Proyek ini menggunakan [Semantic Versioning](https://semver.org/). Untuk versi yang tersedia, lihat [tags di repository ini](https://github.com/digitcodestudiotech/go-middle/tags).

---

**Made with ❤️ by [DigitCode Studio Tech](https://github.com/digitcodestudiotech)**