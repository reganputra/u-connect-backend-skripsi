# Alumni Community Platform – Backend API

RESTful API for an alumni networking platform built with **Go**, **Fiber v2**, **GORM**, and **PostgreSQL**.

📖 **Developer & Integration Documentation:**
- **[Full API Reference](docs/API.md)**
- **[Codebase Architecture Guide](ARCHITECTURE.md)**
- **[Mentor Recommendation System (CBF)](RECOMMENDATION_SYSTEM.md)**
- **[WebSockets & Background Schedulers Guide](WEBSOCKETS_AND_SCHEDULER.md)**
- **[Codebase Quality Assessment & Future Improvements](ASSESSMENT.md)**

For frontend teams, read the integration guardrails in
**[Frontend Integration Contract](docs/API.md#frontend-integration-contract)**.

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
├── utils/            # Helpers (env loader, response, image upload, NLP/TF-IDF)
├── test/             # Integration tests
├── docs/             # API documentation
│   └── API.md
├── .env
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

Create a `.env` file:

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

# CORS — comma-separated origins (* = allow all, use specific origins in production)
CORS_ALLOWED_ORIGINS=*

# Admin seeder — runs once on startup if no admin exists
ADMIN_EMAIL=admin@yourapp.com
ADMIN_PASSWORD=yoursecurepassword
ADMIN_NAME=Administrator
```

### 2. Start the database

```bash
docker compose up -d
```

### 3. Run the server

```bash
go run main.go
```

Server starts at `http://localhost:8080`. Tables are auto-migrated on startup.

> On first boot, if `ADMIN_EMAIL` / `ADMIN_PASSWORD` are set and no admin account exists, one is created automatically.

---

## Roles

| Role      | Description             |
| --------- | ----------------------- |
| `alumni`  | Graduated students      |
| `student` | Current students        |
| `partner` | Company representatives |
| `admin`   | Platform administrators |

---

## Modules Overview

| Module              | Base Path                                             | Docs                                         |
| ------------------- | ----------------------------------------------------- | -------------------------------------------- |
| Auth                | `/api/auth`                                           | [→ Auth](docs/API.md#auth)                   |
| User Profile        | `/api/profile`                                        | [→ Profile](docs/API.md#user-profile)        |
| Company Profile     | `/api/company`                                        | [→ Company](docs/API.md#company-profile)     |
| Portfolio           | `/api/portfolio`                                      | [→ Portfolio](docs/API.md#portfolio)         |
| Feed                | `/api/feed`                                           | [→ Feed](docs/API.md#feed-posting)           |
| Group Forum         | `/api/groups`                                         | [→ Groups](docs/API.md#group-forum)          |
| Events              | `/api/events`                                         | [→ Events](docs/API.md#events)               |
| Jobs                | `/api/jobs`                                           | [→ Jobs](docs/API.md#jobs)                   |
| Content Reporting   | `/api/reports`                                        | [→ Reports](docs/API.md#content-reporting)   |
| Notifications       | `/api/notifications`                                  | [→ Notifications](docs/API.md#notifications) |
| Admin               | `/api/admin`                                          | [→ Admin](docs/API.md#admin-module)          |
| Categories (public) | `/api/categories`                                     | [→ Categories](docs/API.md#categories)       |
| **Mentoring**       | `/api/mentors` · `/api/mentor` · `/api/student`       | [→ Mentoring](docs/API.md#mentoring-module)  |
| **Messaging**       | `/api/messages` · `/api/users/:id/follow` · `/api/ws` | [→ Messaging](docs/API.md#message-module)    |

---

## Response Format

All responses use a consistent envelope:

```json
// Success
{ "success": true, "data": { ... } }

// Error
{ "success": false, "error": "error message" }
```

> ⚠️ Error messages are currently in **Bahasa Indonesia**.

---

## Running Tests

Integration tests require the server running on port `8080`.

```bash
# Run all suites
go test -v ./test/...

# Run a single suite
go test -v -run TestAuth      ./test/...
go test -v -run TestProfile   ./test/...
go test -v -run TestCompany   ./test/...
go test -v -run TestPortfolio ./test/...
go test -v -run TestFeed      ./test/...
go test -v -run TestGroup     ./test/...
go test -v -run TestEvent     ./test/...
go test -v -run TestJob       ./test/...
go test -v -run TestReport    ./test/...
go test -v -run TestNotificationFlows ./test/...
go test -v -run TestAdmin     ./test/...
go test -v -run TestMentor    ./test/...
go test -v -run TestFollow    ./test/...
go test -v -run TestMessage   ./test/...
```

> `TestAdmin` requires a seeded admin account (`ADMIN_EMAIL` + `ADMIN_PASSWORD` in `.env`).

---

## Health Check

```
GET /health
→ { "status": "ok", "message": "Server is running" }
```
