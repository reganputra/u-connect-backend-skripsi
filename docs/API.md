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
- [Categories](#categories)
- [Mentoring Module](#mentoring-module)
- [Message Module](#message-module)
- [Follow Module](#follow-module)
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

### POST `/api/auth/change-password` — Change Password

**Auth required.**

Allows a logged-in user to change their account password. Requires the current (old) password for verification before setting a new one.

**Body (JSON):**

```json
{
  "old_password": "OldPassword123!",
  "new_password": "NewPassword456!",
  "confirm_password": "NewPassword456!"
}
```

| Field              | Required | Notes                                     |
| ------------------ | -------- | ----------------------------------------- |
| `old_password`     | ✅       | Must match the account's current password |
| `new_password`     | ✅       | Minimum 8 characters                      |
| `confirm_password` | ✅       | Must exactly match `new_password`         |

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "message": "password berhasil diperbarui"
  }
}
```

**Error cases:**

| Status | Reason                                                       |
| ------ | ------------------------------------------------------------ |
| `400`  | Any field is empty                                           |
| `400`  | `new_password` is less than 8 characters                     |
| `400`  | `new_password` and `confirm_password` do not match           |
| `400`  | `old_password` does not match the account's current password |
| `401`  | Not authenticated                                            |

### POST `/api/auth/forgot-password` — Forgot Password (Rate Limited)

**No auth required.**

Allows a user to reset their password without knowing the old password, verified using their registration data (`faculty`, `major`, `year_enroll`).
If the data does not match, it counts as a failed attempt. The account is locked after 3 failed attempts, requiring an admin to unlock it.

**Body (JSON):**

```json
{
  "email": "regan@test.com",
  "faculty": "Engineering",
  "major": "Informatics",
  "year_enroll": 2020,
  "new_password": "NewPassword123!",
  "confirm_password": "NewPassword123!"
}
```

| Field              | Required | Notes                             |
| ------------------ | -------- | --------------------------------- |
| `email`            | ✅       | Case-insensitive                  |
| `faculty`          | ✅       | Case-insensitive                  |
| `major`            | ✅       | Case-insensitive                  |
| `year_enroll`      | ✅       | Must match registration data      |
| `new_password`     | ✅       | Minimum 8 characters              |
| `confirm_password` | ✅       | Must exactly match `new_password` |

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "message": "password berhasil direset, silakan login dengan password baru"
  }
}
```

**Error cases:**

| Status | Reason                                                                                                              |
| ------ | ------------------------------------------------------------------------------------------------------------------- |
| `400`  | Any field is missing                                                                                                |
| `400`  | `new_password` is less than 8 characters                                                                            |
| `400`  | `new_password` and `confirm_password` do not match                                                                  |
| `400`  | Registration data does not match (returns generic error message to prevent enumeration, increments attempt counter) |
| `403`  | Account is locked due to too many failed reset attempts                                                             |

### GET `/api/me/activity/summary` — My Activity Summary

**Auth required.**

Returns lightweight counters used for dashboard/activity cards.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "events_owned": 4,
    "events_registered": 11,
    "jobs_owned": 3,
    "jobs_applied": 7,
    "groups_owned": 2,
    "groups_joined": 9
  }
}
```

Notes:

- `events_owned`: events created by current user
- `events_registered`: events where current user is registered
- `jobs_owned`: jobs posted by current user
- `jobs_applied`: job applications submitted by current user
- `groups_owned`: groups created by current user
- `groups_joined`: groups where current user is an active member

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
| `name`                     | ❌                          | Update display name (also updates `users.name`)       |
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
        "user": {
          "ID": 2,
          "Name": "Regan Putra",
          "Role": "alumni",
          "picture_url": "https://res.cloudinary.com/.../profiles/regan.jpg"
        },
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
    "user": {
      "ID": 2,
      "Name": "Regan Putra",
      "Role": "alumni",
      "picture_url": "https://res.cloudinary.com/.../profiles/regan.jpg"
    },
    "reactions": [{ "ID": 1, "UserID": 2, "Type": "like" }],
    "votes": [{ "ID": 1, "UserID": 3, "Value": 1 }],
    "comments": [
      {
        "id": 1,
        "content": "Great post!",
        "user": {
          "ID": 3,
          "Name": "Ani",
          "picture_url": "https://res.cloudinary.com/.../profiles/ani.jpg"
        },
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
> User objects in feed responses now also include `picture_url` when available.

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
| GET    | `/api/groups/owned`               | ✅   | —           | Groups owned by current user (student/alumni only)   |
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
- `GET /api/groups/owned?page=1&limit=20`
- `GET /api/groups/:id/members?page=1&limit=20`

Each paginated response includes `total`, `page`, and `limit`.

- `GET /api/groups` response shape: `{ total, page, limit, data }`
- `GET /api/groups/joined` response shape: `{ total, page, limit, data }`
- `GET /api/groups/owned` response shape: `{ total, page, limit, data }`
- `GET /api/groups/:id/members` response shape: `{ total, page, limit, members }`

User payloads inside group responses now include avatar-ready fields:

- `Owner.picture_url`
- `Members[].User.picture_url`
- `Articles[].User.picture_url`
- `Articles[].Comments[].User.picture_url`
- `GET /api/groups/articles/:id` comment tree user avatars are included in each nested `user`

Frontend can use the returned user `ID` to navigate to the profile page route, for example `/directory/:id`.

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
  - `Owner.picture_url`, `Members[].User.picture_url`, and `Articles[].User.picture_url`
- `GET /api/groups/articles/:id` includes:
  - `comment_count` (total visible comments in the response)
  - `media_urls` (array of all image URLs for the article)
  - `user.picture_url` for the article author
  - nested `comments[].user.picture_url` for visible comment threads

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
| GET    | `/api/events/mine/owned`       | ✅   | —           | List events created by current user      |
| GET    | `/api/events/mine/registered`  | ✅   | —           | List events registered by current user   |
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

### GET `/api/events/mine/owned` — My Owned Events

**Query params:** `?page=1&limit=10`

Returns paginated events where `UserID` equals the authenticated user.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 4,
    "page": 1,
    "limit": 10,
    "events": [
      {
        "ID": 5,
        "Title": "Alumni Seminar 2026",
        "UserID": 2,
        "Status": "upcoming",
        "AttendantCount": 37,
        "SeatLeft": 13
      }
    ]
  }
}
```

