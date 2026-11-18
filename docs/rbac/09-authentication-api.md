# 09 — Mindora Authentication API Contract Specification

**Base URL:** `/api/v1/auth`  
**Protocol:** REST / JSON / HTTP Cookies  
**Response Wrapper:** Standard Mindora `{ "success": bool, "message": string, "data": ... }`  

---

## 1. Public Authentication Endpoints

### 1.1 Register Account
- **Route:** `POST /api/v1/auth/register`
- **Rate Limit:** 10 requests per minute per IP
- **Request Body:**
  ```json
  {
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "SecurePassword123!",
    "turnstile_token": "0x4AAAAAA..."
  }
  ```
- **Responses:**
  - `HTTP 201 Created`:
    ```json
    {
      "success": true,
      "message": "Registration completed successfully",
      "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "token_type": "Bearer",
        "expires_in": 900,
        "user": {
          "id": "c1f7a0b2-3c4d-5e6f-7a8b-9c0d1e2f3a4b",
          "name": "Jane Doe",
          "email": "jane@example.com",
          "role": "user",
          "status": "active",
          "created_at": "2026-11-17T18:00:00Z"
        }
      }
    }
    ```
    *Header:* `Set-Cookie: refresh_token=<random_64_hex>; Path=/api/v1/auth; HttpOnly; SameSite=Lax`
  - `HTTP 409 Conflict`: Email already registered.
  - `HTTP 422 Unprocessable Entity`: Validation failure (weak password, invalid email).
  - `HTTP 403 Forbidden`: Turnstile bot protection check failed.

---

### 1.2 User Login
- **Route:** `POST /api/v1/auth/login`
- **Rate Limit:** 10 requests per minute per IP
- **Request Body:**
  ```json
  {
    "email": "jane@example.com",
    "password": "SecurePassword123!",
    "turnstile_token": "0x4AAAAAA..."
  }
  ```
- **Responses:**
  - `HTTP 200 OK`: Access token payload + `Set-Cookie: refresh_token=...; HttpOnly`
  - `HTTP 401 Unauthorized`: `"Invalid email or password"` (generic).
  - `HTTP 403 Forbidden`: Account is inactive or suspended.

---

### 1.3 Google OAuth Initiation
- **Route:** `GET /api/v1/auth/google`
- **Query Parameters:** `?turnstile_token=...` (optional), `?redirect=false` (optional, to return JSON)
- **Behavior:**
  - Generates secure `state` and `nonce` stored in temporary 10-minute HttpOnly cookies.
  - Default: `HTTP 307 Temporary Redirect` to Google consent screen.
  - If `redirect=false`: Returns `HTTP 200 OK` with JSON `{ "auth_url": "https://accounts.google.com/..." }`.

---

### 1.4 Google OAuth Callback
- **Route:** `GET /api/v1/auth/google/callback`
- **Query Parameters:** `?code=...&state=...`
- **Responses:**
  - `HTTP 200 OK` or `HTTP 307 Redirect` to frontend with `?token=<access_token>`.
  - `HTTP 409 Conflict`: When Google email belongs to an existing password user without prior linking.
  - `HTTP 400 Bad Request`: Invalid state, bad code, or unverified Google email.

---

## 2. Session Endpoints

### 2.1 Refresh Token Rotation
- **Route:** `POST /api/v1/auth/refresh`
- **Authentication:** Delivered via `refresh_token` HttpOnly cookie (or JSON body `{ "refresh_token": "..." }`).
- **Behavior:**
  - Atomically marks previous refresh token revoked.
  - Generates replacement refresh token under the same `family_id`.
  - Sets updated `refresh_token` cookie.
- **Responses:**
  - `HTTP 200 OK`: New access token and rotated refresh token cookie.
  - `HTTP 401 Unauthorized`:
    - Expired/invalid refresh token.
    - **Token Reuse Detected:** Revokes entire session family immediately!

---

### 2.2 Logout
- **Route:** `POST /api/v1/auth/logout`
- **Behavior:** Revokes the session associated with the active refresh token and deletes the `refresh_token` cookie (`Max-Age=-1`).
- **Response:** `HTTP 200 OK`: `{ "success": true, "message": "Logged out successfully" }`.
