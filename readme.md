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
├── test/             # Integration tests
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

# Admin seeder (runs once at startup if no admin exists)
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

Server starts at `http://localhost:8080`. Database tables are created automatically via GORM AutoMigrate.

> On first boot, if `ADMIN_EMAIL` and `ADMIN_PASSWORD` are set and no admin exists yet, an admin account is created automatically.

---

## Roles

| Role      | Description             |
| --------- | ----------------------- |
| `alumni`  | Graduated students      |
| `student` | Current students        |
| `partner` | Company representatives |
| `admin`   | Platform administrators |

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

### Group Forum

> Requires JWT. Read: all roles. Group/article CRUD, membership, reactions: `alumni` and `student` only.

#### Group Management

| Method | Endpoint             | Body        | Description                             |
| ------ | -------------------- | ----------- | --------------------------------------- |
| GET    | `/api/groups`        | —           | List all groups                         |
| POST   | `/api/groups`        | `form-data` | Create group (optional `banner` file)   |
| GET    | `/api/groups/:id`    | —           | Group detail                            |
| PUT    | `/api/groups/:id`    | `form-data` | Update group (owner only)               |
| DELETE | `/api/groups/:id`    | —           | Delete group + all data (owner only)    |
| GET    | `/api/groups/joined` | —           | List groups the current user belongs to |

#### Membership

| Method | Endpoint                          | Description                        |
| ------ | --------------------------------- | ---------------------------------- |
| POST   | `/api/groups/:id/join`            | Join a group                       |
| DELETE | `/api/groups/:id/leave`           | Leave a group (owner cannot leave) |
| GET    | `/api/groups/:id/members`         | List group members                 |
| DELETE | `/api/groups/:id/members/:userID` | Kick a member (owner only)         |

#### Group Articles

| Method | Endpoint                            | Body        | Description                                          |
| ------ | ----------------------------------- | ----------- | ---------------------------------------------------- |
| POST   | `/api/groups/:id/articles`          | `form-data` | Create article (members only, optional `media` file) |
| GET    | `/api/groups/articles/:id`          | —           | Article detail with nested comments                  |
| PUT    | `/api/groups/articles/:id`          | `form-data` | Update own article                                   |
| DELETE | `/api/groups/articles/:id`          | —           | Delete own article                                   |
| POST   | `/api/groups/articles/:id/comments` | JSON        | Add comment or reply                                 |
| POST   | `/api/groups/articles/:id/react`    | JSON        | React to article (members only)                      |
| PUT    | `/api/groups/comments/:id`          | JSON        | Update own comment                                   |
| DELETE | `/api/groups/comments/:id`          | —           | Delete own comment                                   |
| POST   | `/api/groups/comments/:id/react`    | JSON        | React to a comment (members only)                    |

#### Group Fields (form-data)

| Key           | Required | Notes                    |
| ------------- | -------- | ------------------------ |
| `title`       | ✅       | —                        |
| `category`    | ✅       | —                        |
| `description` | ❌       | —                        |
| `rules`       | ❌       | —                        |
| `banner`      | ❌       | file upload → Cloudinary |

#### Article Fields (form-data)

| Key       | Required | Notes                    |
| --------- | -------- | ------------------------ |
| `title`   | ✅       | —                        |
| `content` | ✅       | —                        |
| `media`   | ❌       | file upload → Cloudinary |

#### Group Business Rules

- Non-members can only view the group preview — they cannot create articles or react
- The group owner cannot leave their own group
- Only the group owner can kick members
- When a group is deleted, all articles, comments, and reactions are cascade-deleted
- Reaction types: `like` · `love` · `haha` · `wow` · `sad` · `angry` (same toggle/update logic as Feed)

---

### Events

> Requires JWT. Read/view participants: all roles. Create/update/delete/register/agenda: `alumni` and `student` only.

#### Event Management

