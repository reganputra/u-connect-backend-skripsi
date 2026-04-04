# 📘 Project Plan – Alumni Community Platform

## 1. Project Overview

The **Alumni Community Platform** is a RESTful API designed to support an alumni networking system.  
The backend follows a layered architecture pattern:

Controller → Service → Repository → Database

# 2. Authentication Module

## 3.1 Features

- User Registration
- User Login
- JWT-based Authentication
- Role-based Access Control

## 3.2 Supported Roles

The system supports three roles:

- `alumni`
- `student`
- `partner`

# 4. User Profile Module

The **User Profile Module** manages detailed profile information for users after registration.  
Each registered user (alumni, student) can create, update, or delete their profile data.

## 4.1 Features

- Create profile
- Update profile
- Delete profile
- Upload / Update profile picture
- View profile details

## 4.2 Profile Fields

The profile contains the following attributes:

### Basic Information

- profile_picture (URL or file path)
- major (from registered data)
- faculty (from registered data)
- year_enrolled (from registered data)
- location

### Professional Information (for alumni if work as employed, entrepreneur, etc.)

- job_status (employed, unemployed, student, entrepreneur, etc.)
- position
- company_name (employed)
- industry_name (entrepreneur)
- location
- salary
- experience (job experience)

### Academic Information (if alumni decided to continue study)

- educational_level
- advanced_study_program

### Skills & Interests

- skills (array or comma-separated string)
- interests (array or comma-separated string)

### Mentorship Information (for alumni only)

- mentor_quota (number of mentees allowed)
- mentor_description

### Additional Information (if alumni is unemployed or freelance)

- status_description

## 4.3 Relationships

- One user has one profile (1:1 relationship)
- `user_profiles.user_id` references `users.id`

## 4.4 Business Rules

- Salary field is optional

# 5. Company Profile Module

The **Company Profile Module** is designed specifically for users with the `partner` role.  
This module allows registered partners to create and manage their company information within the Alumni Community Platform.

## 5.1 Features

- Create company profile
- Update company profile
- Delete company profile
- View company profile

## 5.2 Company Profile Fields

The company profile includes the following attributes:

- company_name (can be used during registration)
- industry_type
- location
- employee_size
- website_url

## 5.3 Relationships

- Many partner users can have one company profile (1:N relationship)

# 5.4 Business Rules

- Employee size must be zero or positive
- if there are more than one user from same company, they will share the same company profile, and can update the same company profile

# 6. Portfolio Module

The **Portfolio Module** allows users to showcase their achievements, experiences (non professional), and projects within the Alumni Community Platform.

This module is available only for users with the following roles:

- `student`
- `alumni`

## 6.1 Features

- Create portfolio item
- Update portfolio item
- Delete portfolio item
- View portfolio items (by user)
- Upload media (photo or file link)

## 6.2 Portfolio Fields

Each portfolio item contains:

- title
- description
- category
- tags
- start_date
- end_date
- media_url (photo or file)

## 6.3 Relationships

- One user can have many portfolio entries (1:N relationship)

## 6.4 Business Rules

- Only `student` and `alumni` roles can create portfolios
- Title is required

# 7. Feed Posting Module

The **Feed Posting Module** provides a social interaction feature similar to platforms like Instagram or LinkedIn.  
Users can create posts, interact with other users’ posts, leave comments, reply to comments, and react or vote on posts and comments.
All authenticated users (`alumni`, `student`) can participate in the feed.

## 7.1 Features

Post Features:

- Create post
- Update post
- Delete post
- View feed posts
- View post details

Interaction Features:

- Add comments to posts
- Reply to comments
- React to posts
- React to comments
- Upvote or downvote posts
- Upvote or downvote comments

## 7.2 Post Fields

Each post contains:

- category (optional)
- title
- content (text, links, or formatted content)
- url_image (optional)

## 7.3 Comment Fields

Each comment contains:

- post_id
- user_id
- content
- parent_comment_id (for replies, null for top-level comments)

## 7.4 Reaction Fields

Each reaction contains:

- post_id (optional)
- comment_id (optional)
- user_id
- reaction_type (like, love, haha, wow, sad, angry)

# 8. Group Forum Module

The **Group Forum Module** provides micro-communities inside the Alumni Community Platform.  
Users can create and participate in discussion groups based on shared interests such as industry, skills, hobbies, sports, games, or academic topics.

This feature is conceptually similar to Facebook Groups where members can interact through posts, comments, and reactions within a specific community.
Only users with the roles:

- `student`
- `alumni`

are allowed to create and participate in groups.

## 8.1 Features

Group Management

- Create group
- Update group information
- Update group rules
- Update group banner
- Delete group
- Kick member from group

Membership

- Join group
- Leave group
- View group members
- View joined groups

