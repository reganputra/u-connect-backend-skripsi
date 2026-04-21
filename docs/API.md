# API Reference – Alumni Community Platform

Base URL: `http://localhost:8080`  
All protected endpoints require: `Authorization: Bearer <token>`  
Request body must include `Content-Type: application/json` for JSON endpoints or `Content-Type: multipart/form-data` for file/form endpoints.

---

## Table of Contents

- [Response Format](#response-format)
- [Frontend Integration Contract](#frontend-integration-contract)
- [Auth](#auth)
- [User Profile](#user-profile)
- [Directory](#directory)
- [Company Profile](#company-profile)
- [Portfolio](#portfolio)
- [Feed Posting](#feed-posting)
- [Group Forum](#group-forum)
- [Events](#events)
- [Jobs](#jobs)
- [Content Reporting](#content-reporting)
- [Notifications](#notifications)
- [Admin Module](#admin-module)
- [Categories](#categories)
- [Mentoring Module](#mentoring-module)
- [Message Module](#message-module)
- [Cloudinary Upload Reference](#cloudinary-upload-reference)

---

## Response Format

Every response wraps data in a consistent envelope.

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
  "error": "error description"
}
```

> ⚠️ Error messages are currently in **Bahasa Indonesia**.

**Paginated responses** include:

```json
{
  "success": true,
  "data": {
    "total": 42,
    "page": 1,
    "limit": 10,
    "data": [ ... ]   // or "posts", "events", "jobs" depending on module
  }
}
```

---

## Frontend Integration Contract

This section defines the practical integration contract for frontend clients.

### 1) Stable Envelope

- Success responses always use:
  - `{"success": true, "data": ...}`
- Error responses always use:
  - `{"success": false, "error": "..."}`

### 2) Field Naming and Serialization

- The API currently returns mixed key casing depending on endpoint and DTO/model source.
- Some endpoints return snake_case keys (for example `user_id`, `unread_count`).
- Many resource payloads return PascalCase keys from GORM models (for example `ID`, `CreatedAt`, `NotificationType`).
- Frontend must read keys exactly as returned by each endpoint example in this document.
- Recommended frontend strategy: normalize server payloads into app-level camelCase in one mapper layer.

### 3) Date-Time and Nullability

- Timestamps are serialized as ISO 8601 strings by Go JSON encoder.
- Nullable backend fields may be returned as `null`.
- Frontend should treat optional fields as nullable and avoid assuming empty strings.

### 4) Pagination and Sorting Defaults

- Most paginated endpoints use `page` + `limit` query params.
- If not provided, backend applies per-endpoint defaults (commonly `page=1`).
- Notification list hard limits: `limit <= 100`, invalid/empty values are normalized server-side.
- Sort order is endpoint-specific. When building UI lists, rely on returned order from server.

### 5) Error Handling Contract

- Use HTTP status as primary error classifier:
  - `400`: validation/business rule
  - `401`: auth missing/invalid/expired
  - `403`: role/ownership forbidden
  - `404`: resource not found
  - `409`: conflict (for example duplicate follow)
- Error message text is human-readable and currently in Bahasa Indonesia.
- Frontend should not hardcode business logic using exact message text unless unavoidable.

### 6) Auth Lifecycle Contract

- Auth uses Bearer JWT in `Authorization` header for HTTP endpoints.
- No refresh-token endpoint is documented; frontend should handle `401` by redirecting to login or forcing re-auth.
- WebSocket auth uses query token: `ws://localhost:8080/api/ws?token=<jwt>`.

### 7) WebSocket Event Contract

- WebSocket URL: `/api/ws?token=<jwt>`.
- Outgoing event envelope from server includes `type` and optional `data`/`message`.

Server event types:

- `message`
  - `data` contains persisted message payload.
- `notification`
  - `data` contains persisted notification payload.
- `error`
  - `message` contains error text.

Notes:

- Sender receives a message echo confirmation after successful send.
- Notifications are persisted first, then best-effort pushed over WebSocket.
- If user is offline, frontend can fetch missed notifications/messages via REST.

### 8) File Upload Constraints

- Image uploads use Cloudinary with allowed formats:
  - `jpg`, `jpeg`, `png`, `webp`
- Resume/raw file upload uses Cloudinary raw upload with no app-level extension whitelist.
- This backend does not define an explicit app-level max file size in docs; effective limits may come from Fiber, reverse proxy, or Cloudinary account policy.

### 9) Event Reminder Preconditions

- Event reminder notifications depend on event `start_time`.
- To receive reminder notifications, frontend should provide `start_time` when creating/updating events.

---

## Auth

### POST `/api/auth/register` — Register

**No auth required.**

**Body (JSON):**

```json
// Alumni / Student
{
  "name": "Regan Putra",
  "email": "regan@test.com",
  "password": "secret123",
  "role": "alumni",
  "faculty": "Engineering",
  "major": "Informatics",
  "year_enroll": 2020
}

// Partner
{
  "name": "Budi Santoso",
  "email": "budi@company.com",
  "password": "secret123",
  "role": "partner"
}
```

> For `partner`, `company_name` is optional at registration. If omitted, set it later during company onboarding via `POST /api/company`.

**Response `201`:**

```json
{
  "success": true,
  "data": {
    "ID": 1,
    "Name": "Regan Putra",
    "Email": "regan@test.com",
    "Role": "alumni",
    "IsActive": true,
    "Faculty": "Engineering",
    "Major": "Informatics",
    "YearEnroll": 2020
  }
}
```

---

### POST `/api/auth/login` — Login

**No auth required.**

**Body (JSON):**

```json
{
  "email": "regan@test.com",
  "password": "secret123"
}
```

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "name": "Regan Putra",
      "email": "regan@test.com",
      "role": "student",
      "is_active": true,
      "picture_url": "",
      "faculty": "Engineering",
      "major": "Informatics",
      "year_enroll": 2020,
      "company_name": null
    }
  }
}
```

> `token` is returned on login only.

**JWT Payload (decoded):**

```json
{
  "user_id": 1,
  "email": "regan@test.com",
  "role": "alumni",
  "exp": 1712345678
}
```

---

### GET `/api/me` — Current User

**Auth required.**

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "name": "Regan Putra",
      "email": "regan@test.com",
      "role": "student",
      "is_active": true,
      "picture_url": "",
      "faculty": "Engineering",
      "major": "Informatics",
      "year_enroll": 2020,
      "company_name": null
    },
    "user_id": 1,
    "email": "regan@test.com",
    "role": "student"
  }
}
```

> For backward compatibility, `user_id`, `email`, and `role` are still included at the top level of `data`.
> For `partner` accounts, the nested `user` object also includes `company_name` so the frontend can detect company onboarding without an extra `/api/company` call.

---

## User Profile

> `alumni`, `student`, and `partner`. All create/update use `multipart/form-data`.

| Method | Endpoint                      | Auth | Description             |
| ------ | ----------------------------- | ---- | ----------------------- |
| POST   | `/api/profile`                | ✅   | Create profile          |
| GET    | `/api/profile`                | ✅   | Get own profile         |
| PUT    | `/api/profile`                | ✅   | Update profile          |
| DELETE | `/api/profile`                | ✅   | Delete profile          |
| POST   | `/api/profile/experience`     | ✅   | Add experience entry    |
| PUT    | `/api/profile/experience/:id` | ✅   | Update experience entry |
| DELETE | `/api/profile/experience/:id` | ✅   | Delete experience entry |

### POST/PUT `/api/profile` — Create / Update Profile

**Content-Type:** `multipart/form-data`

| Field                      | Required                    | Notes                                                 |
| -------------------------- | --------------------------- | ----------------------------------------------------- |
| `job_status`               | ✅                          | See status matrix below                               |
| `bio`                      | ❌                          | Short profile bio                                     |
| `location`                 | ❌                          | Free-text location                                    |
| `picture`                  | ❌                          | Image file → Cloudinary                               |
| `skills`                   | ❌                          | Comma-separated text (used by mentor recommendation)  |
| `interests`                | ❌                          | Comma-separated text (used by mentor recommendation)  |
| `position`                 | if `employed`               | Current job title                                     |
| `company_name`             | if `employed`               | Current company name                                  |
| `company_location`         | ❌                          | Company/work location (relevant for `employed`)       |
| `company_size`             | ❌                          | Integer (relevant for `entrepreneur` / self-employed) |
| `industry_name`            | if `entrepreneur`           | Industry/business domain                              |
| `industry_type`            | ❌                          | Example: B2B, B2C, SaaS                               |
| `year_founding`            | ❌                          | Integer year                                          |
| `salary`                   | ❌                          | Integer                                               |
| `educational_level`        | if `continuing_study`       | Degree level                                          |
| `advanced_study_program`   | if `continuing_study`       | Study program                                         |
| `institution_name`         | if `continuing_study`       | University/institution name                           |
| `expected_graduation_year` | ❌                          | Integer year (mainly for `continuing_study`)          |
| `mentor_quota`             | ❌                          | Integer (alumni mentor profile only)                  |
| `mentor_description`       | ❌                          | Mentor bio/description                                |
| `status_description`       | if `unemployed`/`freelance` | Required explanation for current status               |

**`job_status` values:** `employed` · `entrepreneur` · `continuing_study` · `unemployed` · `freelance` · `student`

**Job status requirement matrix:**

| `job_status`       | Required fields                                                   |
| ------------------ | ----------------------------------------------------------------- |
| `employed`         | `position`, `company_name`                                        |
| `entrepreneur`     | `industry_name`                                                   |
| `continuing_study` | `educational_level`, `advanced_study_program`, `institution_name` |
| `unemployed`       | `status_description`                                              |
| `freelance`        | `status_description`                                              |
| `student`          | No additional required fields                                     |

**Field clearing behavior when status changes:**

- If status is not `employed`, backend clears `position` and `company_location`.
- If status is not `employed` and not `entrepreneur`, backend clears `company_name`.
- If status is not `entrepreneur`, backend clears `company_size`, `industry_name`, `industry_type`, and `year_founding`.
- If status is not `continuing_study`, backend clears `educational_level`, `advanced_study_program`, `institution_name`, and `expected_graduation_year`.
- If status is not `unemployed` or `freelance`, backend clears `status_description`.

**Validation notes:**

- Invalid `job_status` returns `400`.
- Missing status-dependent required fields return `400` with a business-rule message.
- Numeric fields (`salary`, `year_founding`, `expected_graduation_year`, `mentor_quota`) are parsed as integers from form-data.

**Response `201` (create) / `200` (update):**

```json
{
  "success": true,
  "data": {
    "ID": 1,
    "UserID": 1,
    "Bio": "Software engineer at tech startup",
    "Location": "Bandung, Indonesia",
    "JobStatus": "employed",
    "Position": "Backend Developer",
    "CompanyName": "PT Startup Indonesia",
    "CompanyLocation": "Jakarta, Indonesia",
    "CompanySize": null,
    "IndustryName": null,
    "IndustryType": null,
    "YearFounding": null,
    "Salary": 12000000,
    "EducationalLevel": null,
    "AdvancedStudyProgram": null,
    "InstitutionName": null,
    "ExpectedGraduationYear": null,
    "Skills": "Go, PostgreSQL, Docker",
    "Interests": "Backend, Distributed Systems",
    "MentorQuota": null,
    "MentorDescription": null,
    "StatusDescription": null,
    "PictureURL": "https://res.cloudinary.com/.../profile.jpg",
    "CreatedAt": "2026-04-06T10:00:00+07:00",
    "UpdatedAt": "2026-04-06T10:00:00+07:00"
  }
}
```

### POST `/api/profile/experience` — Add Experience

**Body (JSON):**

```json
{
  "company_name": "PT Startup Indonesia",
  "position": "Backend Developer",
  "start_year": 2021,
  "end_year": 2023,
  "description": "Built REST APIs using Go."
}
```

> `end_year: null` = currently working there.

---

## Directory

> Browse profiles of students, alumni, and partners; search by skills, company, or interests. Available to `student`, `alumni`, and `partner` roles.

| Method | Endpoint                           | Auth | Description                                      |
| ------ | ---------------------------------- | ---- | ------------------------------------------------ |
| GET    | `/api/directory/:userID`           | ✅   | View a user's full public profile                |
| GET    | `/api/directory/:userID/portfolio` | ✅   | View a user's public portfolio                   |
| GET    | `/api/directory`                   | ✅   | List all profiles (paginated)                    |
| GET    | `/api/directory/search?q=<query>`  | ✅   | Search profiles by name/skills/company           |
| GET    | `/api/directory/role/:role`        | ✅   | Filter profiles by role (student/alumni/partner) |

#### GET `/api/directory/:userID` — View Public Profile

Returns the full public profile for a selected user from the directory.
Response shape is intentionally the same as `GET /api/profile` so frontend can reuse the same detail view.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "ID": 12,
    "CreatedAt": "2026-04-11T10:00:00Z",
    "UpdatedAt": "2026-04-11T10:00:00Z",
    "DeletedAt": null,
    "UserID": 12,
    "User": {
      "ID": 12,
      "Name": "Dev Sofia",
      "Email": "sofia@example.com",
      "Role": "alumni",
      "IsActive": true,
      "Faculty": "Engineering",
      "Major": "Informatics",
      "YearEnroll": 2020,
      "CompanyName": null
    },
    "ProfilePicture": "https://res.cloudinary.com/.../profile-12.jpg",
    "Bio": "Python and AI enthusiast",
    "Location": "Surabaya",
    "JobStatus": "employed",
    "Position": "ML Engineer",
    "CompanyName": "AI Startup",
    "CompanyLocation": "Jakarta",
    "CompanySize": null,
    "IndustryName": null,
    "IndustryType": null,
    "YearFounding": null,
    "Salary": 15000000,
    "EducationalLevel": null,
    "AdvancedStudyProgram": null,
    "InstitutionName": null,
    "ExpectedGraduationYear": null,
    "Skills": "Python, Machine Learning, TensorFlow",
    "Interests": "AI, Data Science",
    "MentorQuota": 3,
    "MentorDescription": "Available for mentoring in backend and AI",
    "StatusDescription": null,
    "Experiences": []
  }
}
```

#### GET `/api/directory/:userID/portfolio` — View Public Portfolio

Returns portfolio items for a selected user.

**Response `200`:**

```json
{
  "success": true,
  "data": [
    {
      "ID": 21,
      "CreatedAt": "2026-04-10T08:00:00Z",
      "UpdatedAt": "2026-04-10T08:00:00Z",
      "DeletedAt": null,
      "UserID": 12,
      "Title": "Backend API Project",
      "Description": "Built a REST API using Go",
      "Category": "Backend",
      "Tags": "go, api, gorm",
      "StartDate": "2025-01",
      "EndDate": "2025-03",
      "MediaURL": "https://res.cloudinary.com/.../portfolio.jpg",
      "Link": "https://github.com/example/project"
    }
  ]
}
```

**Error cases:**
| Status | Reason |
|---|---|
| `400` | Invalid user ID |
| `401` | Not authenticated |
| `404` | Profile not found |

#### GET `/api/directory` — Browse All Profiles

**Query params:** `?page=1&limit=20`

Returns paginated list of all student, alumni, and partner profiles (newest first).

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 127,
    "page": 1,
    "limit": 20,
    "data": [
      {
        "user_id": 5,
        "name": "Alumni Regan",
        "role": "alumni",
        "profile_picture": "https://res.cloudinary.com/.../alumni-profile-5.jpg",
        "bio": "Software engineer passionate about backend development",
        "location": "Jakarta",
        "job_status": "employed",
        "position": "Senior Backend Engineer",
        "company_name": "Tech Company",
        "skills": "Go, TypeScript, PostgreSQL, Docker",
        "interests": "Cloud architecture, Open source",
        "mentor_description": "Available for mentoring in backend development"
      },
      {
        "user_id": 8,
        "name": "Student Budi",
        "role": "student",
        "profile_picture": "https://res.cloudinary.com/.../student-profile-8.jpg",
        "bio": "Final year student, interested in web development",
        "location": "Bandung",
        "job_status": "continuing_study",
        "position": null,
        "company_name": null,
        "skills": "React, Node.js, JavaScript",
        "interests": "Web development, UI/UX",
        "mentor_description": null
      }
    ]
  }
}
```

#### GET `/api/directory/search?q=<query>` — Search Profiles

Searches profiles by name, skills, company name, or interests (case-insensitive).

**Query params:**

- `q` (required): Search term (name, skill, company, interest)
- `page` (optional, default=1): Page number
- `limit` (optional, default=20): Results per page (max 100)

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 5,
    "page": 1,
    "limit": 20,
    "query": "Python",
    "data": [
      {
        "user_id": 12,
        "name": "Dev Sofia",
        "role": "alumni",
        "profile_picture": "...",
        "bio": "Python and AI enthusiast",
        "location": "Surabaya",
        "job_status": "employed",
        "position": "ML Engineer",
        "company_name": "AI Startup",
        "skills": "Python, Machine Learning, TensorFlow",
        "interests": null,
        "mentor_description": null
      }
    ]
  }
}
```

#### GET `/api/directory/role/:role` — Filter by Role

Returns all profiles for a specific role (`student`, `alumni`, or `partner`).

**Query params:** `?page=1&limit=20`

**Response `200`:** (same structure as `GET /api/directory`)

**Error cases:**
| Status | Reason |
|---|---|
| `400` | Role is not `student`, `alumni`, or `partner` |
| `401` | Not authenticated |

---

## Company Profile

> `partner` role only.

| Method | Endpoint                   | Auth | Description                    |
| ------ | -------------------------- | ---- | ------------------------------ |
| POST   | `/api/company`             | ✅   | Create or join company profile |
| GET    | `/api/company`             | ✅   | View own company profile       |
| PUT    | `/api/company`             | ✅   | Update company profile         |
| PATCH  | `/api/company/affiliation` | ✅   | Change company affiliation     |
| DELETE | `/api/company`             | ✅   | Delete company profile         |

> Partners with the same `company_name` **share one profile**. POST returns `201` if created, `200` if joined.
>
> If partner account has no `company_name` yet, include it in the first `POST /api/company` request.
> Company onboarding also ensures a minimal `user_profiles` record exists, so partner accounts can appear in directory listings.

### POST `/api/company` — Create or Join Company Profile

**Body (JSON):**

| Field           | Required | Notes                                                                       |
| --------------- | -------- | --------------------------------------------------------------------------- |
| `company_name`  | Cond.    | Required only for first-time partner onboarding when account value is empty |
| `description`   | ❌       | Company description                                                         |
| `industry_type` | ❌       | Industry type                                                               |
| `location`      | ❌       | Company location                                                            |
| `employee_size` | ❌       | Integer >= 0                                                                |
| `website_url`   | ❌       | Company website URL                                                         |

```json
{
  "company_name": "PT Maju Bersama",
  "description": "Leading campus-to-industry collaboration programs.",
  "industry_type": "Technology",
  "location": "Jakarta, Indonesia",
  "employee_size": 150,
  "website_url": "https://company.com"
}
```

> `company_name` is required only for first-time partner onboarding when account-level `company_name` is still empty. After it is set, subsequent updates can omit it.

**Response behavior:**

- `201 Created`: new company profile created.
- `200 OK`: existing company profile found and partner joined/shared that profile.

### PUT `/api/company` — Update Company Profile

Updates profile content (`description`, `industry_type`, `location`, `employee_size`, `website_url`) for the current company.

> `PUT /api/company` does **not** change `company_name` affiliation.

### PATCH `/api/company/affiliation` — Change Company Affiliation

Use this endpoint when partner wants to switch to a different company.

**Body (JSON):**

```json
{
  "company_name": "PT Baru Nusantara"
}
```

Behavior:

- If target company profile already exists, partner joins that company profile.
- If target company profile does not exist, backend creates a new company profile with that name.
- Partner account `company_name` is updated to the new value.

> Use this endpoint whenever partner wants to switch company identity. Do not use `PUT /api/company` for renaming/switching company.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "joined_existing": true,
    "company": {
      "ID": 10,
      "CompanyName": "PT Baru Nusantara"
    }
  }
}
```

---

## Portfolio

> `alumni` and `student` only. Create/update use `multipart/form-data`.

| Method | Endpoint             | Auth | Description           |
| ------ | -------------------- | ---- | --------------------- |
| POST   | `/api/portfolio`     | ✅   | Create portfolio item |
| GET    | `/api/portfolio`     | ✅   | List own items        |
| PUT    | `/api/portfolio/:id` | ✅   | Update item           |
| DELETE | `/api/portfolio/:id` | ✅   | Delete item           |

**Fields (form-data):**

| Field         | Required | Notes                         |
| ------------- | -------- | ----------------------------- |
| `title`       | ✅       | —                             |
| `description` | ❌       | —                             |
| `category`    | ❌       | e.g. `"Mobile Development"`   |
| `tags`        | ❌       | Comma-separated               |
| `start_date`  | ❌       | Format: `YYYY-MM`             |
| `end_date`    | ❌       | Format: `YYYY-MM`             |
| `media`       | ❌       | Image file → Cloudinary       |
| `link`        | ❌       | External URL (non-Cloudinary) |

---

## Feed Posting

> All roles can read/react/vote. `alumni` and `student` can create posts/comments.

### Endpoints

| Method | Endpoint                  | Auth | Body        | Description                         |
| ------ | ------------------------- | ---- | ----------- | ----------------------------------- |
| GET    | `/api/feed`               | ✅   | —           | List posts (paginated, counts only) |
| GET    | `/api/feed/:id`           | ✅   | —           | Full post with nested comments      |
| POST   | `/api/feed`               | ✅   | `form-data` | Create post                         |
| PUT    | `/api/feed/:id`           | ✅   | `form-data` | Update own post                     |
| DELETE | `/api/feed/:id`           | ✅   | —           | Delete own post                     |
| POST   | `/api/feed/:id/comments`  | ✅   | JSON        | Add comment or reply                |
| POST   | `/api/feed/:id/react`     | ✅   | JSON        | React to a post                     |
| POST   | `/api/feed/:id/vote`      | ✅   | JSON        | Vote on a post                      |
| PUT    | `/api/comments/:id`       | ✅   | JSON        | Update own comment                  |
| DELETE | `/api/comments/:id`       | ✅   | —           | Delete own comment                  |
| POST   | `/api/comments/:id/react` | ✅   | JSON        | React to a comment                  |
| POST   | `/api/comments/:id/vote`  | ✅   | JSON        | Vote on a comment                   |

### POST `/api/feed` — Create Post

**Content-Type:** `multipart/form-data`

| Field      | Required | Notes                                                |
| ---------- | -------- | ---------------------------------------------------- |
| `title`    | ✅       | —                                                    |
| `content`  | ✅       | —                                                    |
| `category` | ❌       | Free text                                            |
| `images`   | ❌       | Multiple image files (form array) → Cloudinary       |
| `image`    | ❌       | Single image file (legacy, kept for backward compat) |

> **Note:** You can send multiple images using the `images` field (e.g., `images[0]`, `images[1]`). The `image` field is still supported for backward compatibility with older clients. If both are provided, `images` takes precedence.

### GET `/api/feed` — List Posts (paginated)

**Query params:** `?page=1&limit=10`

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 25,
    "page": 1,
    "limit": 10,
    "data": [
      {
        "id": 3,
        "user_id": 2,
        "user": { "ID": 2, "Name": "Regan Putra", "Role": "alumni" },
        "title": "Hello everyone",
        "content": "My first post!",
        "category": "General",
        "image_url": "https://res.cloudinary.com/...",
        "image_urls": [
          "https://res.cloudinary.com/.../image1.jpg",
          "https://res.cloudinary.com/.../image2.jpg"
        ],
        "comment_count": 4,
        "reaction_count": 7,
        "vote_count": 3,
        "created_at": "2026-03-04T19:25:24+07:00"
      }
    ]
  }
}
```

> **Note:** `image_url` contains the first image (for backward compatibility). Use `image_urls` array for all images.

### GET `/api/feed/:id` — Post Detail

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "id": 3,
    "user_id": 2,
    "title": "Hello everyone",
    "content": "My first post!",
    "category": "General",
    "image_url": "https://res.cloudinary.com/.../image1.jpg",
    "image_urls": [
      "https://res.cloudinary.com/.../image1.jpg",
      "https://res.cloudinary.com/.../image2.jpg"
    ],
    "user": { "ID": 2, "Name": "Regan Putra", "Role": "alumni" },
    "reactions": [{ "ID": 1, "UserID": 2, "Type": "like" }],
    "votes": [{ "ID": 1, "UserID": 3, "Value": 1 }],
    "comments": [
      {
        "id": 1,
        "content": "Great post!",
        "user": { "ID": 3, "Name": "Ani" },
        "reactions": [],
        "votes": [],
        "replies": [
          {
            "id": 2,
            "content": "Agreed!",
            "parent_comment_id": 1,
            "replies": []
          }
        ]
      }
    ],
    "created_at": "2026-03-04T19:25:24+07:00"
  }
}
```

> **Note:** `image_url` contains the first image (for backward compatibility). Use `image_urls` array for all images.

### PUT `/api/feed/:id` — Update Post

**Content-Type:** `multipart/form-data`

| Field      | Required | Notes                                                |
| ---------- | -------- | ---------------------------------------------------- |
| `title`    | ❌       | Updated title (if provided)                          |
| `content`  | ❌       | Updated content (if provided)                        |
| `category` | ❌       | Updated category (if provided)                       |
| `images`   | ❌       | New multiple image files (replaces all old images)   |
| `image`    | ❌       | Single image file (legacy, kept for backward compat) |

> **Authentication:** Only the post owner can update.  
> **Note:** If new images are provided, all previous images are deleted and replaced with the new ones.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "id": 3,
    "user_id": 2,
    "title": "Updated title",
    "content": "Updated content!",
    "category": "General",
    "image_url": "https://res.cloudinary.com/.../new-image.jpg",
    "image_urls": ["https://res.cloudinary.com/.../new-image.jpg"]
  }
}
```

---

### POST `/api/feed/:id/react` — React to Post

```json
{ "type": "like" }
```

> Types: `like` · `love` · `haha` · `wow` · `sad` · `angry`  
> Same type → **removed**. Different type → **updated**.

### POST `/api/feed/:id/vote` — Vote on Post

```json
{ "value": 1 }
```

> Values: `1` (upvote) · `-1` (downvote)  
> Same value → **removed**. Opposite → **flipped**.

### Vote Count Semantics

The `vote_count` field in post and comment list responses represents the **net vote score**, not the total number of votes cast:

- `vote_count = 1`: net +1 (example: 5 upvotes, 4 downvotes = net +1)
- `vote_count = -1`: net -1 (example: 1 upvote, 2 downvotes = net -1)
- `vote_count = 0`: equal upvotes and downvotes, or no votes at all

**Calculation:** `vote_count = Σ(vote values)` where each vote is `+1` (upvote) or `-1` (downvote).

### POST `/api/feed/:id/comments` — Add Comment or Reply

```json
{ "content": "Great post!" }
```

To reply to an existing comment:

```json
{ "content": "Thanks!", "parent_comment_id": 8 }
```

---

## Group Forum

> All authenticated users can read group previews and group details. `alumni` and `student` can create/join/interact; `partner` and `admin` are blocked from join/create/member actions.

### Group Endpoints

| Method | Endpoint                          | Auth | Body        | Description                                          |
| ------ | --------------------------------- | ---- | ----------- | ---------------------------------------------------- |
| GET    | `/api/groups`                     | ✅   | —           | List all groups                                      |
| POST   | `/api/groups`                     | ✅   | `form-data` | Create group                                         |
| GET    | `/api/groups/joined`              | ✅   | —           | Groups current user belongs to (student/alumni only) |
| GET    | `/api/groups/:id`                 | ✅   | —           | Group detail                                         |
| PUT    | `/api/groups/:id`                 | ✅   | `form-data` | Update group (owner only)                            |
| DELETE | `/api/groups/:id`                 | ✅   | —           | Delete group + all data                              |
| POST   | `/api/groups/:id/join`            | ✅   | —           | Join group (student/alumni only)                     |
| DELETE | `/api/groups/:id/leave`           | ✅   | —           | Leave group (student/alumni only)                    |
| GET    | `/api/groups/:id/members`         | ✅   | —           | List members                                         |
| DELETE | `/api/groups/:id/members/:userID` | ✅   | —           | Kick member (owner only; student/alumni only)        |

Pagination for group list endpoints:

- `GET /api/groups?page=1&limit=20`
- `GET /api/groups/joined?page=1&limit=20`
- `GET /api/groups/:id/members?page=1&limit=20`

Each paginated response includes `total`, `page`, and `limit`.

- `GET /api/groups` response shape: `{ total, page, limit, data }`
- `GET /api/groups/joined` response shape: `{ total, page, limit, data }`
- `GET /api/groups/:id/members` response shape: `{ total, page, limit, members }`

### Group Article Endpoints

| Method | Endpoint                            | Auth | Body        | Description                                       |
| ------ | ----------------------------------- | ---- | ----------- | ------------------------------------------------- |
| POST   | `/api/groups/:id/articles`          | ✅   | `form-data` | Create article (student/alumni members only)      |
| GET    | `/api/groups/articles/:id`          | ✅   | —           | Article detail (comments visible to members only) |
| PUT    | `/api/groups/articles/:id`          | ✅   | `form-data` | Update own article (owner or group owner)         |
| DELETE | `/api/groups/articles/:id`          | ✅   | —           | Delete own article (owner or group owner)         |
| POST   | `/api/groups/articles/:id/comments` | ✅   | JSON        | Add comment or reply (members only)               |
| POST   | `/api/groups/articles/:id/react`    | ✅   | JSON        | React to article (members only)                   |
| PUT    | `/api/groups/comments/:id`          | ✅   | JSON        | Update own comment                                |
| DELETE | `/api/groups/comments/:id`          | ✅   | —           | Delete own comment                                |
| POST   | `/api/groups/comments/:id/react`    | ✅   | JSON        | React to comment (members only)                   |

### Group/Article Count Fields

- `GET /api/groups` includes per-group list item fields:
  - `member_count` (total active members)
  - `article_count` (total active articles)
- `GET /api/groups/:id` includes:
  - `member_count` (total active members)
  - `article_count` (total active articles)
  - `Articles[].comment_count` (total comments for each article in the `Articles` array)
  - `Articles[].media_urls` (array of all image URLs for each article)
- `GET /api/groups/articles/:id` includes:
  - `comment_count` (total visible comments in the response)
  - `media_urls` (array of all image URLs for the article)

### POST `/api/groups/:id/articles` — Create Article (form-data)

| Field     | Required | Notes                                                |
| --------- | -------- | ---------------------------------------------------- |
| `title`   | ✅       | —                                                    |
| `content` | ✅       | —                                                    |
| `medias`  | ❌       | Multiple image files (form array) → Cloudinary       |
| `media`   | ❌       | Single image file (legacy, kept for backward compat) |

> **Note:** You can send multiple images using the `medias` field (e.g., `medias[0]`, `medias[1]`). The `media` field is still supported for backward compatibility with older clients. If both are provided, `medias` takes precedence.

Accepted multipart key variants for multi-image upload:

- `medias`
- `medias[]`
- `medias[0]`, `medias[1]`, etc.

If one or more media files fail to upload or fail to persist, the API returns an error and the article creation is not treated as successful.

### PUT `/api/groups/articles/:id` — Update Article (form-data)

| Field     | Required | Notes                                                |
| --------- | -------- | ---------------------------------------------------- |
| `title`   | ❌       | Updated title (if provided)                          |
| `content` | ❌       | Updated content (if provided)                        |
| `medias`  | ❌       | New multiple image files (replaces all old images)   |
| `media`   | ❌       | Single image file (legacy, kept for backward compat) |

> **Note:** If new images are provided via `medias` or `media`, all previous images are deleted and replaced with the new ones.

Accepted multipart key variants for multi-image upload:

- `medias`
- `medias[]`
- `medias[0]`, `medias[1]`, etc.

If one or more media files fail to upload or fail to persist, the API returns an error and the update is not treated as successful.

### POST `/api/groups` — Create Group (form-data)

| Field         | Required | Notes                   |
| ------------- | -------- | ----------------------- |
| `title`       | ✅       | —                       |
| `category`    | ✅       | —                       |
| `description` | ❌       | —                       |
| `rules`       | ❌       | —                       |
| `banner`      | ❌       | Image file → Cloudinary |

**Response `201`:**

```json
{
  "success": true,
  "data": {
    "ID": 1,
    "Title": "Data Science Indonesia",
    "Category": "Technology",
    "Description": "Group for DS practitioners",
    "OwnerID": 2,
    "BannerURL": null,
    "CreatedAt": "2026-03-10T10:00:00+07:00"
  }
}
```

### Business Rules

- Non-members can only view group preview — cannot post or react
- Group detail is public to authenticated users, but article comments are hidden for non-members
- Group owner cannot leave their own group
- When a group is deleted: articles, comments, reactions, memberships all cascade-deleted

**Join behavior:** leaving a group and joining again is supported; the previous membership is restored rather than creating a duplicate record.

---

## Events

> All roles can read. `alumni` and `student` can create/register. `partner` is blocked.

### Endpoints

| Method | Endpoint                       | Auth | Body        | Description                              |
| ------ | ------------------------------ | ---- | ----------- | ---------------------------------------- |
| GET    | `/api/events`                  | ✅   | —           | List events (paginated)                  |
| POST   | `/api/events`                  | ✅   | `form-data` | Create event                             |
| GET    | `/api/events/:id`              | ✅   | —           | Event detail with agendas & participants |
| PUT    | `/api/events/:id`              | ✅   | `form-data` | Update own event                         |
| DELETE | `/api/events/:id`              | ✅   | —           | Delete event + cascade                   |
| POST   | `/api/events/:id/register`     | ✅   | —           | Register for event                       |
| DELETE | `/api/events/:id/register`     | ✅   | —           | Cancel registration                      |
| GET    | `/api/events/:id/participants` | ✅   | —           | List participants                        |
| POST   | `/api/events/:id/agenda`       | ✅   | JSON        | Add agenda item (owner only)             |
| PUT    | `/api/agenda/:id`              | ✅   | JSON        | Update agenda item (owner only)          |
| DELETE | `/api/agenda/:id`              | ✅   | —           | Delete agenda item (owner only)          |

### POST `/api/events` — Create Event (form-data)

| Field         | Required | Notes                                                        |
| ------------- | -------- | ------------------------------------------------------------ |
| `title`       | ✅       | —                                                            |
| `organizer`   | ❌       | Event organizer name                                         |
| `description` | ❌       | —                                                            |
| `location`    | ❌       | —                                                            |
| `capacity`    | ❌       | Integer ≥ 0 (0 = unlimited)                                  |
| `start_time`  | ❌       | ISO 8601 date-time; triggers auto-status transitions         |
| `end_time`    | ❌       | ISO 8601 date-time; if omitted, defaults to start_time + 24h |
| `status`      | ❌       | `upcoming` (default) · `ongoing` · `completed` · `cancelled` |
| `photo`       | ❌       | Image file → Cloudinary                                      |

> `PUT /api/events/:id` accepts the same form-data fields, including optional `start_time`.

### Event Response Fields

| Field            | Type          | Notes                                                                                                                  |
| ---------------- | ------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `AttendantCount` | Integer       | Total registered attendants for the event                                                                              |
| `SeatLeft`       | Integer\|null | Remaining seats (`capacity - AttendantCount`) when `capacity > 0`; `null` when unlimited (`capacity` is `null` or `0`) |

**Response `201`:**

```json
{
  "success": true,
  "data": {
    "ID": 5,
    "Title": "Alumni Seminar 2026",
    "Organizer": "Himpunan Alumni Teknik Informatika",
    "Location": "Aula Besar, Kampus A",
    "StartTime": "2026-06-01T09:00:00Z",
    "Capacity": 50,
    "AttendantCount": 0,
    "SeatLeft": 50,
    "Status": "upcoming",
    "PhotoURL": null,
    "UserID": 2,
    "CreatedAt": "2026-04-01T08:00:00+07:00"
  }
}
```

### GET `/api/events` — List Events

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 2,
    "page": 1,
    "limit": 10,
    "events": [
      {
        "ID": 5,
        "Title": "Alumni Seminar 2026",
        "Capacity": 50,
        "AttendantCount": 37,
        "SeatLeft": 13,
        "Status": "upcoming"
      },
      {
        "ID": 9,
        "Title": "Open Networking Night",
        "Capacity": 0,
        "AttendantCount": 120,
        "SeatLeft": null,
        "Status": "ongoing"
      }
    ]
  }
}
```

### GET `/api/events/:id` — Event Detail

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "ID": 5,
    "Title": "Alumni Seminar 2026",
    "Capacity": 50,
    "AttendantCount": 37,
    "SeatLeft": 13,
    "Status": "upcoming",
    "Agendas": [],
    "Registrations": []
  }
}
```

### POST `/api/events/:id/agenda` — Add Agenda (JSON)

```json
{
  "description": "Opening Ceremony",
  "agenda_time": "2026-06-01T09:00:00Z"
}
```

### Business Rules

- Event status lifecycle: `upcoming` → `ongoing` (auto at start_time) → `completed` (auto at end_time or start_time + 24h)
- Event reminders are scheduled from `start_time` approximately 24 hours before the event
- `start_time` triggers auto-transition to `ongoing`; provide both `start_time` and `end_time` for full auto-lifecycle
- If `end_time` is omitted, events auto-complete 24 hours after `start_time`
- Manual status override is allowed (e.g., organizer can set `status: ongoing` before start_time)
- Registration blocked when `status` is `completed` or `cancelled`
- Maximum registrations enforced when `capacity > 0`
- Duplicate registration rejected while user is still actively registered
- Re-registration is allowed after user cancels registration
- Only the creator can update, delete, and manage agendas
- Delete cascades: all agendas + registrations removed

---

## Notifications

> Notifications are persisted in the database and delivered in real time over WebSocket when the recipient is online.

### Endpoints

| Method | Endpoint                      | Auth | Description                    |
| ------ | ----------------------------- | ---- | ------------------------------ |
| GET    | `/api/notifications`          | ✅   | List my notifications          |
| GET    | `/api/notifications/unread`   | ✅   | Count unread notifications     |
| PATCH  | `/api/notifications/:id/read` | ✅   | Mark one notification as read  |
| PATCH  | `/api/notifications/read-all` | ✅   | Mark all notifications as read |

### GET `/api/notifications` — List My Notifications

**Query params:** `?page=1&limit=20`

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 12,
    "page": 1,
    "limit": 20,
    "notifications": [
      {
        "ID": 1,
        "UserID": 8,
        "NotificationType": "new_follower",
        "Title": "Pengikut baru",
        "Body": "Ani mulai mengikutimu",
        "ReferenceType": "follow",
        "ReferenceID": 5,
        "IsRead": false,
        "CreatedAt": "2026-04-04T10:30:00Z"
      }
    ]
  }
}
```

### GET `/api/notifications/unread` — Count Unread

**Response `200`:**

```json
{ "success": true, "data": { "unread_count": 3 } }
```

### PATCH `/api/notifications/:id/read` — Mark One as Read

**Response `200`:**

```json
{
  "success": true,
  "data": { "message": "notifikasi berhasil ditandai sudah dibaca" }
}
```

### PATCH `/api/notifications/read-all` — Mark All as Read

**Response `200`:**

```json
{
  "success": true,
  "data": { "message": "semua notifikasi ditandai sudah dibaca" }
}
```

### Notification Types

| Type                      | Source                        |
| ------------------------- | ----------------------------- |
| `new_follower`            | Follow system                 |
| `post_commented`          | Feed comment                  |
| `post_reacted`            | Feed reaction                 |
| `group_kicked`            | Group kick                    |
| `job_application_updated` | Job application status update |
| `mentor_request_received` | New mentor request            |
| `mentor_request_approved` | Mentor approved request       |
| `mentor_request_rejected` | Mentor rejected request       |
| `new_session`             | Mentor scheduled session      |
| `report_rejected`         | Admin rejected report         |
| `new_message`             | WebSocket message delivery    |
| `event_reminder`          | Event reminder scheduler      |

### Business Rules

- Notifications are created on relevant domain actions and persisted even if WebSocket delivery fails
- Unread counts and read flags are scoped to the authenticated user
- WebSocket-delivered notifications are also retrievable later through REST

---

## Jobs

> `alumni` and `partner` can post jobs. `alumni` and `student` can apply.
> Partners must already have an existing company profile before creating a job.

### Endpoints

| Method | Endpoint                            | Auth | Body        | Description                            |
| ------ | ----------------------------------- | ---- | ----------- | -------------------------------------- |
| GET    | `/api/jobs`                         | ✅   | —           | List jobs (paginated + filters)        |
| POST   | `/api/jobs`                         | ✅   | `form-data` | Create job posting                     |
| GET    | `/api/jobs/:id`                     | ✅   | —           | Job detail                             |
| PUT    | `/api/jobs/:id`                     | ✅   | `form-data` | Update own posting                     |
| DELETE | `/api/jobs/:id`                     | ✅   | —           | Delete own posting                     |
| POST   | `/api/jobs/:id/apply`               | ✅   | `form-data` | Apply for job                          |
| DELETE | `/api/jobs/:id/apply`               | ✅   | —           | Withdraw own application               |
| GET    | `/api/jobs/:id/applicants`          | ✅   | —           | View applicants (owner only)           |
| GET    | `/api/jobs/applications/mine`       | ✅   | —           | View my applications                   |
| PUT    | `/api/jobs/applications/:id/status` | ✅   | JSON        | Update application status (owner only) |

### GET `/api/jobs` — Query Parameters

| Param      | Description                                                         |
| ---------- | ------------------------------------------------------------------- |
| `search`   | Search by title or company name                                     |
| `job_type` | `full-time` · `part-time` · `internship` · `contract` · `freelance` |
| `status`   | `open` · `closed` · `filled`                                        |
| `page`     | Default: `1`                                                        |
| `limit`    | Default: `10`                                                       |

### POST `/api/jobs` — Create Job (form-data)

| Field          | Required | Notes                                                                |
| -------------- | -------- | -------------------------------------------------------------------- |
| `title`        | ✅       | —                                                                    |
| `company_name` | ✅       | Must match the poster's company profile when the poster is a partner |
| `openings`     | ❌       | Integer > 0; defaults to `1`                                         |
| `job_type`     | ✅       | `full-time` · `part-time` · `internship` · `contract` · `freelance`  |
| `description`  | ❌       | —                                                                    |
| `location`     | ❌       | —                                                                    |
| `salary_range` | ❌       | e.g. `"5.000.000 - 8.000.000"`                                       |
| `status`       | ❌       | `open` (default) · `closed`                                          |
| `image`        | ❌       | Image file → Cloudinary                                              |

### POST `/api/jobs/:id/apply` — Apply for Job (form-data)

| Field          | Required | Notes                       |
| -------------- | -------- | --------------------------- |
| `resume`       | ✅\*     | PDF file → Cloudinary       |
| `resume_url`   | ✅\*     | URL (if not uploading file) |
| `cover_letter` | ❌       | Optional text               |

> \*One of `resume` (file) or `resume_url` (link) is required.
> Applications are only accepted while the job status is `open`.

### DELETE `/api/jobs/:id/apply` — Withdraw Application

> Applicants can withdraw only while their application is still `pending`.

**Response `200`:**

```json
{
  "success": true,
  "data": { "message": "lamaran berhasil ditarik" }
}
```

### GET `/api/jobs/:id/applicants` — Resume URL Contract

- For Cloudinary-based resumes, `ResumeURL` is returned as a signed temporary download URL.
- Signed URL validity is currently 1 hour from response time.
- The URL may change on every request and should be opened immediately by frontend.
- Frontend should not persist this signed URL for long-term reuse.

**Example response snippet (`200`):**

```json
{
  "success": true,
  "data": [
    {
      "ID": 10,
      "JobID": 7,
      "UserID": 5,
      "Status": "pending",
      "ResumeURL": "https://api.cloudinary.com/v1_1/<cloud_name>/raw/download?public_id=alumni-platform/resumes/abc123&format=pdf&expires_at=1777000000&signature=..."
    }
  ]
}
```

### PUT `/api/jobs/applications/:id/status` — Update Status (JSON)

```json
{ "status": "accepted" }
```

> Status values: `pending` · `reviewed` · `accepted` · `rejected`
> `withdrawn` is set by the applicant withdrawal endpoint and appears in application history.
> Accepting an application decreases the job's remaining openings; when openings reach `0`, the job status becomes `filled`.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "ID": 3,
    "JobID": 7,
    "UserID": 5,
    "Status": "accepted",
    "ResumeURL": "https://res.cloudinary.com/.../resume.pdf",
    "CoverLetter": "I am interested in this role..."
  }
}
```

### Job Response Notes

- `CompanyID` is stored on the job when the company profile exists.
- `Openings` tracks remaining vacancies; default is `1`.
- `CompanyName` remains the public display label.
- `filled` jobs do not accept new applications.
- Withdrawn applications are kept as `withdrawn` and are not shown in the owner applicant list.

---

## Content Reporting

> `alumni`, `student`, and `admin` can submit reports.

| Method | Endpoint            | Auth | Description                |
| ------ | ------------------- | ---- | -------------------------- |
| POST   | `/api/reports`      | ✅   | Submit a report            |
| GET    | `/api/reports/mine` | ✅   | View own submitted reports |

### POST `/api/reports` — Submit Report (JSON)

```json
{
  "target_type": "post",
  "target_id": 12,
  "report_type": "spam",
  "description": "Only required when report_type is 'other'"
}
```

**`target_type` values:** `post` · `comment` · `group` · `group_article` · `event` · `job`

**`report_type` values:** `harassment` · `violence` · `hate_speech` · `spam` · `inappropriate` · `misinformation` · `copyright` · `other`

**Response `201`:**

```json
{
  "success": true,
  "data": {
    "ID": 8,
    "ReporterID": 2,
    "TargetType": "post",
    "TargetID": 12,
    "ReportType": "spam",
    "Status": "pending",
    "Description": null,
    "AdminNote": null
  }
}
```

### Business Rules

- Cannot submit a duplicate pending report on the same content
- `description` is **required** when `report_type` is `"other"`
- Report `Status`: `pending` → `resolved` | `rejected` (admin action)

---

## Admin Module

> `admin` role required for all endpoints (except `GET /api/categories`).

### Dashboard

#### GET `/api/admin/dashboard`

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "users": 109,
    "posts": 34,
    "groups": 5,
    "events": 12,
    "jobs": 8,
    "reports_pending": 3
  }
}
```

---

### User Management

| Method | Endpoint                      | Auth  | Description                                |
| ------ | ----------------------------- | ----- | ------------------------------------------ |
| GET    | `/api/admin/users`            | admin | List all users (paginated, filter by role) |
| GET    | `/api/admin/users/:id`        | admin | User detail                                |
| PATCH  | `/api/admin/users/:id/status` | admin | Activate or deactivate user                |
| PATCH  | `/api/admin/users/:id/role`   | admin | Change user role                           |

**GET `/api/admin/users` query params:** `?page=1&limit=20&role=alumni`

**PATCH `/api/admin/users/:id/status`:**

```json
{ "is_active": false }
```

**PATCH `/api/admin/users/:id/role`:**

```json
{ "role": "student" }
```

> Valid roles: `alumni` · `student` · `partner` · `admin`

**Response `200` (status update):**

```json
{
  "success": true,
  "data": {
    "ID": 5,
    "Name": "Siti Rahma",
    "Email": "siti@test.com",
    "Role": "alumni",
    "IsActive": false
  }
}
```

---

### Report Moderation

| Method | Endpoint                         | Auth  | Description                                |
| ------ | -------------------------------- | ----- | ------------------------------------------ |
| GET    | `/api/admin/reports`             | admin | List all reports (filter by status)        |
| GET    | `/api/admin/reports/:id`         | admin | Report detail                              |
| PATCH  | `/api/admin/reports/:id/resolve` | admin | Resolve report (optionally delete content) |
| PATCH  | `/api/admin/reports/:id/reject`  | admin | Reject report with reason                  |

**GET `/api/admin/reports` query params:** `?status=pending&page=1&limit=10`

**PATCH `/api/admin/reports/:id/resolve`:**

```json
{
  "admin_note": "Content violates community guidelines.",
  "delete_content": true
}
```

> `delete_content: true` → cascades-delete the reported post/group/event/job.

**PATCH `/api/admin/reports/:id/reject`:**

```json
{
  "admin_note": "Does not violate our policies."
}
```

**Response `200` (resolved):**

```json
{
  "success": true,
  "data": {
    "ID": 8,
    "Status": "resolved",
    "AdminNote": "Content violates community guidelines.",
    "ResolvedByID": 1,
    "ResolvedAt": "2026-04-04T08:00:00+07:00"
  }
}
```

---

### Direct Content Deletion

> Delete content without a report. Same cascading cleanup as owner-level deletes.

| Method | Endpoint                | Auth  | Description      |
| ------ | ----------------------- | ----- | ---------------- |
| DELETE | `/api/admin/posts/:id`  | admin | Delete any post  |
| DELETE | `/api/admin/groups/:id` | admin | Delete any group |
| DELETE | `/api/admin/events/:id` | admin | Delete any event |
| DELETE | `/api/admin/jobs/:id`   | admin | Delete any job   |

**Response `200`:**

```json
{ "success": true, "data": { "message": "postingan berhasil dihapus" } }
```

---

## Categories

| Method | Endpoint                    | Auth  | Description                                    |
| ------ | --------------------------- | ----- | ---------------------------------------------- |
| GET    | `/api/categories`           | ✅    | List all categories _(any authenticated user)_ |
| POST   | `/api/admin/categories`     | admin | Create category                                |
| PUT    | `/api/admin/categories/:id` | admin | Update category                                |
| DELETE | `/api/admin/categories/:id` | admin | Delete category                                |

**POST/PUT body (JSON):**

```json
{
  "name": "Technology",
  "description": "Tech-related posts and discussions"
}
```

**GET `/api/categories` response `200`:**

```json
{
  "success": true,
  "data": [
    { "ID": 1, "Name": "Technology", "Description": "Tech-related content" },
    { "ID": 2, "Name": "Career", "Description": null }
  ]
}
```

---

## Mentoring Module

> Connects **students** (mentees) with **alumni** (mentors) using a Content-Based Filtering recommendation engine (TF-IDF + Cosine Similarity implemented in pure Go).

### Role Matrix

| Action                      | alumni           | student            | partner | admin |
| --------------------------- | ---------------- | ------------------ | ------- | ----- |
| Register / manage as mentor | ✅               | ❌                 | ❌      | ❌    |
| Browse & request mentors    | ❌               | ✅                 | ❌      | ❌    |
| Get recommendations         | ❌               | ✅                 | ❌      | ❌    |
| Create / manage sessions    | ✅ (mentor side) | ✅ (create + view) | ❌      | ❌    |
| View own sessions           | ✅               | ✅                 | ❌      | ❌    |

---

### 🎓 Alumni (Mentor) Endpoints — `/api/mentor`

> All require `alumni` role.

#### POST `/api/mentor/register` — Register as Mentor

**Body (JSON):**

```json
{
  "mentor_bio": "5+ years in Python and ML. Happy to guide students.",
  "mentor_quota": 3
}
```

> `mentor_quota` allowed values: `1` · `2` · `3` · `5`

**Response `201`:** returns updated `UserProfile` with `MentorQuota` and `MentorDescription` set.

**Business Rules:**

- Only alumni accounts can register
- Must have created a profile first (`POST /api/profile`)
- Cannot register twice
- `mentor_bio` is required; quota must be `1`, `2`, `3`, or `5`

---

#### GET `/api/mentor/profile` — Get Own Mentor Profile

**Response `200`:** returns `UserProfile` with `User` and `Experiences` preloaded.

---

#### PUT `/api/mentor/profile` — Update Mentor Profile

```json
{
  "mentor_bio": "Updated specialization: Python, ML and Cloud.",
  "mentor_quota": 5
}
```

> New quota cannot be less than the current number of active mentees.

---

#### DELETE `/api/mentor/unregister` — Unregister as Mentor

> Blocked if the mentor still has approved (active) mentees.

**Response `200`:**

```json
{ "success": true, "data": { "message": "berhasil berhenti menjadi mentor" } }
```

---

#### GET `/api/mentor/requests` — View Incoming Requests

Returns all mentoring requests (pending + approved + rejected) sent to this mentor.

---

#### PATCH `/api/mentor/requests/:id/approve` — Approve Request

**Business Rules before approval:**

- Request must be `pending`
- Student must not already have 2 approved mentors
- Mentor must have remaining quota capacity

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "ID": 5,
    "MentorID": 2,
    "StudentID": 8,
    "Status": "approved",
    "SimilarityScore": 0.5263
  }
}
```

---

#### PATCH `/api/mentor/requests/:id/reject` — Reject Request

```json
{ "reason": "Not aligned with my current expertise." }
```

**Response `200`:** returns `MentorRequest` with `Status: "rejected"` and `RejectReason` set.

---

#### GET `/api/mentor/mentees` — List Active Mentees

Returns approved `MentorRequest` records (approved mentees only).

---

#### POST `/api/mentor/sessions` — Create Session

```json
{
  "student_id": 8,
  "topic": "Introduction to Python & ML basics",
  "notes": "Bring your laptop with VS Code installed",
  "session_date": "2026-07-01T10:00:00Z"
}
```

> `student_id` must have an **approved** mentoring request with this mentor. `session_date` uses ISO 8601.

**Response `201`:**

```json
{
  "success": true,
  "data": {
    "ID": 1,
    "RequestID": 5,
    "MentorID": 2,
    "StudentID": 8,
    "Topic": "Introduction to Python & ML basics",
    "Notes": "Bring your laptop with VS Code installed",
    "SessionDate": "2026-07-01T10:00:00Z",
    "Status": "scheduled"
  }
}
```

---

#### GET `/api/mentor/sessions` — List My Sessions (Mentor)

Returns all sessions where the caller is the mentor. Preloads `Student`.

---

#### PATCH `/api/mentor/sessions/:id` — Update Session

```json
{
  "topic": "Advanced Python: decorators and async",
  "status": "completed",
  "session_date": "2026-07-05T14:00:00Z"
}
```

> Valid `status` values: `scheduled` · `completed` · `cancelled`  
> Only the owning mentor can update (403 for others).

---

### 🎒 Student Endpoints — `/api/mentors` & `/api/student`

> All require `student` role.

#### GET `/api/mentors` — Browse Available Mentors

**Query params:** `?page=1&limit=10&search=Python`

Returns alumni who are registered as mentors **and still have capacity**.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 8,
    "page": 1,
    "limit": 10,
    "data": [
      {
        "ID": 3,
        "UserID": 2,
        "MentorDescription": "5+ years in Python and ML.",
        "MentorQuota": 3,
        "Skills": "Python, Machine Learning, Docker",
        "Interests": "AI, Data Science",
        "User": { "ID": 2, "Name": "Mentor Alumni", "Role": "alumni" }
      }
    ]
  }
}
```