| Method | Endpoint          | Body        | Description                              |
| ------ | ----------------- | ----------- | ---------------------------------------- |
| GET    | `/api/events`     | —           | List all events (paginated)              |
| POST   | `/api/events`     | `form-data` | Create event (optional `photo` file)     |
| GET    | `/api/events/:id` | —           | Event detail with agendas & participants |
| PUT    | `/api/events/:id` | `form-data` | Update own event                         |
| DELETE | `/api/events/:id` | —           | Delete event + all data (owner only)     |

#### Event Registration

| Method | Endpoint                       | Description                  |
| ------ | ------------------------------ | ---------------------------- |
| POST   | `/api/events/:id/register`     | Register for an event        |
| DELETE | `/api/events/:id/register`     | Cancel registration          |
| GET    | `/api/events/:id/participants` | List registered participants |

#### Event Agenda

| Method | Endpoint              | Body | Description                     |
| ------ | --------------------- | ---- | ------------------------------- |
| POST   | `/api/events/:id/agenda` | JSON | Add agenda item (owner only)    |
| PUT    | `/api/agenda/:id`     | JSON | Update agenda item (owner only) |
| DELETE | `/api/agenda/:id`     | —    | Delete agenda item (owner only) |

#### Event Fields (form-data)

| Key           | Required | Notes                                                        |
| ------------- | -------- | ------------------------------------------------------------ |
| `title`       | ✅       | —                                                            |
| `description` | ❌       | —                                                            |
| `location`    | ❌       | —                                                            |
| `capacity`    | ❌       | integer, must be zero or positive                            |
| `status`      | ❌       | `upcoming` (default) · `ongoing` · `completed` · `cancelled` |
| `photo`       | ❌       | file upload → Cloudinary                                     |

#### Agenda Fields (JSON)

```json
{
  "description": "Opening Ceremony",
  "agenda_time": "2026-06-01T09:00:00Z"
}
```

#### Event Business Rules

- `partner` role cannot create events, register, or manage agendas
- Registration is **not allowed** when event `status` is `completed` or `cancelled`
- A user can register for the same event **only once**
- When `capacity` is set, it cannot be exceeded
- When an event is deleted, all agendas and registrations are cascade-deleted
- Only the event creator (owner) can update, delete, and manage agendas

---

### Jobs

> Requires JWT. Read: all roles. Create/update/delete: `alumni` and `partner` only. Apply: `alumni` and `student` only.

#### Job Management

| Method | Endpoint       | Body        | Description                               |
| ------ | -------------- | ----------- | ----------------------------------------- |
| GET    | `/api/jobs`    | —           | List jobs (paginated, filter by type/status/search) |
| POST   | `/api/jobs`    | `form-data` | Create job posting (optional `image` file) |
| GET    | `/api/jobs/:id` | —          | Job detail                                |
| PUT    | `/api/jobs/:id` | `form-data` | Update own job posting                   |
| DELETE | `/api/jobs/:id` | —          | Delete own job posting                    |

#### Job Applications

| Method | Endpoint                           | Body        | Description                           |
| ------ | ---------------------------------- | ----------- | ------------------------------------- |
| POST   | `/api/jobs/:id/apply`              | `form-data` | Apply for a job (upload `resume` file or send `resume_url`) |
| GET    | `/api/jobs/:id/applicants`         | —           | List applicants (job owner only)      |
| GET    | `/api/jobs/applications/mine`      | —           | List own job applications             |
| PUT    | `/api/jobs/applications/:id/status` | JSON       | Update application status (owner only) |

#### Job Fields (form-data)

| Key            | Required | Notes                                         |
| -------------- | -------- | --------------------------------------------- |
| `title`        | ✅       | —                                             |
| `company_name` | ✅       | —                                             |
| `job_type`     | ✅       | `full-time` · `part-time` · `internship` · `freelance` · `remote` |
| `description`  | ❌       | —                                             |
| `location`     | ❌       | —                                             |
| `salary_range` | ❌       | e.g. `"5.000.000 - 8.000.000"`               |
| `status`       | ❌       | `open` (default) · `closed`                  |
| `image`        | ❌       | file upload → Cloudinary                      |

#### Application Fields (form-data)

