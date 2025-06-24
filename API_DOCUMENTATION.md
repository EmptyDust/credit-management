# API Documentation

This document provides a comprehensive and detailed overview of the API endpoints for the Credit Management System. All endpoints are proxied through the API Gateway and are prefixed with `/api`.

**Important Note**: All user, student, and teacher IDs are now UUID strings instead of integers for better security and uniqueness.

---

## Table of Contents

- [API Documentation](#api-documentation)
  - [Table of Contents](#table-of-contents)
  - [1. Health Check](#1-health-check)
    - [`GET /health`](#get-health)
  - [2. Authentication Service (`/auth`)](#2-authentication-service-auth)
  - [3. User Management Service (`/users`)](#3-user-management-service-users)
  - [4. Permission Management Service (`/permissions`)](#4-permission-management-service-permissions)
    - [Role Management](#role-management)
    - [Permission Management](#permission-management)
    - [Assignment Management](#assignment-management)
    - [Query Endpoints](#query-endpoints)
  - [5. Student Info Service (`/students`)](#5-student-info-service-students)
  - [6. Teacher Info Service (`/teachers`)](#6-teacher-info-service-teachers)
  - [7. Affair Management Service (`/affairs`)](#7-affair-management-service-affairs)
    - [Affair Model](#affair-model)
    - [AffairStudent Model](#affairstudent-model)
    - [Endpoints](#endpoints)
      - [Create Affair Example](#create-affair-example)
      - [Get Affair with Participants Example](#get-affair-with-participants-example)
      - [Get Affair Applications Example](#get-affair-applications-example)
  - [8. Application Management Service (`/applications`)](#8-application-management-service-applications)
    - [Application Model](#application-model)
    - [Five Credit Types](#five-credit-types)
      - [1. Innovation Practice Credit (创新创业实践活动学分)](#1-innovation-practice-credit-创新创业实践活动学分)
      - [2. Discipline Competition Credit (学科竞赛学分)](#2-discipline-competition-credit-学科竞赛学分)
      - [3. Student Entrepreneurship Project Credit (大学生创业项目学分)](#3-student-entrepreneurship-project-credit-大学生创业项目学分)
      - [4. Entrepreneurship Practice Credit (创业实践项目学分)](#4-entrepreneurship-practice-credit-创业实践项目学分)
      - [5. Paper Patent Credit (论文专利学分)](#5-paper-patent-credit-论文专利学分)
    - [Endpoints](#endpoints-1)
    - [Application Workflow](#application-workflow)
    - [Permission Control](#permission-control)
    - [Example Requests](#example-requests)
      - [Batch Create Applications](#batch-create-applications)
      - [Update Application Details](#update-application-details)
      - [Submit Application](#submit-application)
      - [Review Application](#review-application)
  - [9. Permission Management Service (`/permissions`)](#9-permission-management-service-permissions)
    - [Endpoints](#endpoints-2)
      - [Initialize Permissions Example](#initialize-permissions-example)

---

## 1. Health Check

### `GET /health`

Checks the operational status of the API Gateway.

-   **Response (200 OK):**
    ```json
    {
      "status": "ok",
      "service": "api-gateway",
      "version": "1.0.0"
    }
    ```

---

## 2. Authentication Service (`/auth`)

Handles user authentication and token management. All endpoints are prefixed with `/api/auth`.

| Method | Endpoint             | Description                      | Authentication |
| :----- | :------------------- | :------------------------------- | :------------- |
| `POST` | `/login`             | Logs in a user.                  | None           |
| `POST` | `/validate-token`    | Validates a JWT.                 | None           |
| `POST` | `/refresh-token`     | Refreshes an access token.       | Required       |
| `POST` | `/logout`            | Logs out a user.                 | Required       |

---

## 3. User Management Service (`/users`)

Handles user profiles and administrative user management. All endpoints are prefixed with `/api/users`.

| Method   | Endpoint             | Description                               | Authentication            |
| :------- | :------------------- | :---------------------------------------- | :------------------------ |
| `POST`   | `/register`          | Registers a new user.                     | None                      |
| `GET`    | `/stats`             | Gets user statistics.                     | Admin Required            |
| `GET`    | `/profile`           | Gets the current user's profile.          | Required                  |
| `PUT`    | `/profile`           | Updates the current user's profile.       | Required                  |
| `GET`    | `/:id`               | Gets a specific user's profile by UUID.   | Admin Required            |
| `PUT`    | `/:id`               | Updates a specific user's profile by UUID.| Admin Required            |
| `DELETE` | `/:id`               | Deletes a user by UUID.                   | Admin Required            |
| `GET`    | ``                   | Gets a list of all users.                 | Admin Required            |
| `GET`    | `/type/:userType`    | Gets users by their type (`student`/`teacher`). | Admin Required      |

---

## 4. Permission Management Service (`/permissions`)

Handles roles, permissions, and their assignments. All endpoints are prefixed with `/api/permissions` and require Admin privileges.

### Role Management

| Method   | Endpoint             | Description           |
| :------- | :------------------- | :-------------------- |
| `POST`   | `/roles`             | Creates a new role.   |
| `GET`    | `/roles`             | Gets all roles.       |
| `GET`    | `/roles/:roleID`     | Gets a specific role. |
| `PUT`    | `/roles/:roleID`     | Updates a role.       |
| `DELETE` | `/roles/:roleID`     | Deletes a role.       |

### Permission Management

| Method   | Endpoint        | Description              |
| :------- | :-------------- | :----------------------- |
| `POST`   | ``              | Creates a permission.    |
| `GET`    | ``              | Gets all permissions.    |
| `GET`    | `/:id`          | Gets a single permission.|
| `DELETE` | `/:id`          | Deletes a permission.    |

### Assignment Management

| Method   | Endpoint                                   | Description                         |
| :------- | :----------------------------------------- | :---------------------------------- |
| `POST`   | `/users/:userID/roles`                     | Assigns a role to a user.           |
| `DELETE` | `/users/:userID/roles/:roleID`             | Removes a role from a user.         |
| `POST`   | `/users/:userID/permissions`               | Assigns a permission to a user.     |
| `DELETE` | `/users/:userID/permissions/:permissionID` | Removes a permission from a user.   |
| `POST`   | `/roles/:roleID/permissions`               | Assigns a permission to a role.     |
| `DELETE` | `/roles/:roleID/permissions/:permissionID` | Removes a permission from a role.   |

### Query Endpoints

| Method | Endpoint                   | Description                      |
| :----- | :------------------------- | :------------------------------- |
| `GET`  | `/users/:userID/roles`       | Gets a user's roles.             |
| `GET`  | `/users/:userID/permissions` | Gets a user's permissions.       |

---

## 5. Student Info Service (`/students`)

Manages detailed student information. All endpoints are prefixed with `/api/students`.

| Method   | Endpoint          | Description                             | Authentication |
| :------- | :---------------- | :-------------------------------------- | :------------- |
| `POST`   | ``                | Creates a new student record.           | Admin Required |
| `GET`    | `/:id`            | Gets a student by UUID.                 | Required       |
| `PUT`    | `/:id`            | Updates a student's info by UUID.       | Required       |
| `DELETE` | `/:id`            | Deletes a student by UUID.              | Admin Required |
| `GET`    | ``                | Gets a list of all students.            | Required       |
| `GET`    | `/college/:college`| Gets students by college.              | Required       |
| `GET`    | `/major/:major`   | Gets students by major.                 | Required       |
| `GET`    | `/class/:class`   | Gets students by class.                 | Required       |
| `GET`    | `/status/:status` | Gets students by status.                | Required       |
| `GET`    | `/search`         | Searches for students (`?q=...`).       | Required       |

---

## 6. Teacher Info Service (`/teachers`)

Manages detailed teacher information. All endpoints are prefixed with `/api/teachers`.

| Method   | Endpoint              | Description                               | Authentication |
| :------- | :-------------------- | :---------------------------------------- | :------------- |
| `POST`   | ``                    | Creates a new teacher record.             | Admin Required |
| `GET`    | `/:id`                | Gets a teacher by UUID.                   | Required       |
| `PUT`    | `/:id`                | Updates a teacher's info by UUID.         | Required       |
| `DELETE` | `/:id`                | Deletes a teacher by UUID.                | Admin Required |
| `GET`    | ``                    | Gets a list of all teachers.              | Required       |
| `GET`    | `/department/:department`| Gets teachers by department.           | Required       |
| `GET`    | `/title/:title`       | Gets teachers by title.                   | Required       |
| `GET`    | `/status/:status`     | Gets teachers by status.                  | Required       |
| `GET`    | `/search`             | Searches for teachers (`?q=...`).         | Required       |
| `GET`    | `/active`             | Gets all active teachers.                 | Required       |

---

## 7. Affair Management Service (`/affairs`)

Manages affairs (事务) for which credit can be applied. All endpoints are prefixed with `/api/affairs`.

### Affair Model

| Field        | Type    | Description                |
| ------------| ------- | --------------------------|
| id          | string  | Affair UUID                |
| name        | string  | Affair name                |
| description | string  | Affair description         |
| creator_id  | string  | Creator's user UUID        |
| attachments | string  | JSON string for attachments|

### AffairStudent Model

| Field      | Type   | Description                |
| ----------| ------ | --------------------------|
| affair_id | string | Affair UUID                |
| student_id| string | Student UUID               |
| is_primary| bool   | Is main responsible        |

### Endpoints

| Method   | Endpoint                        | Description                                 | Authentication |
| :------- | :------------------------------ | :------------------------------------------ | :------------- |
| `POST`   | ``                              | Create a new affair (with participants, etc, will auto-create applications) | Required       |
| `GET`    | `/:id`                          | Get a single affair (with participants)     | Required       |
| `PUT`    | `/:id`                          | Update an affair (creator only)             | Creator Only   |
| `DELETE` | `/:id`                          | Delete an affair                            | Creator/Admin  |
| `GET`    | ``                              | Get a list of all affairs                   | Required       |
| `GET`    | `/:id/participants`             | Get all participants of an affair           | Required       |
| `GET`    | `/:id/applications`             | Get all applications under an affair        | Required       |

#### Create Affair Example

```json
POST /api/affairs
{
  "name": "创新创业项目",
  "description": "2024年创新创业大赛",
  "creator_id": "50685abe-8f89-4149-a245-020b8b32ffcb",
  "participants": ["50685abe-8f89-4149-a245-020b8b32ffcb", "9b4e548b-fe91-4769-8f72-8ae7e8954169"],
  "attachments": "[{\"name\":\"附件1.pdf\",\"url\":\"/uploads/1.pdf\"}]"
}
```

#### Get Affair with Participants Example

```json
GET /api/affairs/45dea375-7e7f-4ed4-90db-9d1385dedf7e
{
  "affair": {
    "id": "45dea375-7e7f-4ed4-90db-9d1385dedf7e",
    "name": "创新创业项目",
    "description": "2024年创新创业大赛",
    "creator_id": "50685abe-8f89-4149-a245-020b8b32ffcb",
    "attachments": "[...]"
  },
  "participants": [
    { "affair_id": "45dea375-7e7f-4ed4-90db-9d1385dedf7e", "student_id": "50685abe-8f89-4149-a245-020b8b32ffcb", "is_primary": true },
    { "affair_id": "45dea375-7e7f-4ed4-90db-9d1385dedf7e", "student_id": "9b4e548b-fe91-4769-8f72-8ae7e8954169", "is_primary": false }
  ]
}
```

#### Get Affair Applications Example

```json
GET /api/affairs/45dea375-7e7f-4ed4-90db-9d1385dedf7e/applications
{
  "applications": [
    { "id": 1, "affair_id": "45dea375-7e7f-4ed4-90db-9d1385dedf7e", "student_number": "50685abe-8f89-4149-a245-020b8b32ffcb", ... },
    { "id": 2, "affair_id": "45dea375-7e7f-4ed4-90db-9d1385dedf7e", "student_number": "9b4e548b-fe91-4769-8f72-8ae7e8954169", ... }
  ],
  "total": 2
}
```

---

## 8. Application Management Service (`/applications`)

Manages credit applications with support for five different credit types. All endpoints are prefixed with `/api/applications` and require authentication.

### Application Model

| Field            | Type      | Description                    |
| ----------------| --------- | ------------------------------ |
| id              | uint      | Application ID                 |
| affair_id       | string    | Associated affair UUID         |
| student_number  | string    | Student UUID                   |
| submission_time | time.Time | Submission timestamp           |
| status          | string    | Status: unsubmitted/pending/approved/rejected |
| reviewer_id     | string    | Reviewer UUID                  |
| review_comment  | string    | Review comments                |
| applied_credits | float64   | Applied credits                |
| approved_credits| float64   | Approved credits               |
| review_time     | *time.Time| Review timestamp               |

### Five Credit Types

#### 1. Innovation Practice Credit (创新创业实践活动学分)
| Field           | Type   | Description     |
| ---------------| ------ | --------------- |
| internship     | string | 实习单位        |
| project_id     | string | 项目编号        |
| certifying_body| string | 认证机构        |
| date           | string | 实践日期        |
| hours          | int    | 实践时长        |

#### 2. Discipline Competition Credit (学科竞赛学分)
| Field    | Type   | Description     |
| -------- | ------ | --------------- |
| level    | string | 竞赛级别        |
| name     | string | 竞赛名称        |
| award    | string | 获奖等级        |
| ranking  | int    | 排名            |

#### 3. Student Entrepreneurship Project Credit (大学生创业项目学分)
| Field        | Type   | Description     |
| ------------ | ------ | --------------- |
| project_name | string | 项目名称        |
| project_level| string | 项目级别        |
| project_rank | int    | 项目排名        |

#### 4. Entrepreneurship Practice Credit (创业实践项目学分)
| Field      | Type    | Description     |
| ---------- | ------- | --------------- |
| company_name| string  | 公司名称        |
| company_rep| string   | 公司代表        |
| share_ratio| float64 | 持股比例        |

#### 5. Paper Patent Credit (论文专利学分)
| Field   | Type   | Description     |
| ------- | ------ | --------------- |
| name    | string | 论文/专利名称   |
| category| string | 类别            |
| ranking | int    | 排名/影响因子   |

### Endpoints

| Method | Endpoint                        | Description                           | User Role      |
| :----- | :------------------------------ | :------------------------------------ | :------------- |
| `POST` | ``                              | Create a single application           | Student        |
| `POST` | `/batch`                        | Batch create applications for affair  | System         |
| `GET`  | `/:id`                          | Get application detail                | Owner/Admin    |
| `GET`  | `/:id/detail`                   | Get full application with credit details | Owner/Admin |
| `PUT`  | `/:id/details`                  | Update application details            | Owner Only     |
| `POST` | `/:id/submit`                   | Submit application for review         | Owner Only     |
| `PUT`  | `/:id/status`                   | Update application status (review)    | Admin/Reviewer |
| `GET`  | `/user/:studentNumber`          | Get user's applications               | Owner/Admin    |
| `GET`  | ``                              | Get all applications                  | Admin/Reviewer |

### Application Workflow

1. **Create Affair** → Auto-generate applications (status: unsubmitted)
2. **Edit Application** → Students fill in credit details
3. **Submit Application** → Status changes to pending
4. **Review Application** → Teacher reviews and approves/rejects

### Permission Control

- **Students**: Can only edit their own applications
- **Teachers/Admins**: Can view and review all applications
- **Affair Creator**: Can edit affair information

### Example Requests

#### Batch Create Applications
```json
POST /api/applications/batch
{
  "affair_id": "45dea375-7e7f-4ed4-90db-9d1385dedf7e",
  "creator_id": "50685abe-8f89-4149-a245-020b8b32ffcb",
  "participants": ["50685abe-8f89-4149-a245-020b8b32ffcb", "9b4e548b-fe91-4769-8f72-8ae7e8954169", "45dea375-7e7f-4ed4-90db-9d1385dedf7e"]
}
```

#### Update Application Details
```json
PUT /api/applications/1/details
{
  "applied_credits": 2.0,
  "details": {
    "level": "国家级",
    "name": "全国大学生数学竞赛",
    "award": "一等奖",
    "ranking": 1
  }
}
```

#### Submit Application
```json
POST /api/applications/1/submit
```

#### Review Application
```json
PUT /api/applications/1/status
{
  "status": "approved",
  "review_comment": "Materials complete, requirements met",
  "approved_credits": 2.0
}
```

---

## 9. Permission Management Service (`/permissions`)

Handles roles, permissions, and their assignments. All endpoints are prefixed with `/api/permissions`.

### Endpoints

| Method   | Endpoint             | Description           |
| :------- | :------------------- | :-------------------- |
| `POST`   | `/init`              | Initialize permissions and roles (no auth required) |
| `POST`   | `/roles`             | Creates a new role.   |
| `GET`    | `/roles`             | Gets all roles.       |
| `GET`    | `/roles/:roleID`     | Gets a specific role. |
| `PUT`    | `/roles/:roleID`     | Updates a role.       |
| `DELETE` | `/roles/:roleID`     | Deletes a role.       |
| `POST`   | ``                   | Creates a permission. |
| `GET`    | ``                   | Gets all permissions. |
| `GET`    | `/:id`               | Gets a single permission.|
| `DELETE` | `/:id`               | Deletes a permission. |
| `POST`   | `/users/:userID/roles` | Assigns a role to a user. |
| `DELETE` | `/users/:userID/roles/:roleID` | Removes a role from a user. |
| `POST`   | `/users/:userID/permissions` | Assigns a permission to a user. |
| `DELETE` | `/users/:userID/permissions/:permissionID` | Removes a permission from a user. |
| `POST`   | `/roles/:roleID/permissions` | Assigns a permission to a role. |
| `DELETE` | `/roles/:roleID/permissions/:permissionID` | Removes a permission from a role. |
| `GET`    | `/users/:userID/roles` | Gets a user's roles. |
| `GET`    | `/users/:userID/permissions` | Gets a user's permissions. |

#### Initialize Permissions Example
```json
POST /api/permissions/init
Response: { "message": "Permissions initialized successfully" }
```