---

#### GET `/api/mentors/:id` — Mentor Detail

Returns the mentor's full profile including `User` and `Experiences`. Returns `404` if user is not a registered mentor.

---

#### GET `/api/mentors/recommend` — Get Recommendations ⭐ NLP Engine

> Uses **TF-IDF + Cosine Similarity** (pure Go, no external ML dependency) to rank mentors by relevance to the student's profile or a custom query.

**Query params:**

| Param | Description                                                   | Default                                     |
| ----- | ------------------------------------------------------------- | ------------------------------------------- |
| `q`   | Custom free-text query (e.g. `python machine learning cloud`) | Student's `skills + interests` from profile |
| `top` | Number of results to return                                   | `10`                                        |

**How it works:**

1. Tokenizes each mentor's `skills + interests + mentor_bio + position + company + industry`
2. Tokenizes the student's query (or falls back to profile skills/interests)
3. Builds a TF-IDF corpus (mentors + student query)
4. Ranks mentors by cosine similarity to the student query vector

**NLP Pipeline:** case folding → special character removal → tokenization → Indonesian + English stopword removal → Indonesian suffix/prefix stemming

**Response `200`:**

```json
{
  "success": true,
  "data": [
    {
      "user_id": 2,
      "name": "Mentor Alumni",
      "profile_picture": "https://res.cloudinary.com/.../profile.jpg",
      "mentor_bio": "5+ years in Python and ML.",
      "skills": "Python, Machine Learning, Docker",
      "interests": "AI, Data Science",
      "position": "Senior Software Engineer",
      "company_name": "PT Tech Indonesia",
      "industry_name": "",
      "mentor_quota": 3,
      "similarity_score": 0.5263
    }
  ]
}
```