### GET `/api/events/mine/registered` — My Registered Events

**Query params:** `?page=1&limit=10`

Returns paginated events where the authenticated user has an active registration.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 11,
    "page": 1,
    "limit": 10,
    "events": [
      {
        "ID": 9,
        "Title": "Open Networking Night",
        "Status": "ongoing",
        "AttendantCount": 120,
        "SeatLeft": null
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
    "Registrations": [
      {
        "ID": 21,
        "EventID": 5,
        "UserID": 8,
        "ReminderSent": false,
        "User": {
          "ID": 8,
          "Name": "Student Budi",
          "Role": "student",
          "picture_url": "https://res.cloudinary.com/.../profiles/budi.jpg"
        }
      }
    ]
  }
}
```

> `Registrations[].User` now includes `picture_url` when a profile picture exists.

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

> Notifications are persisted in the database first, then best-effort delivered over WebSocket when the recipient is online.
> Missed notifications can always be fetched later through REST.

### Endpoints

| Method | Endpoint                      | Auth | Description                    |
| ------ | ----------------------------- | ---- | ------------------------------ |
| GET    | `/api/notifications`          | ✅   | List my notifications          |
| GET    | `/api/notifications/unread`   | ✅   | Count unread notifications     |
| PATCH  | `/api/notifications/:id/read` | ✅   | Mark one notification as read  |
| PATCH  | `/api/notifications/read-all` | ✅   | Mark all notifications as read |

### GET `/api/notifications` — List My Notifications

**Query params:** `?page=1&limit=20`

- `page` defaults to `1` when missing or invalid
- `limit` defaults to `20` when missing, invalid, or greater than `100`

Returned notifications are sorted by newest first.

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
        "RedirectURL": "/directory/5",
        "IsRead": false,
        "CreatedAt": "2026-04-04T10:30:00Z"
      }
    ]
  }
}
```

**Notification fields:**

