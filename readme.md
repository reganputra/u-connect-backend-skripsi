# Alumni Community Platform – Backend API

RESTful API for an alumni networking platform built with **Go**, **Fiber v2**, **GORM**, and **PostgreSQL**.

---

## Tech Stack

| Layer     | Technology                                        |
| --------- | ------------------------------------------------- |
| Framework | [Fiber v2](https://gofiber.io/)                   |
| ORM       | [GORM](https://gorm.io/)                          |
| Database  | PostgreSQL (via Docker)                           |
| Auth      | JWT (`gofiber/contrib/jwt` + `golang-jwt/jwt/v5`) |
| Storage   | [Cloudinary](https://cloudinary.com/)             |
| Language  | Go 1.25                                           |

---

## Project Structure

```
backend-skripsi/
├── config/           # DB and Cloudinary initialization
├── controllers/      # HTTP handlers
├── middleware/       # JWT auth & role guard
├── models/           # GORM models
├── repository/       # Database queries
├── routes/           # Route registration
├── service/          # Business logic & validation
├── utils/            # Helpers (env loader, response, image upload)
├── .env              # Environment variables
├── docker-compose.yaml
└── main.go
```

---

## Getting Started

### Prerequisites

- Go 1.21+
- Docker Desktop

### 1. Clone & configure

```bash
git clone <repo-url>
cd backend-skripsi
```

Copy `.env` and fill in your values:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=rootpassword
DB_NAME=alumni_community_db
DB_SSLMODE=disable

APP_PORT=8080
JWT_SECRET=your_secret_key
CLOUDINARY_URL=cloudinary://api_key:api_secret@cloud_name
```

### 2. Start the database

```bash
docker compose up -d
```

### 3. Run the server

```bash
go run main.go
```

Server starts at `http://localhost:8080`. Database tables are created automatically via GORM AutoMigrate.

---

## Roles

| Role      | Description             |
| --------- | ----------------------- |
| `alumni`  | Graduated students      |
| `student` | Current students        |
| `partner` | Company representatives |

---

## API Reference

### Auth

| Method | Endpoint             | Auth | Description                      |
| ------ | -------------------- | ---- | -------------------------------- |
| POST   | `/api/auth/register` | ❌   | Register a new user              |
| POST   | `/api/auth/login`    | ❌   | Login, returns JWT token         |
| GET    | `/api/me`            | ✅   | Get current user info from token |

#### Register – Alumni / Student

```json
{
  "name": "Regan Putra",
  "email": "regan@test.com",
  "password": "secret123",
  "role": "alumni",
  "faculty": "Engineering",
  "major": "Informatics",
  "year_enroll": 2020
}
```

#### Register – Partner

```json
{
  "name": "Budi Santoso",
  "email": "budi@company.com",
  "password": "secret123",
  "role": "partner",
  "company_name": "PT Maju Bersama"
}
```

#### Login

```json
{
  "email": "regan@test.com",
  "password": "secret123"
}
```

Response includes `token`. Use as: `Authorization: Bearer <token>`

---

### User Profile

> Requires JWT. For `alumni` and `student` only.

| Method | Endpoint                      | Description                        |
| ------ | ----------------------------- | ---------------------------------- |
| POST   | `/api/profile`                | Create profile                     |
| GET    | `/api/profile`                | Get own profile                    |
| PUT    | `/api/profile`                | Partial update                     |
| DELETE | `/api/profile`                | Delete profile                     |
| POST   | `/api/profile/picture`        | Upload profile picture (multipart) |
| POST   | `/api/profile/experience`     | Add work experience                |
| PUT    | `/api/profile/experience/:id` | Update experience                  |
| DELETE | `/api/profile/experience/:id` | Delete experience                  |

#### Profile Fields by `job_status`

| `job_status`               | Required Fields                                                   |
| -------------------------- | ----------------------------------------------------------------- |
| `employed`                 | `position`, `company_name`                                        |
| `entrepreneur`             | `industry_name`                                                   |
| `continuing_study`         | `educational_level`, `advanced_study_program`, `institution_name` |
| `unemployed` / `freelance` | `status_description`                                              |
| `student`                  | _(none extra)_                                                    |

> Fields irrelevant to the active `job_status` are automatically set to `null`.

#### Experience Entry

```json
{
  "company_name": "PT Startup Indonesia",
  "position": "Backend Developer",
  "start_year": 2021,
  "end_year": 2023,
  "description": "Built REST APIs using Go and PostgreSQL."
}
```

> `end_year: null` means currently working there.

#### Profile Picture Upload

- `POST /api/profile/picture`
- Body: `form-data`, key: `picture`, value: image file (jpg, jpeg, png, webp)
- Stored on Cloudinary, returns a `picture_url`

---

### Company Profile

> Requires JWT + `partner` role only.

| Method | Endpoint       | Description                             |
| ------ | -------------- | --------------------------------------- |
| POST   | `/api/company` | Create or join existing company profile |
| GET    | `/api/company` | View company profile                    |
| PUT    | `/api/company` | Update company profile                  |
| DELETE | `/api/company` | Delete company profile                  |

Partners with the same `company_name` (set at registration) **share one company profile**. `POST /api/company` returns `201` if newly created, `200` if joined an existing one.

#### Company Profile Fields

```json
{
  "industry_type": "B2B",
  "location": "Jakarta, Indonesia",
  "employee_size": 150,
  "website_url": "https://company.com"
}
```

> `employee_size` must be zero or positive.

---

## Response Format

All responses follow a consistent structure:

**Success**

```json
{
  "success": true,
  "data": { ... }
}
```

**Error**

```json
{
  "success": false,
  "error": "message describing the error"
}
```

---

## Health Check

```
GET /health
→ { "status": "ok", "message": "Server is running" }
```