> Results are sorted by `similarity_score` descending. Score of `0` means no textual overlap.

---

#### POST `/api/mentors/:id/request` — Request Mentoring

```json
{
  "message": "Hi! I'd love to learn Python and ML from you.",
  "similarity_score": 0.5263
}
```

> `message` and `similarity_score` are optional. Pass `similarity_score` from the recommendation response to record it for analytics.

**Business Rules:**

- Cannot request yourself
- Cannot send a duplicate request (pending or approved to same mentor)
- Student may have **at most 2 approved mentors simultaneously**
- Mentor's quota must not be exceeded

**Response `201`:** returns `MentorRequest` with `Status: "pending"`.

---

#### GET `/api/student/mentors` — My Approved Mentors

Returns `MentorRequest` records where `Status = "approved"` for the current student.

---

#### GET `/api/student/requests` — My Sent Requests

Returns all `MentorRequest` records sent by the current student (all statuses).

---

#### POST `/api/student/sessions` — Create Session (Student)

```json
{
  "mentor_id": 2,
  "topic": "Need help with Python fundamentals",
  "notes": "Can we focus on OOP and API basics?",
  "session_date": "2026-07-03T10:00:00Z"
}
```

> `mentor_id` must have an **approved** mentoring request with the current student.

**Response `201`:** returns `MentoringSession` with `status = "scheduled"`.

