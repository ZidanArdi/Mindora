# Mindora

Mindora adalah platform aplikasi produktivitas akademik mahasiswa yang mengintegrasikan manajemen tugas, pelacakan sesi belajar, fokus Pomodoro, pencatatan gaya hidup harian, sistem rekomendasi berbasis aturan, dan mesin prediksi estimasi produktivitas berbasis machine learning (Ridge Regression).

## Arsitektur & Teknologi

- **Frontend:** React 19, TypeScript, Vite, Tailwind CSS, Lucide Icons, Axios
- **Backend:** Go (Golang 1.26.1), Fiber v2, GORM, PostgreSQL, JWT Authentication, RBAC
- **Machine Learning:** Ridge Regression (L2-Regularized OLS) dievaluasi secara native di Go (runtime mandiri tanpa ketergantungan runtime Python)
- **Database:** PostgreSQL (15 migrasi terkelola)

## Struktur Repositori

```text
Mindora App/
├── Backend/                 # Layanan API Go Fiber & GORM
│   ├── cmd/                 # CLI migrasi & seeding admin
│   ├── internal/            # Core logic (handlers, services, repositories, models, middleware)
│   ├── migrations/          # 15 berkas migrasi SQL PostgreSQL
│   ├── ml/                  # Referensi model & skrip internal
│   ├── pkg/                 # Package utilitas bersama (JWT, crypto, OAuth, validator)
│   ├── .env.example         # Template konfigurasi environment backend
│   ├── go.mod               # Modul Go & dependensi
│   └── main.go              # Entrypoint server API
├── Frontend/                # Aplikasi web antarmuka React
│   ├── src/
│   │   ├── components/      # Komponen UI modular
│   │   ├── features/        # Fitur terisolasi (tugas, belajar, fokus, analitik, rekomendasi, dll.)
│   │   ├── layouts/         # Layout utama aplikasi
│   │   ├── lib/             # Klien API & helper
│   │   ├── pages/           # Halaman rute aplikasi
│   │   └── types/           # Definisi tipe TypeScript
│   ├── .env.example         # Template konfigurasi environment frontend
│   ├── package.json         # Dependensi Node.js
│   └── vite.config.ts       # Konfigurasi bundler Vite
├── ML/                      # Subsistem Machine Learning
│   ├── datasets/            # Dataset acuan tantangan prediksi
│   ├── models/              # Artefak model (model.joblib, model_metadata.json)
│   ├── scripts/             # Skrip pelatihan (train.py) & inferensi CLI (infer.py)
│   ├── requirements.txt     # Dependensi Python untuk pelatihan
│   └── README.md            # Dokumentasi model & bobot matematis
├── .gitignore               # Konfigurasi pengabaian Git
└── README.md                # Dokumentasi utama proyek
```

## Panduan Menjalankan Aplikasi

### 1. Backend (Go API)

```bash
cd Backend
# Salin konfigurasi jika belum ada
cp .env.example .env
# Jalankan server
go run .
```
Backend aktif pada `http://localhost:8080`.

### 2. Frontend (React)

```bash
cd Frontend
# Salin konfigurasi jika belum ada
cp .env.example .env
# Pasang dependensi
npm install
# Jalankan dev server
npm run dev
```
Frontend aktif pada `http://localhost:5173`.

### 3. Machine Learning (Inferensi CLI / Pelatihan)

```bash
cd ML
python scripts/infer.py --study 4.0 --sleep 7.0 --phone 2.5 --stress 4.0
```
*Catatan: Backend Go mengevaluasi model ML secara native tanpa memerlukan instalasi Python pada lingkungan produksi.*