- `ID`: notification record ID
- `UserID`: recipient user ID
- `NotificationType`: event key such as `new_follower` or `report_rejected`
- `Title`: short headline shown in UI
- `Body`: detail text
- `ReferenceType`: related resource type such as `follow`, `post`, `mentor_request`, `report`, or `event`
- `ReferenceID`: related resource ID
- `RedirectURL`: frontend route hint generated by backend for click navigation
- `IsRead`: read state scoped to the authenticated user
- `CreatedAt`: creation timestamp

### GET `/api/notifications/unread` — Count Unread

**Response `200`:**

```json
{ "success": true, "data": { "unread_count": 3 } }
```

### PATCH `/api/notifications/:id/read` — Mark One as Read

**Path params:** `id` = notification ID

**Response `200`:**

```json
{
  "success": true,
  "data": { "message": "notifikasi berhasil ditandai sudah dibaca" }
}
```

**Error cases:**

| Status | Reason                                                             |
| ------ | ------------------------------------------------------------------ |
| `400`  | Invalid notification ID                                            |
| `401`  | Not authenticated                                                  |
| `404`  | Notification does not belong to the current user or does not exist |

### PATCH `/api/notifications/read-all` — Mark All as Read

**Response `200`:**

```json
{
  "success": true,
  "data": { "message": "semua notifikasi ditandai sudah dibaca" }
}
```

### Real-Time Delivery

Connected clients receive notification events through the WebSocket channel at `/api/ws?token=<jwt>`.

Server envelope:

```json
{
  "type": "notification",
  "data": {
    "ID": 1,
    "UserID": 8,
    "NotificationType": "new_follower",
    "Title": "Pengikut baru",
    "Body": "Ani mulai mengikutimu",
    "ReferenceType": "follow",
    "ReferenceID": 5,
    "RedirectURL": "/directory/5",
    "IsRead": false,
    "CreatedAt": "2026-04-04T10:30:00Z"
  }
}
```

### Redirect Mapping

Notification payloads now include `RedirectURL` so frontend can navigate directly when a notification is clicked.

| ReferenceType     | RedirectURL format                    |
| ----------------- | ------------------------------------- |
| `post`            | `/feed/:id`                           |
| `comment`         | `/feed/:postId`                       |
| `group`           | `/groups/:id`                         |
| `group_article`   | `/groups/:groupId/article/:articleId` |
| `group_comment`   | `/groups/:groupId/article/:articleId` |
| `event`           | `/events/:id`                         |
| `job`             | `/jobs/:id`                           |
| `job_application` | `/jobs/applications/mine`             |
| `follow`          | `/directory/:id`                      |
| `mentor_request`  | `/student/requests`                   |
| `report`          | `/reports/mine`                       |
| `message`         | `/messages/:partnerUserID`            |

Delivery rules:

- The notification is always inserted into the database first
- WebSocket delivery is best-effort and does not block the request
- If the recipient is offline, the event is not delivered live but remains available via REST
- Marking as read is always done through the REST API, not through WebSocket ack
- For `message`, backend resolves `partnerUserID` from the message sender/receiver pair for the notification recipient
- For `report_rejected`, backend now sets `ReferenceType` and `ReferenceID` to the reported target content (for example `post` + post ID), so the notification click opens that content directly

### Notification Types

| Type                        | Source                                                                                                |
| --------------------------- | ----------------------------------------------------------------------------------------------------- |
| `new_follower`              | Follow system                                                                                         |
| `post_commented`            | Feed comment                                                                                          |
| `post_replied`              | Feed comment reply                                                                                    |
| `post_reacted`              | Feed reaction                                                                                         |
| `group_article_commented`   | Group article comment                                                                                 |
| `group_article_replied`     | Group article comment reply                                                                           |
| `group_kicked`              | Group kick                                                                                            |
| `content_deleted_by_admin`  | Admin removed user content (includes content title/name and deletion reason)                          |
| `job_application_updated`   | Job application status update                                                                         |
| `mentor_request_received`   | New mentor request                                                                                    |
| `mentor_request_approved`   | Mentor approved request                                                                               |
| `mentor_request_rejected`   | Mentor rejected request                                                                               |
| `mentor_relationship_ended` | Mentor ended active mentorship                                                                        |
| `new_session`               | Mentor scheduled session                                                                              |
| `report_rejected`           | Admin rejected report (includes reported content title/name and clickable target)                     |
| `report_resolved_deleted`   | Admin accepted report and deleted reported content (includes content title/name and clickable target) |
| `new_message`               | WebSocket message delivery                                                                            |
| `event_reminder`            | Event reminder scheduler                                                                              |