---

#### GET `/api/student/sessions` — My Sessions (Student)

Returns all `MentoringSession` records where the caller is the student. Preloads `Mentor`.

---

### Business Rules Summary

| Rule                     | Detail                                                           |
| ------------------------ | ---------------------------------------------------------------- |
| Mentor role              | `alumni` only                                                    |
| Mentor quota             | Must be `1`, `2`, `3`, or `5`                                    |
| Mentee limit per student | Max **2 approved mentors** at a time                             |
| Capacity enforcement     | Request blocked when mentor's active mentees ≥ quota             |
| Duplicate request        | Only one `pending` or `approved` request per student-mentor pair |
| Session guard            | Session can only be created if an **approved** request exists    |
| Unregister guard         | Cannot unregister while active mentees exist                     |
| Status flow (Request)    | `pending` → `approved` \| `rejected`                             |
| Status flow (Session)    | `scheduled` → `completed` \| `cancelled`                         |

---

## Message Module

> Enables private real-time messaging between users with a prerequisite follow system.
> Only `student` and `alumni` can follow or send messages. One follow in either direction unlocks messaging for both users.

### How It Works

```
User A follows User B
  → Both A→B and B→A messaging is now permitted (symmetric)

WebSocket (real-time)  +  REST (history / unread count)
  ↓                           ↓
messages persisted to DB first, then delivered live if recipient is connected
```

