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

| Method | Endpoint                      | Body        | Description                              |
| ------ | ----------------------------- | ----------- | ---------------------------------------- |
| POST   | `/api/profile`                | `form-data` | Create profile (optional `picture` file) |
| GET    | `/api/profile`                | —           | Get own profile                          |
| PUT    | `/api/profile`                | `form-data` | Partial update (optional `picture` file) |
| DELETE | `/api/profile`                | —           | Delete profile                           |
| POST   | `/api/profile/experience`     | JSON        | Add work experience                      |
| PUT    | `/api/profile/experience/:id` | JSON        | Update experience                        |
| DELETE | `/api/profile/experience/:id` | —           | Delete experience                        |

> **All profile create/update requests use `multipart/form-data`.** Include a `picture` file field to upload/change the profile picture in the same request.

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

### Portfolio

> Requires JWT + `alumni` or `student` role only.

| Method | Endpoint             | Body        | Description                         |
| ------ | -------------------- | ----------- | ----------------------------------- |
| POST   | `/api/portfolio`     | `form-data` | Create item (optional `media` file) |
| GET    | `/api/portfolio`     | —           | List own portfolio items            |
| PUT    | `/api/portfolio/:id` | `form-data` | Update item (optional `media` file) |
| DELETE | `/api/portfolio/:id` | —           | Delete item                         |

> **All portfolio create/update requests use `multipart/form-data`.** Include a `media` file field to upload/change the media in the same request.

#### Portfolio Item Fields

```json
{
  "title": "Alumni Mobile App",
  "description": "A cross-platform app for alumni networking.",
  "category": "Mobile Development",
  "tags": "Flutter, Dart, Firebase",
  "start_date": "2023-01",
  "end_date": "2023-06"
}
```

- `title` is **required**
- `start_date` / `end_date` format: `YYYY-MM`
- `tags` is comma-separated
- Add a `media` key with an image file in the same `form-data` request to attach media

---

## Response Format

### Feed Posting

> Requires JWT. Read/react/vote: all roles. Post/comment CRUD: `alumni` and `student` only.

| Method | Endpoint                  | Body        | Description                           |
| ------ | ------------------------- | ----------- | ------------------------------------- |
| GET    | `/api/feed`               | —           | List posts (paginated, counts only)   |
| GET    | `/api/feed/:id`           | —           | Post detail with full nested comments |
| POST   | `/api/feed`               | `form-data` | Create post (optional `image` file)   |
| PUT    | `/api/feed/:id`           | `form-data` | Update own post                       |
| DELETE | `/api/feed/:id`           | —           | Delete own post                       |
| POST   | `/api/feed/:id/comments`  | JSON        | Add comment or reply to a post        |
| POST   | `/api/feed/:id/react`     | JSON        | React to post (toggle/change)         |
| POST   | `/api/feed/:id/vote`      | JSON        | Vote post (toggle/flip)               |
| PUT    | `/api/comments/:id`       | JSON        | Update own comment                    |
| DELETE | `/api/comments/:id`       | —           | Delete own comment                    |
| POST   | `/api/comments/:id/react` | JSON        | React to a comment or reply           |
| POST   | `/api/comments/:id/vote`  | JSON        | Vote on a comment or reply            |

#### Post Fields (form-data)

| Key        | Required | Notes                    |
| ---------- | -------- | ------------------------ |
| `title`    | ✅       | —                        |
| `content`  | ✅       | —                        |
| `category` | ❌       | optional                 |
| `image`    | ❌       | file upload → Cloudinary |

#### Comment / Reply

```json
{ "content": "Great post!" }
```

To reply to a comment, include `parent_comment_id`:

```json
{ "content": "Thanks!", "parent_comment_id": 8 }
```

> Replies can be nested infinitely.

#### Reaction Types

`like` · `love` · `haha` · `wow` · `sad` · `angry`

- Same type again → **removed**
- Different type → **updated**

#### Vote Values

`1` = upvote · `-1` = downvote

- Same value again → **removed**
- Opposite value → **flipped**

#### Feed List Response (`GET /api/feed`)

Returns `comment_count`, `reaction_count`, `vote_count` — no full arrays.

#### Feed Detail Response (`GET /api/feed/:id`)

Returns full nested comment tree (infinite depth) with reactions and votes at every level.

---

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