### Business Rules

- Notifications are created on relevant domain actions and persisted even if WebSocket delivery fails
- Unread counts and read flags are scoped to the authenticated user
- WebSocket-delivered notifications are also retrievable later through REST
- When admin resolves a report with `delete_content: true` and deletion succeeds, reporter receives `report_resolved_deleted`

---

## Jobs

> `alumni` and `partner` can post jobs. `alumni` and `student` can apply.
> Partners must already have an existing company profile before creating a job.

### Endpoints

| Method | Endpoint                            | Auth | Body        | Description                            |
| ------ | ----------------------------------- | ---- | ----------- | -------------------------------------- |
| GET    | `/api/jobs`                         | ✅   | —           | List jobs (paginated + filters)        |
| GET    | `/api/jobs/mine/owned`              | ✅   | —           | View jobs posted by current user       |
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

### GET `/api/jobs/mine/owned` — My Owned Jobs

**Query params:** `?page=1&limit=10`

Returns paginated jobs where `UserID` equals the authenticated user.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 3,
    "page": 1,
    "limit": 10,
    "jobs": [
      {
        "ID": 14,
        "UserID": 2,
        "Title": "Backend Engineer",
        "CompanyName": "PT Tech Indonesia",
        "Status": "open",
        "Openings": 2
      }
    ]
  }
}
```

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
      "User": {
        "ID": 5,
        "Name": "Siti Rahma",
        "picture_url": "https://res.cloudinary.com/.../profiles/siti.jpg"
      },
      "Status": "pending",
      "ResumeURL": "https://api.cloudinary.com/v1_1/<cloud_name>/raw/download?public_id=alumni-platform/resumes/abc123&format=pdf&expires_at=1777000000&signature=..."
    }
  ]
}
```

> Applicant rows include nested `User` info with `Name` and `picture_url` for avatar rendering.

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

Reporting is a user-submitted moderation queue. Reports start as `pending`, can only be processed once, and are visible to the reporter through `GET /api/reports/mine`.

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
  "description": "Optional unless report_type is 'other'"
}
```

**`target_type` values:** `post` · `comment` · `group` · `group_article` · `event` · `job`

**`report_type` values:** `harassment` · `violence` · `hate_speech` · `spam` · `inappropriate` · `misinformation` · `copyright` · `other`

**Business rules:**

- `description` is required when `report_type` is `other`
- A user cannot submit more than one pending report for the same target
- Report status flow is `pending` → `resolved` or `rejected`
- Only admin users can process reports in the admin module

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

### GET `/api/reports/mine` — View My Reports

**Query params:** `?page=1&limit=10`

Returns the authenticated user's own submitted reports in newest-first order.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 2,
    "page": 1,
    "limit": 10,
    "reports": [
      {
        "ID": 8,
        "ReporterID": 2,
        "TargetType": "post",
        "TargetID": 12,
        "ReportType": "spam",
        "Status": "pending",
        "Description": null,
        "AdminNote": null
      }
    ]
  }
}
```

### Business Rules

- Cannot submit a duplicate pending report on the same content
- `description` is required when `report_type` is `other`
- Report `Status`: `pending` → `resolved` | `rejected`
- Rejected reports create a `report_rejected` notification for the reporter

---

## Admin Module

> `admin` role required for all endpoints in this section, except `GET /api/categories`, which is readable by any authenticated user.

The admin module covers dashboard statistics, user moderation, report moderation, direct content deletion, and category management.

### Dashboard

#### GET `/api/admin/dashboard`

Returns summary counts for core platform objects and pending moderation workload.

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

Admins can inspect users, toggle account status, and change roles.

| Method | Endpoint                            | Auth  | Description                                   |
| ------ | ----------------------------------- | ----- | --------------------------------------------- |
| GET    | `/api/admin/users`                  | admin | List all users (paginated, filter by role)    |
| GET    | `/api/admin/users/:id`              | admin | User detail                                   |
| PATCH  | `/api/admin/users/:id/status`       | admin | Activate or deactivate user                   |
| PATCH  | `/api/admin/users/:id/role`         | ✅    | Change a user's role                          |
| PATCH  | `/api/admin/users/:id/unlock-reset` | ✅    | Unlock an account locked from forgot-password |