---

### Follow System — `/api/users`

> Requires `student` or `alumni` role. Partners and admins are blocked.

| Method | Endpoint                   | Auth | Description                   |
| ------ | -------------------------- | ---- | ----------------------------- |
| POST   | `/api/users/:id/follow`    | ✅   | Follow a user                 |
| DELETE | `/api/users/:id/follow`    | ✅   | Unfollow a user               |
| GET    | `/api/users/:id/followers` | ✅   | List users who follow `:id`   |
| GET    | `/api/users/:id/following` | ✅   | List users that `:id` follows |

#### POST `/api/users/:id/follow` — Follow User

**Response `201`:**

```json
{ "success": true, "data": { "message": "berhasil mengikuti pengguna" } }
```

**Error cases:**
| Status | Reason |
|---|---|
| `400` | Self-follow or unfollowing someone not followed |
| `401` | Not authenticated |
| `403` | Role is not `student` or `alumni` |
| `409` | Already following this user |

#### DELETE `/api/users/:id/follow` — Unfollow User

**Response `200`:**

```json
{
  "success": true,
  "data": { "message": "berhasil berhenti mengikuti pengguna" }
}
```

#### GET `/api/users/:id/followers` — List Followers

**Response `200`:** returns array of `User` objects who follow `:id`.

