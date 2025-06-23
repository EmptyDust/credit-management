# API Documentation

This document provides a comprehensive and detailed overview of the API endpoints for the Credit Management System. All endpoints are proxied through the API Gateway and are prefixed with `/api`.

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
  - [8. Application Management Service (`/applications`)](#8-application-management-service-applications)

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
| `GET`    | `/:username`         | Gets a specific user's profile.           | Admin Required            |
| `PUT`    | `/:username`         | Updates a specific user's profile.        | Admin Required            |
| `DELETE` | `/:username`         | Deletes a user.                           | Admin Required            |
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
| `GET`    | `/:studentID`     | Gets a student by Student ID.           | Required       |
| `PUT`    | `/:studentID`     | Updates a student's info.               | Required       |
| `DELETE` | `/:studentID`     | Deletes a student.                      | Admin Required |
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
| `GET`    | `/:username`          | Gets a teacher by username.               | Required       |
| `PUT`    | `/:username`          | Updates a teacher's info.                 | Required       |
| `DELETE` | `/:username`          | Deletes a teacher.                        | Admin Required |
| `GET`    | ``                    | Gets a list of all teachers.              | Required       |
| `GET`    | `/department/:department`| Gets teachers by department.           | Required       |
| `GET`    | `/title/:title`       | Gets teachers by title.                   | Required       |
| `GET`    | `/status/:status`     | Gets teachers by status.                  | Required       |
| `GET`    | `/search`             | Searches for teachers (`?q=...`).         | Required       |
| `GET`    | `/active`             | Gets all active teachers.                 | Required       |

---

## 7. Affair Management Service (`/affairs`)

Manages the types of affairs for which credit can be applied. All endpoints are prefixed with `/api/affairs`.

| Method   | Endpoint | Description                 | Authentication |
| :------- | :------- | :-------------------------- | :------------- |
| `POST`   | ``       | Creates a new affair type.  | Admin Required |
| `GET`    | `/:id`   | Gets a single affair type.  | Required       |
| `PUT`    | `/:id`   | Updates an affair type.     | Admin Required |
| `DELETE` | `/:id`   | Deletes an affair type.     | Admin Required |
| `GET`    | ``       | Gets a list of all affairs. | Required       |

---

## 8. Application Management Service (`/applications`)

Manages credit applications. All endpoints are prefixed with `/api/applications` and require authentication.

| Method | Endpoint                | Description                            | User Role      |
| :----- | :---------------------- | :------------------------------------- | :------------- |
| `POST` | ``                      | Creates a new application.             | Student        |
| `GET`  | `/:id`                  | Gets a single application by ID.       | Owner/Admin    |
| `PUT`  | `/:id/status`           | Updates an application's status.       | Admin/Reviewer |
| `GET`  | `/user/:studentNumber`  | Gets all applications for a student.   | Owner/Admin    |