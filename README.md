# Mindora

Mindora adalah aplikasi produktivitas mahasiswa yang dirancang untuk mendukung manajemen aktivitas akademik dan pengembangan diri, dengan arsitektur terpisah antara frontend dan backend serta siap untuk integrasi Machine Learning (ML) dan Large Language Model (LLM) pada tahap berikutnya.

## Teknologi

### Frontend
- **Framework / Library:** React.js (v19)
- **Build Tool:** Vite
- **Language:** TypeScript
- **Styling:** Tailwind CSS (v3)
- **Routing:** React Router DOM (v7)

### Backend
- **Language:** Go (Golang)
- **Web Framework:** Go Fiber (v2)
- **Configuration & Env:** Godotenv
- **Middleware:** Fiber CORS Middleware

## Struktur Direktori

```text
mindora/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── handlers/
│   │   │   └── health.go
│   │   ├── middleware/
│   │   │   └── cors.go
│   │   ├── models/
│   │   │   └── health.go
│   │   ├── repositories/
│   │   │   └── repository.go
│   │   ├── routes/
│   │   │   └── routes.go
│   │   └── services/
│   │       └── service.go
│   ├── pkg/
│   │   └── response/
│   │       └── response.go
│   ├── .env.example
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   │   └── index.ts
│   │   ├── hooks/
│   │   │   └── index.ts
│   │   ├── layouts/
│   │   │   └── RootLayout.tsx
│   │   ├── pages/
│   │   │   └── HomePage.tsx
│   │   ├── services/
│   │   │   └── api.ts
│   │   ├── types/
│   │   │   └── index.ts
│   │   ├── utils/
│   │   │   └── index.ts
│   │   ├── App.tsx
│   │   ├── index.css
│   │   └── main.tsx
│   ├── .env.example
│   ├── package.json
│   ├── postcss.config.js
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   └── vite.config.ts
├── docs/
│   └── README.md
├── .gitignore
└── README.md
```

## Persyaratan Instalasi

Pastikan sistem telah terinstal:
- **Node.js** (v18 ke atas disarankan) dan **npm**
- **Go** (v1.20 ke atas disarankan)

## Panduan Menjalankan Frontend

1. Masuk ke direktori `frontend`:
   ```bash
   cd frontend
   ```

2. Salin environment variable (jika belum ada file `.env`):
   ```bash
   cp .env.example .env
   ```

3. Pasang dependensi:
   ```bash
   npm install
   ```

4. Jalankan development server:
   ```bash
   npm run dev
   ```

5. Buka browser pada URL:
   `http://localhost:5173`

## Panduan Menjalankan Backend

1. Masuk ke direktori `backend`:
   ```bash
   cd backend
   ```

2. Salin environment variable (jika belum ada file `.env`):
   ```bash
   cp .env.example .env
   ```

3. Unduh dependensi Go:
   ```bash
   go mod tidy
   ```

4. Jalankan backend server:
   ```bash
   go run ./cmd/api
   ```

5. Backend aktif di port 8080:
   `http://localhost:8080`

## URL Aplikasi & Endpoint Pengujian

- **Frontend URL:** `http://localhost:5173`
- **Backend Base URL:** `http://localhost:8080`
- **Health Check Endpoint:** `GET http://localhost:8080/api/health`

### Response Health Check:
```json
{
  "success": true,
  "message": "Mindora API is running"
}
```