#### GET `/api/users/:id/following` — List Following

**Response `200`:** returns array of `User` objects that `:id` follows.

---

### REST Messaging — `/api/messages`

> Requires `student` or `alumni` role.

| Method | Endpoint                     | Auth | Description                                                     |
| ------ | ---------------------------- | ---- | --------------------------------------------------------------- |
| GET    | `/api/messages`              | ✅   | List all conversations (last message per partner, unread count) |
| GET    | `/api/messages/unread`       | ✅   | Total unread message count                                      |
| GET    | `/api/messages/:userID`      | ✅   | Conversation history with a specific user (paginated)           |
| PATCH  | `/api/messages/:userID/read` | ✅   | Mark all messages from `:userID` as read                        |

#### GET `/api/messages` — Conversation List

**Response `200`:**

```json
{
  "success": true,
  "data": [
    {
      "partner_id": 8,
      "partner_name": "Student Budi",
      "last_message": "Halo kak, ada waktu untuk diskusi?",
      "last_message_at": "2026-04-04T10:30:00Z",
      "unread_count": 3
    }
  ]
}
```

#### GET `/api/messages/unread` — Total Unread Count

**Response `200`:**

```json
{ "success": true, "data": { "unread_count": 5 } }
```

#### GET `/api/messages/:userID` — Conversation History

