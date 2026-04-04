# API Reference – Alumni Community Platform

Base URL: `http://localhost:8080`  
All protected endpoints require: `Authorization: Bearer <token>`  
Request body must include `Content-Type: application/json` for JSON  endpoints or `Content-Type: multipart/form-data` for file/form endpoints.

---

## Table of Contents

- [Response Format](#response-format)
- [Auth](#auth)
- [User Profile](#user-profile)
- [Company Profile](#company-profile)
- [Portfolio](#portfolio)
- [Feed Posting](#feed-posting)
- [Group Forum](#group-forum)
- [Events](#events)
- [Jobs](#jobs)
- [Content Reporting](#content-reporting)
- [Admin Module](#admin-module)
- [Categories](#categories)
- [Mentoring Module](#mentoring-module)
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
  "role": "partner",
  "company_name": "PT Maju Bersama"
}
```

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
      "ID": 1,
      "Name": "Regan Putra",
      "Email": "regan@test.com",
      "Role": "alumni",
      "IsActive": true
    }
  }
}
```

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
    "user_id": 1,
    "email": "regan@test.com",
    "role": "alumni"
  }
}
```

---

## User Profile

> `alumni` and `student` only. All create/update use `multipart/form-data`.

| Method | Endpoint                      | Auth | Description                       |
| ------ | ----------------------------- | ---- | --------------------------------- |
| POST   | `/api/profile`                | ✅   | Create profile                    |
| GET    | `/api/profile`                | ✅   | Get own profile                   |
| PUT    | `/api/profile`                | ✅   | Update profile                    |
| DELETE | `/api/profile`                | ✅   | Delete profile                    |
| POST   | `/api/profile/experience`     | ✅   | Add experience entry              |
| PUT    | `/api/profile/experience/:id` | ✅   | Update experience entry           |
| DELETE | `/api/profile/experience/:id` | ✅   | Delete experience entry           |

### POST/PUT `/api/profile` — Create / Update Profile

**Content-Type:** `multipart/form-data`

| Field                   | Required | Notes                                    |
| ----------------------- | -------- | ---------------------------------------- |
| `job_status`            | ✅       | See values below                         |
| `bio`                   | ❌       | —                                        |
| `phone`                 | ❌       | —                                        |
| `linkedin_url`          | ❌       | —                                        |
| `picture`               | ❌       | Image file → Cloudinary                  |
| `position`              | if `employed` | Current job title                 |
| `company_name`          | if `employed` | Current company name              |
| `industry_name`         | if `entrepreneur` | Industry                      |
| `educational_level`     | if `continuing_study` | Degree level              |
| `advanced_study_program`| if `continuing_study` | Study program             |
| `institution_name`      | if `continuing_study` | University name           |
| `status_description`    | if `unemployed`/`freelance` | Description         |

**`job_status` values:** `employed` · `entrepreneur` · `continuing_study` · `unemployed` · `freelance` · `student`

**Response `201` (create) / `200` (update):**
```json
{
  "success": true,
  "data": {
    "ID": 1,
    "UserID": 1,
    "Bio": "Software engineer at tech startup",
    "JobStatus": "employed",
    "Position": "Backend Developer",
    "CompanyName": "PT Startup Indonesia",
    "PictureURL": "https://res.cloudinary.com/.../profile.jpg",
    "LinkedinURL": "https://linkedin.com/in/regan"
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

## Company Profile

> `partner` role only.

| Method | Endpoint       | Auth | Description                        |
| ------ | -------------- | ---- | ---------------------------------- |
| POST   | `/api/company` | ✅   | Create or join company profile     |
| GET    | `/api/company` | ✅   | View own company profile           |
| PUT    | `/api/company` | ✅   | Update company profile             |
| DELETE | `/api/company` | ✅   | Delete company profile             |

> Partners with the same `company_name` (from registration) **share one profile**. POST returns `201` if created, `200` if joined.

**Body (JSON):**
```json
{
  "industry_type": "Technology",
  "location": "Jakarta, Indonesia",
  "employee_size": 150,
  "website_url": "https://company.com"
}
```

---

## Portfolio

> `alumni` and `student` only. Create/update use `multipart/form-data`.

| Method | Endpoint             | Auth | Description          |
| ------ | -------------------- | ---- | -------------------- |
| POST   | `/api/portfolio`     | ✅   | Create portfolio item |
| GET    | `/api/portfolio`     | ✅   | List own items        |
| PUT    | `/api/portfolio/:id` | ✅   | Update item           |
| DELETE | `/api/portfolio/:id` | ✅   | Delete item           |

**Fields (form-data):**

| Field         | Required | Notes                      |
| ------------- | -------- | -------------------------- |
| `title`       | ✅       | —                          |
| `description` | ❌       | —                          |
| `category`    | ❌       | e.g. `"Mobile Development"` |
| `tags`        | ❌       | Comma-separated            |
| `start_date`  | ❌       | Format: `YYYY-MM`          |
| `end_date`    | ❌       | Format: `YYYY-MM`          |
| `media`       | ❌       | Image file → Cloudinary    |

---

## Feed Posting

> All roles can read/react/vote. `alumni` and `student` can create posts/comments.

### Endpoints

| Method | Endpoint                  | Auth | Body        | Description                           |
| ------ | ------------------------- | ---- | ----------- | ------------------------------------- |
| GET    | `/api/feed`               | ✅   | —           | List posts (paginated, counts only)   |
| GET    | `/api/feed/:id`           | ✅   | —           | Full post with nested comments        |
| POST   | `/api/feed`               | ✅   | `form-data` | Create post                           |
| PUT    | `/api/feed/:id`           | ✅   | `form-data` | Update own post                       |
| DELETE | `/api/feed/:id`           | ✅   | —           | Delete own post                       |
| POST   | `/api/feed/:id/comments`  | ✅   | JSON        | Add comment or reply                  |
| POST   | `/api/feed/:id/react`     | ✅   | JSON        | React to a post                       |
| POST   | `/api/feed/:id/vote`      | ✅   | JSON        | Vote on a post                        |
| PUT    | `/api/comments/:id`       | ✅   | JSON        | Update own comment                    |
| DELETE | `/api/comments/:id`       | ✅   | —           | Delete own comment                    |
| POST   | `/api/comments/:id/react` | ✅   | JSON        | React to a comment                    |
| POST   | `/api/comments/:id/vote`  | ✅   | JSON        | Vote on a comment                     |

### POST `/api/feed` — Create Post

**Content-Type:** `multipart/form-data`

| Field      | Required | Notes                    |
| ---------- | -------- | ------------------------ |
| `title`    | ✅       | —                        |
| `content`  | ✅       | —                        |
| `category` | ❌       | Free text                |
| `image`    | ❌       | Image file → Cloudinary  |

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
    "posts": [
      {
        "ID": 3,
        "UserID": 2,
        "Title": "Hello everyone",
        "Content": "My first post!",
        "Category": "General",
        "ImageURL": null,
        "CommentCount": 4,
        "ReactionCount": 7,
        "VoteCount": 3,
        "CreatedAt": "2026-03-04T19:25:24+07:00"
      }
    ]
  }
}
```

### GET `/api/feed/:id` — Post Detail

**Response `200`:**
```json
{
  "success": true,
  "data": {
    "ID": 3,
    "Title": "Hello everyone",
    "Content": "My first post!",
    "User": { "ID": 2, "Name": "Regan Putra", "Role": "alumni" },
    "Reactions": [{ "ID": 1, "UserID": 2, "Type": "like" }],
    "Votes": [{ "ID": 1, "UserID": 3, "Value": 1 }],
    "Comments": [
      {
        "ID": 1,
        "Content": "Great post!",
        "User": { "ID": 3, "Name": "Ani" },
        "Reactions": [],
        "Votes": [],
        "Replies": [
          {
            "ID": 2,
            "Content": "Agreed!",
            "ParentCommentID": 1,
            "Replies": []
          }
        ]
      }
    ]
  }
}
```

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

> All roles can read. `alumni` and `student` can create/join/interact.

### Group Endpoints

| Method | Endpoint                          | Auth | Body        | Description                      |
| ------ | --------------------------------- | ---- | ----------- | -------------------------------- |
| GET    | `/api/groups`                     | ✅   | —           | List all groups                  |
| POST   | `/api/groups`                     | ✅   | `form-data` | Create group                     |
| GET    | `/api/groups/joined`              | ✅   | —           | Groups current user belongs to   |
| GET    | `/api/groups/:id`                 | ✅   | —           | Group detail                     |
| PUT    | `/api/groups/:id`                 | ✅   | `form-data` | Update group (owner only)        |
| DELETE | `/api/groups/:id`                 | ✅   | —           | Delete group + all data          |
| POST   | `/api/groups/:id/join`            | ✅   | —           | Join group                       |
| DELETE | `/api/groups/:id/leave`           | ✅   | —           | Leave group                      |
| GET    | `/api/groups/:id/members`         | ✅   | —           | List members                     |
| DELETE | `/api/groups/:id/members/:userID` | ✅   | —           | Kick member (owner only)         |

### Group Article Endpoints

| Method | Endpoint                            | Auth | Body        | Description                        |
| ------ | ----------------------------------- | ---- | ----------- | ---------------------------------- |
| POST   | `/api/groups/:id/articles`          | ✅   | `form-data` | Create article (members only)      |
| GET    | `/api/groups/articles/:id`          | ✅   | —           | Article detail + nested comments   |
| PUT    | `/api/groups/articles/:id`          | ✅   | `form-data` | Update own article                 |
| DELETE | `/api/groups/articles/:id`          | ✅   | —           | Delete own article                 |
| POST   | `/api/groups/articles/:id/comments` | ✅   | JSON        | Add comment or reply               |
| POST   | `/api/groups/articles/:id/react`    | ✅   | JSON        | React to article (members only)    |
| PUT    | `/api/groups/comments/:id`          | ✅   | JSON        | Update own comment                 |
| DELETE | `/api/groups/comments/:id`          | ✅   | —           | Delete own comment                 |
| POST   | `/api/groups/comments/:id/react`    | ✅   | JSON        | React to comment (members only)    |

### POST `/api/groups` — Create Group (form-data)

| Field         | Required | Notes                    |
| ------------- | -------- | ------------------------ |
| `title`       | ✅       | —                        |
| `category`    | ✅       | —                        |
| `description` | ❌       | —                        |
| `rules`       | ❌       | —                        |
| `banner`      | ❌       | Image file → Cloudinary  |

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
- Group owner cannot leave their own group
- When a group is deleted: articles, comments, reactions, memberships all cascade-deleted

---

## Events

> All roles can read. `alumni` and `student` can create/register. `partner` is blocked.

### Endpoints

| Method | Endpoint                       | Auth | Body        | Description                           |
| ------ | ------------------------------ | ---- | ----------- | ------------------------------------- |
| GET    | `/api/events`                  | ✅   | —           | List events (paginated)               |
| POST   | `/api/events`                  | ✅   | `form-data` | Create event                          |
| GET    | `/api/events/:id`              | ✅   | —           | Event detail with agendas & participants |
| PUT    | `/api/events/:id`              | ✅   | `form-data` | Update own event                      |
| DELETE | `/api/events/:id`              | ✅   | —           | Delete event + cascade                |
| POST   | `/api/events/:id/register`     | ✅   | —           | Register for event                    |
| DELETE | `/api/events/:id/register`     | ✅   | —           | Cancel registration                   |
| GET    | `/api/events/:id/participants` | ✅   | —           | List participants                     |
| POST   | `/api/events/:id/agenda`       | ✅   | JSON        | Add agenda item (owner only)          |
| PUT    | `/api/agenda/:id`              | ✅   | JSON        | Update agenda item (owner only)       |
| DELETE | `/api/agenda/:id`              | ✅   | —           | Delete agenda item (owner only)       |

### POST `/api/events` — Create Event (form-data)

| Field         | Required | Notes                                                         |
| ------------- | -------- | ------------------------------------------------------------- |
| `title`       | ✅       | —                                                             |
| `description` | ❌       | —                                                             |
| `location`    | ❌       | —                                                             |
| `capacity`    | ❌       | Integer ≥ 0 (0 = unlimited)                                   |
| `status`      | ❌       | `upcoming` (default) · `ongoing` · `completed` · `cancelled`  |
| `photo`       | ❌       | Image file → Cloudinary                                       |

**Response `201`:**
```json
{
  "success": true,
  "data": {
    "ID": 5,
    "Title": "Alumni Seminar 2026",
    "Location": "Aula Besar, Kampus A",
    "Capacity": 50,
    "Status": "upcoming",
    "PhotoURL": null,
    "OwnerID": 2,
    "CreatedAt": "2026-04-01T08:00:00+07:00"
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

- Registration blocked when `status` is `completed` or `cancelled`
- Maximum registrations enforced when `capacity > 0`
- Duplicate registration rejected
- Only the creator can update, delete, and manage agendas
- Delete cascades: all agendas + registrations removed

---

## Jobs

> `alumni` and `partner` can post jobs. `alumni` and `student` can apply.

### Endpoints

| Method | Endpoint                            | Auth | Body        | Description                            |
| ------ | ----------------------------------- | ---- | ----------- | -------------------------------------- |
| GET    | `/api/jobs`                         | ✅   | —           | List jobs (paginated + filters)        |
| POST   | `/api/jobs`                         | ✅   | `form-data` | Create job posting                     |
| GET    | `/api/jobs/:id`                     | ✅   | —           | Job detail                             |
| PUT    | `/api/jobs/:id`                     | ✅   | `form-data` | Update own posting                     |
| DELETE | `/api/jobs/:id`                     | ✅   | —           | Delete own posting                     |
| POST   | `/api/jobs/:id/apply`               | ✅   | `form-data` | Apply for job                          |
| GET    | `/api/jobs/:id/applicants`          | ✅   | —           | View applicants (owner only)           |
| GET    | `/api/jobs/applications/mine`       | ✅   | —           | View my applications                   |
| PUT    | `/api/jobs/applications/:id/status` | ✅   | JSON        | Update application status (owner only) |

### GET `/api/jobs` — Query Parameters

| Param      | Description                                          |
| ---------- | ---------------------------------------------------- |
| `search`   | Search by title or company name                      |
| `job_type` | `full-time` · `part-time` · `internship` · `freelance` · `remote` |
| `status`   | `open` · `closed`                                    |
| `page`     | Default: `1`                                         |
| `limit`    | Default: `10`                                        |

### POST `/api/jobs` — Create Job (form-data)

| Field          | Required | Notes                                                            |
| -------------- | -------- | ---------------------------------------------------------------- |
| `title`        | ✅       | —                                                                |
| `company_name` | ✅       | —                                                                |
| `job_type`     | ✅       | `full-time` · `part-time` · `internship` · `freelance` · `remote` |
| `description`  | ❌       | —                                                                |
| `location`     | ❌       | —                                                                |
| `salary_range` | ❌       | e.g. `"5.000.000 - 8.000.000"`                                   |
| `status`       | ❌       | `open` (default) · `closed`                                      |
| `image`        | ❌       | Image file → Cloudinary                                          |

### POST `/api/jobs/:id/apply` — Apply for Job (form-data)

| Field          | Required | Notes                           |
| -------------- | -------- | ------------------------------- |
| `resume`       | ✅*      | PDF file → Cloudinary           |
| `resume_url`   | ✅*      | URL (if not uploading file)     |
| `cover_letter` | ❌       | Optional text                   |

> *One of `resume` (file) or `resume_url` (link) is required.

### PUT `/api/jobs/applications/:id/status` — Update Status (JSON)

```json
{ "status": "accepted" }
```
> Status values: `pending` · `reviewed` · `accepted` · `rejected`

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

---

## Content Reporting

> `alumni`, `student`, and `admin` can submit reports.

| Method | Endpoint            | Auth | Description                 |
| ------ | ------------------- | ---- | --------------------------- |
| POST   | `/api/reports`      | ✅   | Submit a report             |
| GET    | `/api/reports/mine` | ✅   | View own submitted reports  |

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

| Method | Endpoint                         | Auth  | Description                              |
| ------ | -------------------------------- | ----- | ---------------------------------------- |
| GET    | `/api/admin/reports`             | admin | List all reports (filter by status)      |
| GET    | `/api/admin/reports/:id`         | admin | Report detail                            |
| PATCH  | `/api/admin/reports/:id/resolve` | admin | Resolve report (optionally delete content) |
| PATCH  | `/api/admin/reports/:id/reject`  | admin | Reject report with reason                |

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
| GET    | `/api/categories`           | ✅    | List all categories *(any authenticated user)* |
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

| Action | alumni | student | partner | admin |
|---|---|---|---|---|
| Register / manage as mentor | ✅ | ❌ | ❌ | ❌ |
| Browse & request mentors | ❌ | ✅ | ❌ | ❌ |
| Get recommendations | ❌ | ✅ | ❌ | ❌ |
| Create / manage sessions | ✅ (mentor side) | ❌ | ❌ | ❌ |
| View own sessions | ✅ | ✅ | ❌ | ❌ |

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

#### GET `/api/mentors/recommend` — Get Recommendations  ⭐ NLP Engine

> Uses **TF-IDF + Cosine Similarity** (pure Go, no external ML dependency) to rank mentors by relevance to the student's profile or a custom query.

**Query params:**

| Param | Description | Default |
|---|---|---|
| `q` | Custom free-text query (e.g. `python machine learning cloud`) | Student's `skills + interests` from profile |
| `top` | Number of results to return | `10` |

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

#### GET `/api/student/sessions` — My Sessions (Student)

Returns all `MentoringSession` records where the caller is the student. Preloads `Mentor`.

---

### Business Rules Summary

| Rule | Detail |
|---|---|
| Mentor role | `alumni` only |
| Mentor quota | Must be `1`, `2`, `3`, or `5` |
| Mentee limit per student | Max **2 approved mentors** at a time |
| Capacity enforcement | Request blocked when mentor's active mentees ≥ quota |
| Duplicate request | Only one `pending` or `approved` request per student-mentor pair |
| Session guard | Session can only be created if an **approved** request exists |
| Unregister guard | Cannot unregister while active mentees exist |
| Status flow (Request) | `pending` → `approved` \| `rejected` |
| Status flow (Session) | `scheduled` → `completed` \| `cancelled` |

---

## Cloudinary Upload Reference

Every file field is **optional** — omit to skip upload. Accepted formats: `jpg` · `jpeg` · `png` · `webp`

| Module           | Form field | Cloudinary folder                 | Endpoint                    |
| ---------------- | ---------- | --------------------------------- | --------------------------- |
| Profile picture  | `picture`  | `alumni-platform/profiles`        | POST/PUT `/api/profile`     |
| Feed post image  | `image`    | `alumni-platform/feed`            | POST/PUT `/api/feed`        |
| Group banner     | `banner`   | `alumni-platform/groups/banners`  | POST/PUT `/api/groups`      |
| Group article    | `media`    | `alumni-platform/groups/articles` | POST/PUT `/api/groups/articles/:id` |
| Portfolio item   | `media`    | `alumni-platform/portfolio`       | POST/PUT `/api/portfolio`   |
| Event photo      | `photo`    | `alumni-platform/events`          | POST/PUT `/api/events`      |
| Job image        | `image`    | `alumni-platform/jobs`            | POST/PUT `/api/jobs`        |
| Job resume (PDF) | `resume`   | `alumni-platform/resumes`         | POST `/api/jobs/:id/apply`  |

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
Allowed headers: `Origin, Content-Type, Accept, Authorization`