**GET `/api/admin/users` query params:** `?page=1&limit=20&role=alumni`

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 42,
    "page": 1,
    "limit": 20,
    "data": [
      {
        "ID": 5,
        "Name": "Siti Rahma",
        "Email": "siti@test.com",
        "Role": "alumni",
        "IsActive": true
      }
    ]
  }
}
```

**PATCH `/api/admin/users/:id/status`:**

```json
{ "is_active": false }
```

**PATCH `/api/admin/users/:id/role`:**

```json
{ "role": "student" }
```

> Valid roles: `alumni` · `student` · `partner` · `admin`

**Business rules:**

- `role` must be one of the valid roles above
- Status updates and role updates both return the updated user object

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

#### PATCH `/api/admin/users/:id/unlock-reset` — Unlock Reset Password

Unlocks an account that was locked out from too many failed forgot-password attempts, resetting the counter to 0.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "message": "akun berhasil dibuka kuncinya"
  }
}
```

---

### Report Moderation

Admin report moderation supports list, detail, resolve, and reject flows. Reports can only be processed while they are still `pending`.

| Method | Endpoint                         | Auth  | Description                                |
| ------ | -------------------------------- | ----- | ------------------------------------------ |
| GET    | `/api/admin/reports`             | admin | List all reports (filter by status)        |
| GET    | `/api/admin/reports/:id`         | admin | Report detail                              |
| PATCH  | `/api/admin/reports/:id/resolve` | admin | Resolve report (optionally delete content) |
| PATCH  | `/api/admin/reports/:id/reject`  | admin | Reject report with reason                  |

**GET `/api/admin/reports` query params:** `?status=pending&page=1&limit=10`

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "total": 12,
    "page": 1,
    "limit": 10,
    "data": [
      {
        "ID": 8,
        "ReporterID": 2,
        "TargetType": "post",
        "TargetID": 12,
        "TargetLabel": "Post \"Reportable Post\"",
        "TargetRedirectURL": "/feed/12",
        "TargetExists": true,
        "ReportType": "spam",
        "Status": "pending",
        "AdminNote": null,
        "ResolvedByID": null,
        "ResolvedAt": null
      }
    ]
  }
}
```

Admin review helper fields:

- `TargetLabel`: Human-readable label for the reported content (title/name/snippet)
- `TargetRedirectURL`: Frontend route to open the reported content directly
- `TargetExists`: `false` when the target content is already deleted/unavailable

These helper fields are included in both:

- `GET /api/admin/reports`
- `GET /api/admin/reports/:id`

**PATCH `/api/admin/reports/:id/resolve`:**

```json
{
  "admin_note": "Content violates community guidelines.",
  "delete_content": true
}
```

> `delete_content: true` → cascades-delete the reported post/group/event/job.
>
> Current implementation only cascades-delete `post`, `group`, `event`, and `job` targets. Reports can also be filed for `comment` and `group_article`, but those target types are not auto-deleted by this moderation action.

**Business rules:**

- A report must still be `pending` to be resolved or rejected
- `admin_note` is optional for resolve, but required for reject
- `delete_content` is best-effort; report processing still continues if the content delete call fails internally
- Rejecting a report sends a `report_rejected` notification to the reporter
- Resolving a report with `delete_content: true` sends `content_deleted_by_admin` to the content owner (for supported target types)

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

This is a separate admin-only moderation path for removing content directly when a report does not exist or is not needed.

Each successful direct deletion sends `content_deleted_by_admin` to the deleted content owner.

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

### Engagement Analytics

Analytics provide a compact admin dashboard for core engagement metrics only.

| Method | Endpoint                           | Auth  | Description                           |
| ------ | ---------------------------------- | ----- | ------------------------------------- |
| GET    | `/api/admin/analytics/overview`    | admin | KPI summary + latest snapshot values  |
| GET    | `/api/admin/analytics/trends`      | admin | Daily trend series (default 30 days)  |
| GET    | `/api/admin/analytics/top-content` | admin | Top N items by views (default 7 days) |

**GET `/api/admin/analytics/overview` response `200`:**

```json
{
  "success": true,
  "data": {
    "total_reactions": 9800,
    "total_comments": 4300,
    "new_users_last_7d": 48,
    "yesterday": {
      "active_users": 142,
      "post_views": 320,
      "group_views": 210
    }
  }
}
```

Notes:

- `total_reactions` and `total_comments` are lifetime totals.
- `new_users_last_7d` is a rolling 7-day count.
- `yesterday.active_users`, `yesterday.post_views`, and `yesterday.group_views` come from the latest completed daily snapshot.

**GET `/api/admin/analytics/trends` query params:** `?days=30`

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "days": 30,
    "data": [
      {
        "date": "2026-04-05",
        "active_users": 98,
        "new_users": 5,
        "post_views": 280,
        "group_views": 210
      }
    ]
  }
}
```