**Query params:** `?page=1&limit=20`

Returns paginated messages newest-first between the caller and `:userID`.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 42,
    "page": 1,
    "limit": 20,
    "messages": [
      {
        "ID": 15,
        "SenderID": 8,
        "ReceiverID": 2,
        "Content": "Halo kak!",
        "IsRead": true,
        "CreatedAt": "2026-04-04T10:30:00Z"
      }
    ]
  }
}
```

#### PATCH `/api/messages/:userID/read` — Mark as Read

Marks all messages **from `:userID` to the caller** as read.

**Response `200`:**

```json
{
  "success": true,
  "data": { "message": "pesan berhasil ditandai sudah dibaca" }
}
```

---

### WebSocket — Real-time Messaging

#### `GET /api/ws` — Connect

> Auth via query param (not header) because WebSocket upgrade is a GET request and browsers cannot set custom headers for it.

```
ws://localhost:8080/api/ws?token=<jwt>
```

**Connection flow:**

1. Server validates JWT from `?token=` **before** the HTTP → WebSocket upgrade
   - Missing or invalid token → `401 Unauthorized` HTTP response _(upgrade never attempted)_
   - Role not `student` or `alumni` → `403 Forbidden` HTTP response _(upgrade never attempted)_
2. Server upgrades the connection (`101 Switching Protocols`)
3. Registers client in the Hub (one connection per user; existing connection is replaced)
4. Keepalive starts: server sends a **Ping** frame every **30 s**; client must reply with a **Pong** within **60 s** or the connection is closed
5. Starts write pump (Hub → client) and read pump (client → Hub → DB → recipient)

#### Client → Server (send message)

```json
{
  "receiver_id": 8,
  "content": "Halo! Ada waktu untuk diskusi Python?"
}
```

#### Server → Client (new message received)

```json
{
  "type": "message",
  "data": {
    "ID": 15,
    "SenderID": 2,
    "ReceiverID": 8,
    "Content": "Halo! Ada waktu untuk diskusi Python?",
    "IsRead": false,
    "CreatedAt": "2026-04-04T10:30:00Z"
  }
}
```

> The sender also receives an echo of their own sent message as confirmation.

#### Server → Client (error)

```json
{
  "type": "error",
  "message": "anda harus mengikuti pengguna ini sebelum mengirim pesan"
}
```

**Error cases:**

| Cause                              | How it surfaces                                                                                         |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Missing token                      | `401 Unauthorized` HTTP response — before upgrade                                                       |
| Invalid or expired token           | `401 Unauthorized` HTTP response — before upgrade                                                       |
| Role not `student` or `alumni`     | `403 Forbidden` HTTP response — before upgrade                                                          |
| Not following recipient            | `{"type":"error","message":"anda harus mengikuti pengguna ini sebelum mengirim pesan"}` WebSocket frame |
| Missing `receiver_id` or `content` | `{"type":"error","message":"receiver_id dan content wajib diisi"}` WebSocket frame                      |
| Invalid JSON payload               | `{"type":"error","message":"format pesan tidak valid"}` WebSocket frame                                 |

---

#### Frontend Integration Example (JavaScript / React)

> The examples below assume a **Vite + React** project. All three files work together:
> `lib/chatSocket.js` → `hooks/useChat.js` → `components/ChatWindow.jsx`.
>
> **Ping / Pong is fully transparent** — the browser WebSocket API handles Pong replies automatically. Your JavaScript code never needs to deal with it.

---

##### `lib/chatSocket.js` — Standalone WebSocket class

```javascript
// lib/chatSocket.js
//
// Wraps the native WebSocket API with:
//   • JWT auth via ?token= query param (no custom headers needed)
//   • JSON message framing matching the server envelope
//   • Automatic reconnect with exponential back-off (up to MAX_RETRIES)

const WS_BASE_URL = import.meta.env.VITE_WS_URL ?? "ws://localhost:8080";

const MAX_RETRIES = 5;
const BASE_DELAY_MS = 1_000; // first reconnect delay; doubles each attempt

export class ChatSocket {
  #ws = null;
  #token;
  #retries = 0;
  #reconnectTimer = null;
  #intentionallyClosed = false;

  // ── Callbacks (assign before calling connect()) ───────────────────────────
  /** Called with the persisted Message object when a new message arrives. */
  onMessage = null; // (msg)   => void
  /** Called with the persisted Notification object when a notification arrives. */
  onNotification = null; // (notif) => void
  /** Called with the server error string when an error frame arrives. */
  onError = null; // (text)  => void
  /** Called whenever the connection status changes. */
  onStatusChange = null; // ('connecting'|'open'|'closed'|'error') => void

  constructor(token) {
    this.#token = token;
  }

  // ── Public API ────────────────────────────────────────────────────────────

  connect() {
    if (this.#ws?.readyState === WebSocket.OPEN) return;

    this.#intentionallyClosed = false;
    this.#notify("connecting");

    const url = `${WS_BASE_URL}/api/ws?token=${encodeURIComponent(this.#token)}`;
    this.#ws = new WebSocket(url);

    this.#ws.onopen = () => {
      this.#retries = 0;
      this.#notify("open");
    };

    this.#ws.onmessage = ({ data }) => {
      let envelope;
      try {
        envelope = JSON.parse(data);
      } catch {
        console.error("[ChatSocket] Unparseable frame:", data);
        return;
      }

      switch (envelope.type) {
        case "message":
          this.onMessage?.(envelope.data);
          break;
        case "notification":
          this.onNotification?.(envelope.data);
          break;
        case "error":
          this.onError?.(envelope.message);
          break;
        default:
          console.warn("[ChatSocket] Unknown envelope type:", envelope.type);
      }
    };

    // onerror always fires before onclose — notify status but let onclose drive reconnect.
    this.#ws.onerror = () => this.#notify("error");

