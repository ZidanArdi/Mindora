# 07 — Mindora Authentication Subsystem Architecture

**System:** Mindora Student Productivity Application  
**Subsystem:** Authentication, Identity Providers, Token Security  
**Status:** Implemented, Tested & Verified (Phase 3)  

---

## 1. Architectural Overview

The Mindora Authentication Subsystem provides identity verification, credential handling, session management, and external provider integration. It integrates directly with the Phase 2 RBAC data model and Standard Go Project Layout.

```text
                                  +---------------------------------------+
                                  |            HTTP Request               |
                                  +---------------------------------------+
                                                      |
                                                      v
                                  +---------------------------------------+
                                  |       Rate Limiting Middleware        |
                                  |     (In-Memory IP Bucket Limiter)     |
                                  +---------------------------------------+
                                                      |
                                                      v
                                  +---------------------------------------+
                                  |           Transport Handler           |
                                  |       (internal/handlers/auth)        |
                                  +---------------------------------------+
                                                      |
                                                      v
                                  +---------------------------------------+
                                  |         Authentication Service        |
                                  |       (internal/services/auth)        |
                                  +---------------------------------------+
                                  /            |             \           \
                                 /             |              \           \
                                v              v               v           v
                        +-------------+ +-------------+ +-------------+ +---------------+
                        |   Bcrypt    | |     JWT     | |  Turnstile  | | Google OAuth  |
                        |   Hasher    | |   Manager   | |  Verifier   | |  OIDC Service |
                        +-------------+ +-------------+ +-------------+ +---------------+
                                               |
                                               v
                                  +---------------------------------------+
                                  |          Repositories Layer           |
                                  |   (user, identity, session repos)     |
                                  +---------------------------------------+
                                                      |
                                                      v
                                  +---------------------------------------+
                                  |             PostgreSQL 18             |
                                  | (users, auth_identities, sessions)    |
                                  +---------------------------------------+
```

---

## 2. Authentication Methods & Business Rules

### 2.1 Email & Password Authentication
- **Registration (`POST /api/v1/auth/register`):**
  - Publicly accessible to unauthenticated guests.
  - Automatically and strictly assigns the `user` role on the backend.
  - Rejects any client-supplied `role_id` or role tampering attempts.
  - Hashes passwords using bcrypt (cost factor 12 in production, cost factor 10 in unit tests).
  - Enforces case-insensitive email uniqueness via PostgreSQL functional index `LOWER(email)`.
- **Login (`POST /api/v1/auth/login`):**
  - Authenticates via email and password.
  - Returns generic error messages (`"invalid email or password"`) for both unknown emails and incorrect passwords to prevent user enumeration.
  - Validates account status (`active`, `inactive`, `suspended`).
  - Google-only accounts (where `password_hash` is NULL) cannot log in via password and are returned generic invalid credentials.

### 2.2 Google OAuth 2.0 / OpenID Connect
- **Protocol:** OpenID Connect Authorization Code Flow with backend callback.
- **Identity Scope:** Strictly limited to `openid`, `email`, and `profile`. No Gmail, Drive, or Calendar scopes are requested.
- **ID Token Verification:** Cryptographically validates Google's signature via Google public JWKS, issuer (`accounts.google.com`), audience (`GOOGLE_CLIENT_ID`), expiry, and `email_verified` claim.
- **Account Creation:**
  - If Google `sub` (Subject ID) already exists in `auth_identities`, logs the user in immediately.
  - If Google `sub` is new and email does not exist in `users`, registers a new user with `user` role and links the `auth_identities` record.

---

## 3. Strict No-Account-Linking Policy

Mindora enforces a **strict prohibition against automatic account linking**:

> **Rule:** If a Google OAuth identity arrives with an email that matches an existing Mindora account created via email/password, the system **MUST NOT** silently link the Google identity or log the user in.

- **Reasoning:** Prevents identity hijacking if an attacker creates or compromises a Google account under a victim's secondary unverified email.
- **System Behavior:** Rejects the login attempt with HTTP 409 Conflict:
  `"this email is already registered using another sign-in method; please sign in with your email and password"`.
- Any future account-linking capability must be performed from an authenticated settings panel where the user proves ownership of both credentials.

---

## 4. RBAC Integration & Boundaries

- **Persisted Roles:** Only `user` and `admin`.
- **Guest:** An unauthenticated visitor context; not stored in the database.
- **Role Immutability:** Public registration cannot provision `admin` accounts.
- **Admin Accounts:** Must be pre-provisioned via secure administrative CLI tools.

---

## 5. Distributed Refresh Token Rotation & Idempotency Architecture (Phase 3.6)

To support horizontal scale-out across $\ge 2$ backend service instances sharing a single PostgreSQL database without introducing Redis, Mindora coordinates token rotation at the database level:

### 5.1 PostgreSQL Transactional Advisory Locks
- **Primitive:** `pg_try_advisory_xact_lock(hashtext(tokenHash), hashtext(reverse(tokenHash)))`.
- **64-bit Advisory Key Space:** Uses two 32-bit hashes to eliminate collision risk.
- **Deterministic Contention Probing:** Probes `pg_try_advisory_xact_lock` before blocking on `pg_advisory_xact_lock`. If the lock was not acquired immediately, the service marks `hadLockContention = true`.
- **Automatic Lifecycle:** The lock is strictly transaction-scoped (`_xact_lock`) and automatically released upon `COMMIT` or `ROLLBACK`.

### 5.2 Cryptographic Idempotency & Operation Tracking
- **Operation Identifier:** Clients attach an `X-Operation-ID` header (cryptographically random UUIDv4) to every refresh operation.
- **Payload Encryption:** Successful rotation results are encrypted with AES-256-GCM authenticated encryption using a 32-byte secret key and stored in `auth_sessions.encrypted_response` with a 60-second TTL. Plaintext tokens are never stored.
- **Idempotent Retry:** If a client retries after a network drop using the same operation identifier, the backend decrypts and returns the identical response without minting new tokens or branching families.

### 5.3 Deterministic Concurrency Conflict vs. Reuse Replay
| Request State | Lock Contention | Operation ID Match | System Decision | Action |
| :--- | :---: | :---: | :--- | :--- |
| **Fresh Token** | False | N/A | Valid Rotation | Issues new token, stores encrypted payload, commits. |
| **Rotated Token** | True/False | **Matches** | Idempotent Retry | Returns cached decrypted payload (200 OK). No family change. |
| **Rotated Token** | **True** | Different / Empty | Concurrent Conflict | Returns `ErrConcurrentConflict` (409 Conflict). **Does NOT revoke family.** |
| **Rotated Token** | **False** | Different / Empty | **Sequential Replay** | **Confirmed Attack.** Executes `RevokeFamily()`, terminates all sessions (401 Unauthorized). |

### 5.4 Frontend Request Coordination
- In [`frontend/src/services/api.ts`](file:///D:/1.%20PROYEK%203/Mindora%20App/frontend/src/services/api.ts), parallel API calls encountering HTTP 401 are queued behind a single shared `refreshPromise`.
- Generates a single `X-Operation-ID` per refresh cycle.
- Marks retried requests with `_retry: true` to prevent infinite refresh loops.
- On failure (401), clears in-memory credentials and redirects to `/login`.