**GET `/api/admin/analytics/top-content` query params:** `?type=post&days=7&limit=10`

`type` must be one of: `post`, `group`.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "type": "post",
    "days": 7,
    "limit": 10,
    "data": [
      {
        "target_id": 42,
        "view_count": 180
      }
    ]
  }
}
```

#### Frontend Notes

- `Total Views` is not returned as a separate backend field; derive it in FE as `post_views + group_views` if needed.
- The trend chart can use `date` for the x-axis and `active_users`, `new_users`, `post_views`, `group_views` for the y-series.
- For top content, only `Posts` and `Groups` are supported in this simplified contract.

---

## Categories

Categories are readable by any authenticated user, but writes are admin-only.

Category changes now propagate to content rows that store category as plain text:

- renaming a category updates matching posts, groups, and portfolio items
- deleting a category reassigns matching content to `General`
- `General` is seeded automatically and treated as the fallback category

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

Returns all mentoring requests (pending + approved + rejected + withdrawn) sent to this mentor.
Each record includes the full `Mentor` and `Student` objects, with their nested `Profile`.

**Response `200`:**

```json
{
  "success": true,
  "data": [
    {
      "ID": 5,
      "CreatedAt": "2026-04-20T10:00:00Z",
      "MentorID": 2,
      "Mentor": {
        "ID": 2,
        "Name": "Alumni Regan",
        "Email": "regan@example.com",
        "Role": "alumni",
        "Profile": {
          "ID": 3,
          "UserID": 2,
          "ProfilePicture": "https://res.cloudinary.com/.../mentor-profile.jpg",
          "Bio": "5+ years in backend and ML",
          "Location": "Jakarta",
          "Skills": "Go, Python, Docker",
          "Interests": "AI, Distributed Systems",
          "MentorQuota": 3,
          "MentorDescription": "Happy to guide students in backend dev"
        }
      },
      "StudentID": 8,
      "Student": {
        "ID": 8,
        "Name": "Student Budi",
        "Email": "budi@example.com",
        "Role": "student",
        "Profile": {
          "ID": 9,
          "UserID": 8,
          "ProfilePicture": "https://res.cloudinary.com/.../student-profile.jpg",
          "Bio": "Final year CS student",
          "Skills": "React, Node.js",
          "Interests": "Web development"
        }
      },
      "Message": "Hi! I'd love to learn Go from you.",
      "Status": "pending",
      "SimilarityScore": 0.5263
    }
  ]
}
```

---

#### PATCH `/api/mentor/requests/:id/approve` — Approve Request

**Business Rules before approval:**

- Request must be `pending`
- Student must not already have 2 approved mentors
- Mentor must have remaining quota capacity
- Approval is executed transactionally to avoid race conditions under concurrent approvals

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

Returns `MentorRequest` records where `Status = "approved"` for the current mentor.
Each record includes the full `Mentor` and `Student` objects, with their nested `Profile`.

**Response `200`:** same shape as `GET /api/mentor/requests` above, filtered to `Status: "approved"` records only.

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
> Session in terminal state (`completed` or `cancelled`) cannot be edited again.
> Completing a session before `session_date` is rejected when `session_date` is in the future.

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

#### GET `/api/mentors/recommend` — Auto Recommendations ⭐ NLP Engine

> Uses **TF-IDF + Cosine Similarity** (pure Go, no external ML dependency) to rank mentors by relevance to the student's profile or a custom query.

**Query params:**

| Param | Description                                                   | Default                                     |
| ----- | ------------------------------------------------------------- | ------------------------------------------- |
| `q`   | Custom free-text query (e.g. `python machine learning cloud`) | Student's `skills + interests` from profile |
| `top` | Number of results to return                                   | `10`                                        |

**How it works:**

1. Reads each mentor's `skills + interests + mentor_bio + position + company + industry + experiences`
2. Parses lightweight constraints from free-text query, including:

- `N years/tahun` minimum experience intent
- `<keyword> industry` intent

3. Applies matching filters (when detected) before scoring
4. Builds TF-IDF corpus (mentors + student query)
5. Computes hybrid score from text similarity + keyword overlap + experience/industry fit

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
      "years_experience": 5,
      "matched_keywords": ["python", "cloud", "machine"],
      "score_breakdown": {
        "text_similarity": 0.61,
        "keyword_overlap": 0.5,
        "experience_fit": 1,
        "industry_fit": 0
      },
      "mentor_quota": 3,
      "similarity_score": 0.5263
    }
  ]
}
```