| Key            | Required | Notes                              |
| -------------- | -------- | ---------------------------------- |
| `resume`       | ✅*      | PDF/file upload → Cloudinary       |
| `resume_url`   | ✅*      | URL if not uploading a file        |
| `cover_letter` | ❌       | optional text                      |

> *Either `resume` (file upload) or `resume_url` (link) is required.

#### Application Status Values

`pending` (default) · `reviewed` · `accepted` · `rejected`

#### Job Business Rules

- Only `alumni` and `partner` can create/update/delete job postings
- Only `alumni` and `student` can apply for jobs
- A user can apply to the same job **only once**
- Only the job owner can view applicants and update application status
- When a job is deleted, all applications are cascade-deleted

#### Query Parameters for `GET /api/jobs`

| Param      | Description                                |
| ---------- | ------------------------------------------ |
| `search`   | Search by title or company name            |
| `job_type` | Filter by job type                         |
| `status`   | Filter by `open` or `closed`               |
| `page`     | Page number (default: 1)                   |
| `limit`    | Items per page (default: 10)               |

---

### Content Reporting

> Requires JWT. Available to `alumni`, `student`, and `admin`.

| Method | Endpoint           | Body | Description                   |
| ------ | ------------------ | ---- | ----------------------------- |
| POST   | `/api/reports`     | JSON | Submit a report on content    |
| GET    | `/api/reports/mine` | —   | View own submitted reports    |

#### Report Fields (JSON)

```json
{
  "target_type": "post",
  "target_id": 12,
  "report_type": "spam",
  "description": "Optional — required only for 'other' type"
}
```

#### `target_type` Values

`post` · `comment` · `group` · `group_article` · `event` · `job`

#### `report_type` Values

`harassment` · `violence` · `hate_speech` · `spam` · `inappropriate` · `misinformation` · `copyright` · `other`

> When `report_type` is `other`, `description` is **required**.

#### Report Business Rules

- A user cannot submit a duplicate pending report on the same content
- Reports have status: `pending` → `resolved` or `rejected` (managed by admin)

---

### Admin Module

> Requires JWT + `admin` role for all endpoints except `GET /api/categories` (any authenticated user).

#### Dashboard

| Method | Endpoint               | Description                               |
| ------ | ---------------------- | ----------------------------------------- |
| GET    | `/api/admin/dashboard` | Platform stats (users, posts, groups, events, jobs, pending reports) |

#### User Management

| Method | Endpoint                        | Body | Description                    |
| ------ | ------------------------------- | ---- | ------------------------------ |
| GET    | `/api/admin/users`              | —    | List all users (paginated, filter by `?role=`) |
| GET    | `/api/admin/users/:id`          | —    | User detail                    |
| PATCH  | `/api/admin/users/:id/status`   | JSON | Activate or deactivate a user  |
| PATCH  | `/api/admin/users/:id/role`     | JSON | Change user role               |

```json
// PATCH /api/admin/users/:id/status
{ "is_active": false }

// PATCH /api/admin/users/:id/role
{ "role": "student" }  // alumni | student | partner | admin
```

#### Report Moderation

| Method | Endpoint                          | Body | Description                                    |
| ------ | --------------------------------- | ---- | ---------------------------------------------- |
| GET    | `/api/admin/reports`              | —    | List all reports (filter by `?status=pending`) |
| GET    | `/api/admin/reports/:id`          | —    | Report detail                                  |
| PATCH  | `/api/admin/reports/:id/resolve`  | JSON | Mark resolved (optionally delete the content)  |
| PATCH  | `/api/admin/reports/:id/reject`   | JSON | Reject report with reason                      |

```json
// PATCH /api/admin/reports/:id/resolve
{
  "admin_note": "Content removed — community guidelines violation.",
  "delete_content": true
}

// PATCH /api/admin/reports/:id/reject
{
  "admin_note": "Does not violate our policies."
}
```

#### Direct Content Deletion (without report)