Group Interaction

- Create article/post inside group
- Update article
- Delete article
- Comment on article
- Reply to comment
- React to article
- React to comment

## 8.2 Access Control

Group Owner Permissions:

- Update group details
- Update group rules
- Change group banner
- CRUD articles
- Kick members
- Delete the group
- Comment on articles
- Reply to comments
- React to articles and comments

Group Member Permissions:

- CRUD articles
- Comment on articles
- Reply to comments
- React to articles and comments

# 8.3 Group Fields

- category
- title
- description
- url_media (optional banner or image)

Article Fields:

- title
- content (text etc)
- url_media (optional banner or image)

# 8.4 Relationships

- One user can create many groups
- One group has many members
- One group has many articles
- One article has many comments
- Comments support nested replies via `parent_id`
- Articles and comments can receive reactions

# 8.5 Business Rules

- A user can join multiple groups
- A group must always have one owner
- When the owner deletes a group, all related data must be deleted
- Only group members can create articles
- Only group owner can remove members
- user who not a member can only see preview of the group content

# 9 Event Module

The **Event Module** enables users to organize and participate in events within the Alumni Community Platform.  
This feature allows students and alumni to create events such as seminars, webinars, workshops, alumni gatherings, or community meetups.

Users can create events, manage event details, and allow other users to register or cancel their registration.
Only users with the roles:

- `student`
- `alumni`

## 9.1 Features

Event Management

- CRUD Event
- View event list
- View event details

Event Registration

- Register for event
- Cancel event registration
- View registered participants

Event Agenda

- Add event agenda items
- Update agenda
- Remove agenda items

# 9.2 Fields

Each event contains the following attributes:

- title
- description
- url_photo
- location
- capacity
- status

Event status examples:

- upcoming
- ongoing
- completed
- cancelled

Agenda Fields:

- description
- agenda_time

# 9.3 Relationships

- Only `student` and `alumni` roles can create events
- Event capacity cannot be exceeded
- When an event is deleted, all related agendas and registrations must be deleted
- Registration is not allowed when event status is `completed` or `cancelled`
- A user can register only once per event

# 10 Job Module

The **Job Module** provides job opportunity management within the Alumni Community Platform.  
This feature allows alumni and partner users to post job vacancies, while students and alumni can search and apply for available jobs.

The module also provides a job applicant management system where job owners can review applicants, view their profiles, download uploaded resumes or portfolios, and update the application status.

## 10.1 Authorization Roles

Job Creation:

- `alumni`
- `partner`

Job Application:

- `student`
- `alumni`

Job Management (applicants and status):

- Only the **job owner** can manage applicants.

## 10.2 Features

Job Management

- Create job
- Update job
- Delete job
- View job list
- View job details
- Search jobs

Application Management

- Apply for job
- Upload resume / CV / portfolio
- Write cover letter
- View application status

Applicant Management (Job Owner)

- View all applicants
- View applicant profile details
- Download uploaded resume / CV / portfolio
- Update applicant status

# 10.3 Field

Each job contains the following attributes:

- title
- description
- company_name
- location
- job_type
- status
- salary_range
- url_image (optional)

Job type examples:

- full-time
- part-time
- internship
- contract
- freelance

Job status examples:

- open
- closed
- filled

Each applicant contains:

- applicant_id
- job_id
- user_id
- cover_letter
- url_file (resume / CV / portfolio)

# 10.4 Relationships

- One user can create many jobs (1:N)
- One job can have many applicants (1:N)
- One user can apply to multiple jobs (N:M)

# 10.5 Business Rules

- Only `alumni` and `partner` roles can create job posts
- Only `student` and `alumni` roles can apply for jobs
- A user can apply only once per job
- Applicants must upload a resume/CV/portfolio file
- Job owners can view applicant profiles and application details
- Only the job owner can change the applicant status
- When a job is deleted, all related applications must also be deleted

# 11 Admin Module
The **Admin Module** introduces a new role: `admin`, which is responsible for managing, moderating, and maintaining the overall system.

Admins have elevated privileges to monitor platform activity, manage users, moderate content, and oversee system-wide configurations such as categories.

This module ensures platform integrity, safety, and proper content governance.

Admin responsibilities:
- Monitor system activity
- Moderate user-generated content
- Manage users and roles
- View platform statistics
- Manage categories

## 11.1 Features
Dashboard Admin
- View platform statistics
- Overview of system activity

Content Moderation
- Admin can received a report from user
- Admin can delete or reject with reason
- Delete posts
- Delete group forums
- Delete events
- Delete jobs

User Management
- View all users
- Update user status (active / inactive)
- Change user role (student, alumni, partner, admin)

Category Management
- Create category
- Update category
- Delete category
- View category list