> Results are sorted by `similarity_score` descending. Score of `0` means no textual overlap.

---

#### POST `/api/mentors/recommend/search` — Natural Language Search ⭐ NLP Engine

> Uses the same **TF-IDF + Cosine Similarity** engine as auto-recommendations, but accepts a **free-text JSON body** instead of a URL query param — suitable for long or complex natural language queries sent from a search form.

**Body (JSON):**

| Field   | Required | Notes                                  |
| ------- | -------- | -------------------------------------- |
| `query` | ✅       | Free-text natural language description |
| `top`   | ❌       | Max results to return (default: `10`)  |

```json
{
  "query": "mentor with 3 year experience in Data Science, with skill of Big Data, Go, Cloud Computing, prefer with interest in machine learning, IoT, Data Mining",
  "top": 10
}
```

**How the query is parsed:**

| Input phrase                                                                                               | Extracted as                                      |
| ---------------------------------------------------------------------------------------------------------- | ------------------------------------------------- |
| `"3 year experience"`                                                                                      | `minYears = 3` (hard filter)                      |
| `"Data Science industry"`                                                                                  | `industry` filter (if `industry` keyword follows) |
| Remaining terms: `"Big Data"`, `"Go"`, `"Cloud Computing"`, `"machine learning"`, `"IoT"`, `"Data Mining"` | TF-IDF scored terms after tokenization + stemming |
| Common words: `"mentor"`, `"skill"`, `"prefer"`, `"interest"`                                              | Filtered by stopword list — no effect on score    |

**Response `200`:** same shape as `GET /api/mentors/recommend`.

**Error cases:**

| Status | Reason                            |
| ------ | --------------------------------- |
| `400`  | `query` field is missing or empty |
| `401`  | Not authenticated                 |
| `403`  | Not a `student` account           |

---

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

**Error cases:**

- `409 Conflict`: duplicate active request (`pending` or `approved`) for the same mentor

**Response `201`:** returns `MentorRequest` with `Status: "pending"`.

---

#### GET `/api/student/mentors` — My Approved Mentors

Returns `MentorRequest` records where `Status = "approved"` for the current student.
Each record includes the full `Mentor` and `Student` objects, with their nested `Profile`.

**Response `200`:** same shape as `GET /api/mentor/requests` above, filtered to `Status: "approved"` records only.

---

#### GET `/api/student/requests` — My Sent Requests

Returns all `MentorRequest` records sent by the current student (all statuses).
Each record includes the full `Mentor` and `Student` objects, with their nested `Profile`.

**Response `200`:** same shape as `GET /api/mentor/requests` above, all statuses included.

---

#### DELETE `/api/student/requests/:id` — Withdraw Pending Request

Withdraws a mentoring request sent by current student.

**Business Rules:**

- Only request owner can withdraw
- Only `pending` requests can be withdrawn
- Withdrawn requests remain in history with `Status = "withdrawn"`

**Response `200`:**

```json
{
  "success": true,
  "data": { "message": "permintaan mentoring berhasil ditarik" }
}
```

---

#### PATCH `/api/mentor/mentees/:id/end` — End Mentorship

Ends an active mentoring relationship for the current mentor.

**Body:**

```json
{
  "reason": "Program selesai"
}
```

> `reason` is optional.

**Business Rules:**