| Method | Endpoint                | Description          |
| ------ | ----------------------- | -------------------- |
| DELETE | `/api/admin/posts/:id`  | Delete any post      |
| DELETE | `/api/admin/groups/:id` | Delete any group     |
| DELETE | `/api/admin/events/:id` | Delete any event     |
| DELETE | `/api/admin/jobs/:id`   | Delete any job       |

> All direct deletions perform the same cascading cleanup as owner-level deletes.

#### Category Management

| Method | Endpoint                    | Body | Description                           |
| ------ | --------------------------- | ---- | ------------------------------------- |
| GET    | `/api/categories`           | —    | List all categories *(public — any authenticated user)* |
| POST   | `/api/admin/categories`     | JSON | Create category (admin only)          |
| PUT    | `/api/admin/categories/:id` | JSON | Update category (admin only)          |
| DELETE | `/api/admin/categories/:id` | —    | Delete category (admin only)          |

```json
// POST/PUT /api/admin/categories
{
  "name": "Technology",
  "description": "Tech-related content"
}
```

#### Admin Business Rules

- `delete_content: true` in resolve-report will cascade-delete the reported post/group/event/job
- A report can only be resolved or rejected **once** (status is immutable after processing)
- Category names must be **unique**
- `is_active: false` deactivates users (flag stored in DB; JWT tokens remain valid until expiry)
- Role changes apply immediately to new logins; existing tokens retain the old role until expiry

---

## Cloudinary Upload Summary

Every file upload is **optional** — omit the field to skip uploading. Accepted formats: `jpg` · `jpeg` · `png` · `webp`

| Module          | Form field | Cloudinary folder                 |
| --------------- | ---------- | --------------------------------- |
| Profile picture | `picture`  | `alumni-platform/profiles`        |
| Feed post image | `image`    | `alumni-platform/feed`            |
| Group banner    | `banner`   | `alumni-platform/groups/banners`  |
| Group article   | `media`    | `alumni-platform/groups/articles` |
| Portfolio item  | `media`    | `alumni-platform/portfolio`       |
| Event photo     | `photo`    | `alumni-platform/events`          |
| Job image       | `image`    | `alumni-platform/jobs`            |
| Job resume      | `resume`   | `alumni-platform/resumes`         |

---

## Response Format

All responses follow a consistent envelope:

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

## Running Tests

Tests are integration tests — the server must be running on port `8080` before executing them.

```bash
# Run all suites
go test -v ./test/...

# Run a specific suite
go test -v -run TestAuth      ./test/...
go test -v -run TestProfile   ./test/...
go test -v -run TestCompany   ./test/...
go test -v -run TestPortfolio ./test/...
go test -v -run TestFeed      ./test/...
go test -v -run TestGroup     ./test/...
go test -v -run TestEvent     ./test/...
go test -v -run TestJob       ./test/...
go test -v -run TestReport    ./test/...
go test -v -run TestAdmin     ./test/...
```

| Suite           | Coverage                                                                              |
| --------------- | ------------------------------------------------------------------------------------- |
| `TestAuth`      | Register (alumni/partner/duplicate), login, JWT `/me` guard                           |
| `TestProfile`   | Profile CRUD, job status variants, experience CRUD                                    |
| `TestCompany`   | Company CRUD, role guard, shared profile joining                                      |
| `TestPortfolio` | Portfolio CRUD, role guard, ownership checks                                          |
| `TestFeed`      | Post CRUD, nested comments, reactions toggle, vote flip                               |
| `TestGroup`     | Group CRUD, membership, articles, nested comments, reactions, kick                    |
| `TestEvent`     | Event CRUD, registration, capacity enforcement, status guards, agenda CRUD            |
| `TestJob`       | Job CRUD, role guards, apply, applicant management, application status update         |
| `TestReport`    | Submit reports, duplicate guard, type validation, `other` + description rule, my reports |
| `TestAdmin`     | Dashboard stats, user status/role, report resolve/reject, direct deletion, category CRUD |

> `TestAdmin` requires the admin account to be seeded first (`ADMIN_EMAIL` + `ADMIN_PASSWORD` in `.env`).

---

## Health Check

```
GET /health
→ { "status": "ok", "message": "Server is running" }
```