    this.#ws.onclose = ({ code, reason, wasClean }) => {
      console.log(
        `[ChatSocket] Closed — code: ${code}, clean: ${wasClean}, reason: "${reason}"`,
      );
      this.#notify("closed");
      if (!this.#intentionallyClosed) this.#scheduleReconnect();
    };
  }

  /**
   * Send a chat message to a recipient.
   * @param {number} receiverId  Target user's ID.
   * @param {string} content     Message text (non-empty).
   */
  send(receiverId, content) {
    if (this.#ws?.readyState !== WebSocket.OPEN) {
      console.warn("[ChatSocket] Cannot send — socket is not open");
      return;
    }
    this.#ws.send(JSON.stringify({ receiver_id: receiverId, content }));
  }

  /** Permanently close the connection (no automatic reconnect). */
  disconnect() {
    this.#intentionallyClosed = true;
    clearTimeout(this.#reconnectTimer);
    this.#ws?.close();
    this.#ws = null;
  }

  get isConnected() {
    return this.#ws?.readyState === WebSocket.OPEN;
  }

  // ── Private helpers ───────────────────────────────────────────────────────

  #scheduleReconnect() {
    if (this.#retries >= MAX_RETRIES) {
      console.error("[ChatSocket] Max reconnect attempts reached — giving up");
      return;
    }
    const delay = BASE_DELAY_MS * 2 ** this.#retries;
    this.#retries++;
    console.log(
      `[ChatSocket] Reconnecting in ${delay} ms (attempt ${this.#retries}/${MAX_RETRIES})`,
    );
    this.#reconnectTimer = setTimeout(() => this.connect(), delay);
  }

  #notify(status) {
    this.onStatusChange?.(status);
  }
}
```

---

##### `hooks/useChat.js` — React hook

```javascript
// hooks/useChat.js
//
// Manages a single shared WebSocket connection for the logged-in user.
// Pass `token = null` (e.g. when the user is not authenticated) to skip connecting.

import { useCallback, useEffect, useRef, useState } from "react";
import { ChatSocket } from "../lib/chatSocket";

/**
 * @param {string|null} token  JWT obtained after login.
 * @returns {{
 *   messages:          Array,
 *   send:              (receiverId: number, content: string) => void,
 *   isConnected:       boolean,
 *   connectionStatus:  string,
 * }}
 */
export function useChat(token) {
  const socketRef = useRef(null);
  const [messages, setMessages] = useState([]);
  const [connectionStatus, setStatus] = useState("closed");

  useEffect(() => {
    if (!token) return;

    const socket = new ChatSocket(token);
    socketRef.current = socket;

    // Append every incoming message to local state.
    // Filter by partner in the component that renders the conversation.
    socket.onMessage = (msg) => {
      setMessages((prev) => {
        // Avoid duplicate echo frames (sender receives its own message back)
        const exists = prev.some((m) => m.ID === msg.ID);
        return exists ? prev : [...prev, msg];
      });
    };

    // Route notifications to your global notification store / toast system.
    socket.onNotification = (notif) => {
      console.log("[useChat] notification:", notif);
      // e.g. notificationStore.add(notif);
    };

    socket.onError = (errMsg) => {
      console.error("[useChat] server error:", errMsg);
    };

    socket.onStatusChange = setStatus;

    socket.connect();

    return () => {
      socket.disconnect();
      socketRef.current = null;
    };
  }, [token]); // reconnects automatically if the token changes (e.g. after refresh)

  const send = useCallback((receiverId, content) => {
    socketRef.current?.send(receiverId, content);
  }, []);

  return {
    messages,
    send,
    isConnected: connectionStatus === "open",
    connectionStatus,
  };
}
```

---

##### `components/ChatWindow.jsx` — Usage example

```jsx
// components/ChatWindow.jsx
//
// Renders a conversation with a single partner and lets the user send messages.
// Assumes useAuthStore exposes the raw JWT string as `token`.

import { useState } from "react";
import { useChat } from "../hooks/useChat";
import { useAuthStore } from "../stores/authStore"; // adjust to your auth solution

/**
 * @param {{ partnerId: number }} props
 */
export function ChatWindow({ partnerId }) {
  const token = useAuthStore((s) => s.token);
  const { messages, send, isConnected, connectionStatus } = useChat(token);
  const [draft, setDraft] = useState("");

  // Show only messages that belong to this conversation
  const conversation = messages.filter(
    (m) => m.SenderID === partnerId || m.ReceiverID === partnerId,
  );

  const handleSend = () => {
    const text = draft.trim();
    if (!text || !isConnected) return;
    send(partnerId, text);
    setDraft("");
  };

  return (
    <div className="chat-window">
      {/* ── Connection status badge ── */}
      <div className={`status status--${connectionStatus}`}>
        {connectionStatus === "open" && "🟢 Connected"}
        {connectionStatus === "connecting" && "🟡 Connecting…"}
        {connectionStatus === "closed" && "🔴 Disconnected"}
        {connectionStatus === "error" && "🔴 Connection error"}
      </div>

      {/* ── Message list ── */}
      <ul className="message-list">
        {conversation.map((m) => (
          <li
            key={m.ID}
            className={
              m.SenderID === partnerId
                ? "message--incoming"
                : "message--outgoing"
            }
          >
            <span className="message__content">{m.Content}</span>
            <span className="message__time">
              {new Date(m.CreatedAt).toLocaleTimeString()}
            </span>
          </li>
        ))}
      </ul>

      {/* ── Composer ── */}
      <div className="composer">
        <input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && !e.shiftKey && handleSend()}
          placeholder={isConnected ? "Ketik pesan…" : "Menghubungkan…"}
          disabled={!isConnected}
        />
        <button onClick={handleSend} disabled={!isConnected || !draft.trim()}>
          Kirim
        </button>
      </div>
    </div>
  );
}
```

---

##### Environment variable

Add to your `.env` (Vite project):

```env
# Development
VITE_WS_URL=ws://localhost:8080

# Production
VITE_WS_URL=wss://api.yourapp.com
```

> Use `wss://` (WebSocket Secure) in production — equivalent to HTTPS for WebSocket connections.

---

##### Key behaviours to know

| Behaviour                | Detail                                                                                                                                                                                |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Auth failure (401 / 403) | `onclose` fires immediately with `wasClean: false`; the class schedules a reconnect — check `onError` in your auth store and call `socket.disconnect()` instead if it's an auth error |
| Duplicate echo frame     | The server echoes every sent message back to the sender as delivery confirmation; `useChat` deduplicates by `msg.ID`                                                                  |
| Ping / Pong              | Handled silently by the browser — your JS code never sees raw Ping or Pong frames                                                                                                     |
| Reconnect back-off       | 1 s → 2 s → 4 s → 8 s → 16 s, then gives up; reset on successful open                                                                                                                 |
| Token refresh            | If your auth library refreshes the JWT, pass the new token to `useChat` — the effect re-runs and opens a fresh connection                                                             |
| Missed messages          | If the user was offline, fetch history via `GET /api/messages/:userID` on reconnect                                                                                                   |

---

### Business Rules Summary

| Rule                       | Detail                                                                      |
| -------------------------- | --------------------------------------------------------------------------- |
| Role restriction           | Only `student` and `alumni` can follow or message                           |
| Follow prerequisite        | **At least one** follow must exist between two users (either direction)     |
| Symmetric messaging        | A follows B → both A→B and B→A messages are permitted                       |
| Self-follow / self-message | Blocked                                                                     |
| Duplicate follow           | Returns `409 Conflict`                                                      |
| Message immutability       | Messages **cannot be edited or deleted**                                    |
| Privacy                    | Users can only view their own conversation history                          |
| Offline delivery           | Messages persisted to DB; fetched via REST when recipient comes back online |

---

## Cloudinary Upload Reference

Every file field is **optional** — omit to skip upload. Accepted formats: `jpg` · `jpeg` · `png` · `webp`

| Module           | Form field                           | Cloudinary folder                 | Endpoint                            |
| ---------------- | ------------------------------------ | --------------------------------- | ----------------------------------- |
| Profile picture  | `picture`                            | `alumni-platform/profiles`        | POST/PUT `/api/profile`             |
| Feed post image  | `image`                              | `alumni-platform/feed`            | POST/PUT `/api/feed`                |
| Group banner     | `banner`                             | `alumni-platform/groups/banners`  | POST/PUT `/api/groups`              |
| Group article    | `medias` (array) or `media` (single) | `alumni-platform/groups/articles` | POST/PUT `/api/groups/articles/:id` |
| Portfolio item   | `media`                              | `alumni-platform/portfolio`       | POST/PUT `/api/portfolio`           |
| Event photo      | `photo`                              | `alumni-platform/events`          | POST/PUT `/api/events`              |
| Job image        | `image`                              | `alumni-platform/jobs`            | POST/PUT `/api/jobs`                |
| Job resume (PDF) | `resume`                             | `alumni-platform/resumes`         | POST `/api/jobs/:id/apply`          |

---

## HTTP Status Codes

| Code | Meaning                                                         |
| ---- | --------------------------------------------------------------- |
| 200  | Success (read, update, delete)                                  |
| 201  | Created (new resource created)                                  |
| 400  | Bad request (validation failed, duplicate, business rule error) |
| 401  | Unauthorized (missing or invalid token)                         |
| 403  | Forbidden (authenticated but insufficient role/ownership)       |
| 404  | Not found                                                       |
| 500  | Internal server error                                           |

---

## CORS

CORS is enabled globally. Configure allowed origins via `.env`:

```env
# Development (allow all)
CORS_ALLOWED_ORIGINS=*

# Production (restrict to your frontend domain)
CORS_ALLOWED_ORIGINS=https://yourapp.com,https://www.yourapp.com
```

Allowed methods: `GET, POST, PUT, PATCH, DELETE, OPTIONS`  
Allowed headers: `Origin, Content-Type, Accept, Authorization, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Extensions, Connection, Upgrade`