- Only the mentor who owns the relationship can end it
- Only `approved` relationships can be ended
- The relationship becomes inactive with `Status = "ended"`
- Any scheduled sessions for the same mentor-student pair are automatically cancelled
- The end action is recorded with audit fields such as `EndedAt`

**Response `200`:** returns the updated `MentorRequest` with `Status: "ended"`.

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

| Rule                     | Detail                                                                     |
| ------------------------ | -------------------------------------------------------------------------- |
| Mentor role              | `alumni` only                                                              |
| Mentor quota             | Must be `1`, `2`, `3`, or `5`                                              |
| Mentee limit per student | Max **2 approved mentors** at a time                                       |
| Capacity enforcement     | Request blocked when mentor's active mentees ≥ quota                       |
| Duplicate request        | Only one active (`pending` or `approved`) request per student-mentor pair  |
| Session guard            | Session can only be created if an **approved** request exists              |
| Unregister guard         | Cannot unregister while active mentees exist                               |
| Status flow (Request)    | `pending` → `approved` \| `rejected` \| `withdrawn` \| `ended`             |
| Status flow (Session)    | `scheduled` → `completed` \| `cancelled` (terminal)                        |
| End mentorship guard     | Only approved relationships can be ended; ended relationships are inactive |

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
> The follow relationship is immediate and direct; there is no pending/request state.
> Duplicate follow pairs are blocked by the database and surfaced as `409 Conflict`.

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
| `409` | Already following this user (duplicate follow pair) |

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
      "partner_picture": "https://res.cloudinary.com/.../profiles/budi.jpg",
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
        "Sender": {
          "ID": 8,
          "Name": "Student Budi",
          "picture_url": "https://res.cloudinary.com/.../profiles/budi.jpg"
        },
        "Content": "Halo kak!",
        "IsRead": true,
        "CreatedAt": "2026-04-04T10:30:00Z"
      }
    ]
  }
}
```

> Conversation list includes `partner_picture`, and message history includes `Sender.picture_url` when available.

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

## Follow Module

> Allows `alumni` and `student` roles to follow and unfollow other users. Any authenticated user can view the followers and following lists.

| Method | Endpoint                   | Auth | Description                          |
| ------ | -------------------------- | ---- | ------------------------------------ |
| POST   | `/api/users/:id/follow`    | ✅   | Follow a user                        |
| DELETE | `/api/users/:id/follow`    | ✅   | Unfollow a user                      |
| GET    | `/api/users/:id/followers` | ✅   | List users following the target user |
| GET    | `/api/users/:id/following` | ✅   | List users the target user follows   |

### POST `/api/users/:id/follow` — Follow User

Starts following the user with the specified `id`.
Only `student` and `alumni` roles are permitted to follow others.

**Response `201`:**

```json
{
  "success": true,
  "data": {
    "message": "berhasil mengikuti pengguna"
  }
}
```

**Error cases:**
| Status | Reason |
|---|---|
| `400` | Target ID is invalid or trying to follow yourself |
| `401` | Not authenticated |
| `403` | User role is not `alumni` or `student` |
| `409` | Already following this user |

### DELETE `/api/users/:id/follow` — Unfollow User

Stops following the user with the specified `id`.

**Response `200`:**

```json
{
  "success": true,
  "data": {
    "message": "berhasil berhenti mengikuti pengguna"
  }
}
```

**Error cases:**
| Status | Reason |
|---|---|
| `400` | Not currently following this user |

### GET `/api/users/:id/followers` — Get Followers List

Returns a list of users who are following the specified user `id`.

**Response `200`:**

```json
{
  "success": true,
  "data": [
    {
      "ID": 2,
      "Name": "Student Budi",
      "Email": "budi@test.com",
      "Role": "student",
      "IsActive": true,
      "picture_url": "https://res.cloudinary.com/...",
      "Faculty": "Engineering",
      "Major": "Informatics",
      "YearEnroll": 2020,
      "CompanyName": null,
      "Profile": {
        "ID": 5,
        "Bio": "Interested in backend",
        "Position": null,
        "CompanyName": null
      }
    }
  ]
}
```

### GET `/api/users/:id/following` — Get Following List

Returns a list of users that the specified user `id` is currently following.

**Response `200`:** (Same structure as `GET /api/users/:id/followers`)

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
