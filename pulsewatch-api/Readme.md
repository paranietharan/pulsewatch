# PulseWatch API (`pulsewatch-api`)

Production-grade Go REST API for the **PulseWatch** uptime monitoring platform.

---

## 1. Authentication Architecture & Token Lifecycle

PulseWatch implements a defense-in-depth authentication model combining short-lived signed JWTs with stateful, rotating opaque refresh tokens:

```
[ Client / Browser ] 
       │
       │  POST /api/v1/auth/login
       ▼
[ PulseWatch API ] ──► Validates bcrypt hash
       │
       ├──► Returns Access Token (JWT, 15m lifetime)
       ├──► Sets HttpOnly Cookie: refresh_token (Opaque, 7d lifetime)
       └──► Stores SHA-256 hash of refresh token in PostgreSQL (with Family ID)
```

### Access Tokens (Stateless JWT)
- **Algorithm**: HMAC-SHA256 (`HS256`).
- **Lifetime**: 15 minutes (`JWTAccessExpiry`).
- **Claims**: `sub` (User UUID), `email`, `role`, `iss` (`pulsewatch-api`), `aud` (`pulsewatch-client`), `exp`, `iat`, `jti`.
- **Transmission**: `Authorization: Bearer <token>` HTTP header.

### Refresh Tokens & Session Lifecycle
- **Type**: 32 cryptographically random bytes (`crypto/rand`), 64-character hex string.
- **Storage**: Only the **SHA-256 hash** of the token is persisted in `refresh_tokens`.
- **Transmission**: Stored in a secure `HttpOnly`, `SameSite=Lax`, `Path=/api/v1/auth` cookie (plus response DTO for native clients).
- **Rotation**: Every call to `POST /api/v1/auth/refresh` revokes the old refresh token and issues a brand new one within the same **Token Family (`family_id`)**.
- **Theft / Reuse Detection**: If an already-revoked refresh token is presented, the system flags a replay attack and **immediately revokes the entire family**, locking out the attacker and forcing re-authentication.

### Transactional Outbox Pattern for Emails
To eliminate the dual-write problem between PostgreSQL and RabbitMQ, user registration and password resets insert both the account/token state and an `auth_outbox_events` record in the **same atomic database transaction**. A dedicated background worker (`outbox.Dispatcher`) streams pending events to RabbitMQ with publisher confirmation before marking them published.

---

## 2. API Endpoints

All authentication endpoints are versioned under `/api/v1/auth`.

### 2.1 User Registration
`POST /api/v1/auth/signup`
- **Rate Limit**: 5 requests / min per IP.
- **Request Body**:
  ```json
  {
    "first_name": "Jane",
    "last_name": "Doe",
    "email": "jane@example.com",
    "password": "StrongPassword123!"
  }
  ```
- **Response** (`201 Created`):
  ```json
  {
    "message": "Account created successfully. Please check your email to verify your account.",
    "user": {
      "id": "c1f7cf8c-...",
      "first_name": "Jane",
      "last_name": "Doe",
      "email": "jane@example.com",
      "role": "USER",
      "is_active": true,
      "is_verified": false,
      "created_at": "2026-10-10T14:35:00Z"
    }
  }
  ```

---

### 2.2 Login
`POST /api/v1/auth/login`
- **Rate Limit**: 5 requests / min per IP.
- **Request Body**:
  ```json
  {
    "email": "jane@example.com",
    "password": "StrongPassword123!"
  }
  ```
- **Response** (`200 OK`):
  ```json
  {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "a1b2c3d4...",
    "expires_in": 900,
    "user": {
      "id": "c1f7cf8c-...",
      "email": "jane@example.com",
      "role": "USER",
      "is_verified": true
    }
  }
  ```
  *(Also sets `Set-Cookie: refresh_token=...; HttpOnly; Path=/api/v1/auth; SameSite=Lax`)*

---

### 2.3 Token Refresh (Rotation)
`POST /api/v1/auth/refresh`
- **Request**: Accepts either `refresh_token` in cookie or JSON body:
  ```json
  {
    "refresh_token": "a1b2c3d4..."
  }
  ```
- **Response** (`200 OK`):
  ```json
  {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "e5f6g7h8...",
    "expires_in": 900,
    "user": { ... }
  }
  ```

---

### 2.4 Logout
`POST /api/v1/auth/logout`
- Revokes the active refresh token session and clears the cookie (`Max-Age=-1`).
- **Response** (`200 OK`):
  ```json
  {
    "message": "Logged out successfully"
  }
  ```

---

### 2.5 Current User Profile
`GET /api/v1/auth/me`
- **Headers**: `Authorization: Bearer <access_token>`
- **Response** (`200 OK`):
  ```json
  {
    "user": {
      "id": "c1f7cf8c-...",
      "first_name": "Jane",
      "last_name": "Doe",
      "email": "jane@example.com",
      "role": "USER",
      "is_active": true,
      "is_verified": true,
      "created_at": "2026-10-10T14:35:00Z"
    }
  }
  ```

---

### 2.6 Email Verification
`POST /api/v1/auth/verify-email`
- **Request Body**:
  ```json
  {
    "token": "4a5b6c7d..."
  }
  ```
- **Response** (`200 OK`):
  ```json
  {
    "message": "Email address verified successfully. Your account is now active."
  }
  ```

`POST /api/v1/auth/resend-verification`
- **Request Body**:
  ```json
  {
    "email": "jane@example.com"
  }
  ```
- **Response** (`200 OK`): Generic response preventing account enumeration.

---

### 2.7 Password Reset Flow
`POST /api/v1/auth/forgot-password`
- **Request Body**:
  ```json
  {
    "email": "jane@example.com"
  }
  ```
- **Response** (`200 OK`): Generic response preventing account enumeration.

`POST /api/v1/auth/reset-password`
- **Request Body**:
  ```json
  {
    "token": "9f8e7d6c...",
    "new_password": "NewStrongPassword456!"
  }
  ```
- **Response** (`200 OK`):
  ```json
  {
    "message": "Password has been reset successfully. You can now log in with your new password."
  }
  ```
  *(Atomically consumes the single-use token and invalidates all active sessions for the user across all devices).*

---

## 3. RabbitMQ Event Contracts

- **Exchange**: `pulsewatch.direct` (durable direct)
- **Queue**: `notifications.email` (durable queue)
- **Routing Keys**:
  - `notification.email.verification.requested`
  - `notification.email.password_reset.requested`
  - `notification.email.resend_verification.requested`

### Event Payload Schema
```json
{
  "event_id": "31a8db8a-0089-4e57-bf92-036b799b0b53",
  "event_type": "notification.email.verification.requested",
  "schema_version": "1.0",
  "occurred_at": "2026-10-10T14:35:02Z",
  "payload": {
    "user_id": "00701001-0d91-40c4-a2da-9cc5898f400f",
    "email": "jane@example.com",
    "name": "Jane Doe",
    "verification_token": "d8e7f6...",
    "expires_at": "2026-10-11T14:35:02Z"
  }
}
```

---

## 4. Running and Testing Locally

### Start Infrastructure
```bash
docker compose up -d postgres rabbitmq
```

### Apply Migrations
```bash
go run pulse-watch-db/main.go up
```

### Run Automated Tests
```bash
cd pulsewatch-api
go test -v ./...
```

### Start API Server
```bash
go run pulsewatch-api/cmd/api/main.go
```

### Start Notification Consumer
```bash
go run pulsewatch-notification-service/main.go
```
