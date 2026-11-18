# PHASE 3 — AUTHENTICATION SECURITY VALIDATION GATE
## Focused Application Security Audit & Verification Report

---

**Application:** Mindora — Student Productivity Management Application  
**Auditor Role:** Senior Application Security Engineer & Software Architect  
**Audit Scope:** Mindora Phase 3 (Authentication Subsystem, Email Auth, Google OAuth 2.0 / OIDC, JWT & Session Security)  
**Target Document:** `docs/rbac/12-phase-3-security-validation.md`  
**Date of Audit:** November 17, 2026  
**Final Classification:** **PASS WITH LIMITATIONS**  

---

## 1. Executive Summary

This independent security audit evaluates the technical implementation, cryptographic integrity, and security claims of the **Mindora Phase 3 Authentication Subsystem**. 

The audit involved an exhaustive code-level inspection of the transport handlers, business services, database repositories, cryptographic packages, and third-party integration modules in `Backend/`. In addition, empirical concurrency and race-condition tests were executed against the PostgreSQL 18 database engine.

The core authentication subsystem is well-structured, follows the approved Standard Go Project Layout, properly isolates concerns across layered packages, enforces bcrypt password hashing, verifies Google OIDC ID token signatures against Google JWKS public keys, and validates account lifecycle statuses (`active`, `inactive`, `suspended`).

However, the audit identified **four critical/high-severity security gaps and edge cases** that must be formally documented and addressed:
1. **Missing OAuth 2.0 PKCE (RFC 7636):** The Google OAuth authorization code flow lacks Proof Key for Code Exchange.
2. **False-Positive Session Invalidation under Concurrent Refresh (Race Condition):** Parallel requests with the same refresh token trigger reuse detection and immediately revoke the user's active session family.
3. **JWT Access Token Exfiltration Risk via OAuth Redirect Query Parameters:** The Google callback redirects to the frontend with the raw access token in URL query parameters (`?token=...`).
4. **Turnstile Production Fail-Open & Test Token Bypass:** Turnstile can be silently disabled in production if unconfigured, and hardcoded test bypass tokens are accepted unconditionally.

---

## 2. HttpOnly Cookie and XSS Protection Audit

### 2.1 Refresh Token Cookie Configuration

The refresh token cookie is configured in [auth_handler.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/handlers/auth_handler.go#L398-L427):

```go
func (h *AuthHandler) setRefreshTokenCookie(c *fiber.Ctx, rawRefreshToken string) {
    c.Cookie(&fiber.Cookie{
        Name:     "refresh_token",
        Value:    rawRefreshToken,
        Path:     "/api/v1/auth",
        Domain:   h.cfg.Cookie.Domain,
        MaxAge:   int(h.cfg.JWT.RefreshExpiry.Seconds()),
        Expires:  time.Now().Add(h.cfg.JWT.RefreshExpiry),
        Secure:   h.cfg.Cookie.Secure,
        HTTPOnly: true,
        SameSite: h.cfg.Cookie.SameSite,
    })
}
```

#### Inspection Results:
- **`HTTPOnly: true`**: Successfully prevents client-side JavaScript access via `document.cookie`.
- **`Secure`**: Tied to `h.cfg.Cookie.Secure`, which defaults to `true` when `APP_ENV=production` ([config.go:138](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/config/config.go#L138)).
- **`SameSite`**: Defaults to `"Lax"` ([config.go:139](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/config/config.go#L139)).
- **`Path: "/api/v1/auth"`**: Properly isolates the cookie to authentication routes, preventing it from being transmitted on regular API requests (`/api/v1/tasks`, `/api/v1/users`, etc.).
- **`Domain`**: Defaults to empty string, binding the cookie strictly to the origin host without subdomain exposure.

### 2.2 Access Token Delivery & Frontend Storage Assumptions

- **Delivery Mechanism:** Access tokens are delivered within the JSON payload of registration, login, and token refresh responses.
- **Frontend Storage Risk:** The frontend application is expected to store the JWT access token in memory (e.g., React component state or in-memory closure). If the frontend persists the access token in `localStorage` or `sessionStorage`, any Cross-Site Scripting (XSS) vulnerability will result in instantaneous token extraction.
- **Exfiltration via URL Query Parameter:** In `GoogleCallback` ([auth_handler.go:384](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/handlers/auth_handler.go#L384)):
  ```go
  if h.cfg.FrontendURL != "" && strings.Contains(c.Get("Accept"), "text/html") {
      redirectURL := h.cfg.FrontendURL + "/login?token=" + result.AccessToken
      return c.Redirect(redirectURL, http.StatusTemporaryRedirect)
  }
  ```
  Passing JWT access tokens in URL query strings violates security best practices:
  - URLs are preserved in browser history.
  - Query parameters are logged in HTTP server access logs and intermediate proxy logs.
  - The `Referer` header may leak the access token to external third-party assets loaded on the `/login` page.

### 2.3 CSRF Protection for Refresh and Logout

- `POST /api/v1/auth/refresh` and `POST /api/v1/auth/logout` rely on `SameSite=Lax`.
- Modern browsers do not attach `SameSite=Lax` cookies to cross-origin POST requests initiated via `fetch()`, `XMLHttpRequest`, or `<form>` submissions, mitigating standard CSRF.
- **Limitation:** There is no explicit CSRF synchronization token (Double Submit Cookie or `X-CSRF-Token`). In addition, `auth_handler.go` supports fallback extraction from the JSON body and Authorization header, bypassing cookie security entirely if an attacker controls request payload construction.

### 2.4 Analysis of HttpOnly XSS Protection Claims

> [!WARNING]
> **Security Reality:** HttpOnly cookies **DO NOT** completely eliminate XSS token theft or session compromise.

While `HttpOnly` stops attackers from executing `alert(document.cookie)` or directly reading the refresh token:
1. **Direct Access Token Theft:** Any XSS payload executing on the frontend origin can steal the active JWT Access Token from in-memory variables or HTTP client interceptors.
2. **Client-Side Request Forgery / Session Riding:** An XSS payload can issue an asynchronous background request:
   ```javascript
   fetch('/api/v1/auth/refresh', { method: 'POST', credentials: 'include' })
     .then(r => r.json())
     .then(data => sendToAttacker(data.data.access_token));
   ```
   Because the request originates from the legitimate origin, the browser attaches the `HttpOnly` cookie automatically. The response containing the newly minted access token is delivered directly to the attacker's script.

### 2.5 Recommended CSP and XSS Mitigations

To effectively mitigate XSS threats, the application must deploy a Content Security Policy (CSP) and defense-in-depth headers:

```http
Content-Security-Policy: default-src 'self'; script-src 'self' https://challenges.cloudflare.com; frame-src https://challenges.cloudflare.com; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' https://challenges.cloudflare.com; object-src 'none'; base-uri 'self'; form-action 'self';
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: strict-origin-when-cross-origin
```

---

## 3. Google OAuth PKCE Verification Audit

### 3.1 PKCE Implementation Inspection

A line-by-line inspection of [google.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/pkg/oauth/google.go) and [auth_handler.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/handlers/auth_handler.go#L221-L396) revealed:

| PKCE Component | Expected Standard (RFC 7636) | Current Implementation | Status |
|---|---|---|---|
| `code_verifier` Generation | Cryptographically random string (43–128 chars, unreserved charset) | **Not implemented** | ❌ Missing |
| `code_challenge` Generation | `BASE64URL-ENCODE(SHA256(code_verifier))` | **Not implemented** | ❌ Missing |
| `code_challenge_method` | `S256` parameter in Google Auth URL | **Not implemented** | ❌ Missing |
| Verifier Persistence | Stored in temporary HttpOnly cookie or encrypted session | **Not implemented** | ❌ Missing |
| Verifier Transmission | Supplied to Google during `/token` exchange | **Not implemented** | ❌ Missing |
| `state` Parameter | High-entropy random string verified on callback | Implemented (32-byte URL-safe base64, HttpOnly cookie) | ✅ Verified |
| `nonce` Parameter | High-entropy random string verified in ID token | Implemented (32-byte URL-safe base64, HttpOnly cookie) | ✅ Verified |
| Cookie Expiration | Short-lived OAuth transaction state | 10 minutes lifetime | ✅ Verified |

### 3.2 Security Risk Evaluation

> [!IMPORTANT]
> **State and Nonce DO NOT Replace PKCE.**
> - `state` protects against Cross-Site Request Forgery (CSRF).
> - `nonce` protects against ID token replay attacks.
> - `PKCE` binds the authorization code to the specific client that initiated the flow, preventing **Authorization Code Interception and Injection Attacks**.

According to the **OAuth 2.0 Security Best Current Practice (BCP)** (draft-ietf-oauth-security-topics Section 2.1), PKCE is mandatory for all OAuth clients, including confidential server-side clients.

**Classification:** **SECURITY GAP / HIGH SEVERITY**.

---

## 4. Refresh Token Rotation Concurrency Audit

### 4.1 Transaction Boundaries & Rotation Logic

In [session_repository.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/repositories/session_repository.go#L54-L68):

```go
func (r *GormSessionRepository) RotateSession(ctx context.Context, oldSessionID uuid.UUID, newSession *models.AuthSession) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        now := time.Now().UTC()
        if err := tx.Model(&models.AuthSession{}).
            Where("id = ? AND revoked_at IS NULL", oldSessionID).
            Update("revoked_at", now).Error; err != nil {
            return err
        }
        return tx.Create(newSession).Error
    })
}
```

In [auth_service.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/services/auth_service.go#L305-L310):

```go
if session.RevokedAt != nil {
    _ = s.sessionRepo.RevokeFamily(ctx, session.FamilyID)
    return nil, ErrTokenReuseDetected
}
```

### 4.2 Empirical Concurrency Verification

An automated race-condition audit test was executed against the PostgreSQL database. Two concurrent goroutines submitted the same refresh token simultaneously.

#### Execution Output from PostgreSQL Logs:
```text
=== RUN   TestAuthService_ConcurrentRefreshRace
[UPDATE] "auth_sessions" SET "revoked_at"=... WHERE id = '0e34bdf5-...' AND revoked_at IS NULL  -- rows: 1
[INSERT] "auth_sessions" (new session '9ad1043a-...')                                           -- rows: 1
[SELECT] "auth_sessions" WHERE token_hash = '559edd56...'                                      -- rows: 1 (sees revoked_at != NULL)
[UPDATE] "auth_sessions" SET "revoked_at"=... WHERE family_id = 'e648302c-...' AND revoked_at IS NULL -- rows: 1 (REVOKES '9ad1043a-...')

Call 0 result: err=refresh token reuse detected; all associated sessions have been revoked, res=false
Call 1 result: err=<nil>, res=true
Number of unrevoked active sessions: 0
--- PASS
```

### 4.3 Vulnerability Analysis

1. **Denial-of-Service / False Positive Session Invalidation:**
   When multiple asynchronous HTTP queries fire simultaneously from a single browser instance (e.g. React Query loading user profile, notifications, and settings in parallel right as the JWT expires), both requests trigger `/api/v1/auth/refresh`:
   - Call 1 succeeds and rotates the session.
   - Call 0 executes a few milliseconds later with the same token, encounters `session.RevokedAt != nil`, and invokes `RevokeFamily()`.
   - **Result:** Call 0 destroys Call 1's newly issued session. All sessions in the user's family are revoked, abruptly logging out the legitimate user.
2. **Missing `RowsAffected` Check:**
   In `RotateSession()`, `tx.Model().Update()` does not check if `RowsAffected == 0`. In GORM, an update matching zero rows returns `err = nil`. If two transactions evaluate the check concurrently before committing, both could insert child sessions without returning an error.

**Classification:** **HIGH SEVERITY (Functional Defect & Availability Risk)**.

---

## 5. Cloudflare Turnstile Production Safety Audit

### 5.1 Verification Logic & Configuration

In [turnstile.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/pkg/turnstile/turnstile.go#L80-L94):

```go
func (c *Client) Verify(ctx context.Context, token, remoteIP, action string) error {
    if !c.cfg.Enabled {
        return nil
    }

    if strings.TrimSpace(token) == "" {
        return ErrTokenRequired
    }

    if token == "1x0000000000000000000000000000000AA" || token == "dummy-test-token" {
        return nil
    }
    // ...
}
```

In [config.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/config/config.go#L131-L132):

```go
turnstileSecretKey := os.Getenv("TURNSTILE_SECRET_KEY")
turnstileEnabled := getEnvAsBool("TURNSTILE_ENABLED", turnstileSecretKey != "")
```

### 5.2 Findings

1. **Silent Production Disablement (Fail-Open):**
   If `TURNSTILE_SECRET_KEY` is omitted from the deployment environment and `TURNSTILE_ENABLED` is left unset, `turnstileEnabled` defaults to `false`. When disabled, `Verify()` immediately returns `nil`. There is no startup validation ensuring that `APP_ENV=production` enforces `TURNSTILE_ENABLED=true`.
2. **Hardcoded Test Bypass in Production Code:**
   Lines 91–93 allow `"dummy-test-token"` and `"1x0000000000000000000000000000000AA"` to bypass verification unconditionally. Because there is no check for `APP_ENV != "production"`, an attacker in production could bypass CAPTCHA challenges by submitting `"dummy-test-token"`.
3. **Fail-Closed on Network Failure:**
   If Cloudflare is unreachable or returns HTTP 5xx, `Verify()` returns an error, preventing unauthenticated bots from registering during Cloudflare outages.
4. **Mock vs. Actual Cloudflare Verification:**
   All Phase 3 automated test suites used `turnstile.Config{Enabled: false}` or local HTTP test servers. Real Cloudflare API validation has not been performed in a staging/production network.

**Classification:** **MEDIUM SEVERITY**.

---

## 6. Authentication Rate Limiting Audit

### 6.1 Route Coverage Inspection

In [routes.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/routes/routes.go#L70-L83):

```go
auth := v1.Group("/auth")
authLimiterStrict := middleware.AuthRateLimiter(10, 1*time.Minute)
authLimiterNormal := middleware.AuthRateLimiter(30, 1*time.Minute)

auth.Post("/register", authLimiterStrict, authHandler.Register)
auth.Post("/login", authLimiterStrict, authHandler.Login)
auth.Get("/google", authLimiterNormal, authHandler.GoogleAuth)
auth.Get("/google/callback", authLimiterNormal, authHandler.GoogleCallback)
auth.Post("/refresh", authLimiterNormal, authHandler.Refresh)
auth.Post("/logout", authHandler.Logout) // Unprotected
```

#### Coverage Summary:
- `/register`: 10 req/min (Strict) — **Protected**
- `/login`: 10 req/min (Strict) — **Protected**
- `/google`: 30 req/min (Normal) — **Protected**
- `/google/callback`: 30 req/min (Normal) — **Protected**
- `/refresh`: 30 req/min (Normal) — **Protected**
- `/logout`: **Unprotected** (No rate limiter applied)

### 6.2 Trusted Proxy IP Handling & Multi-Instance Limitations

In [main.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/cmd/api/main.go#L34-L36):
```go
app := fiber.New(fiber.Config{
    AppName: "Mindora API",
})
```

In [limiter.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/middleware/limiter.go#L23-L25):
```go
KeyGenerator: func(c *fiber.Ctx) string {
    return c.IP()
}
```

1. **Missing Trusted Proxy Configuration:**
   Fiber is initialized without `EnableTrustedProxyCheck: true` or `TrustedProxies`. When deployed behind reverse proxies (Nginx, Cloudflare, AWS ALB), `c.IP()` returns the reverse proxy's private IP. Consequently, **all external users share a single rate-limiting bucket**, causing rapid denial-of-service for legitimate clients.
2. **In-Memory Store:**
   The rate limiter uses an in-memory map. When scaled horizontally across multiple container replicas, rate limit counters are not synchronized, allowing attackers to multiply allowable attempts by the replica count.

**Classification:** **MEDIUM SEVERITY**.

---

## 7. Live Google OAuth Readiness Audit

### 7.1 Configuration & Implementation State

- **Client ID & Secret:** Read from `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`.
- **Redirect URL:** `GOOGLE_REDIRECT_URL` (defaults to `http://localhost:8080/api/v1/auth/google/callback`).
- **Signature Verification:** Uses Google's official `idtoken.Validate()` library verifying signatures against Google JWKS.
- **Account Linking:** Strict `ErrEmailConflictExternalIdentity` prevents unauthorized account hijacking.

### 7.2 Readiness Classification

> [!NOTE]
> The Google OAuth backend code is functionally complete and compiles with unit test mocks. However:
> - No real Google Cloud Console credentials have been provisioned in `.env`.
> - No live browser authorization code exchange has been conducted against live Google accounts.
> 
> Therefore, Live Google OAuth readiness is classified as **PENDING EXTERNAL CREDENTIALS & LIVE VERIFICATION**.

---

## 8. Consolidated Security Findings Matrix

| Finding ID | Finding Description | Severity | Relevant File & Function | Current Behavior | Recommended Remediation | Mandatory Before Phase 4? |
|---|---|---|---|---|---|:---:|
| **SEC-01** | Missing OAuth 2.0 PKCE in Google Flow | **High** | [google.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/pkg/oauth/google.go), [auth_handler.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/handlers/auth_handler.go#L278) | Only `state` and `nonce` are generated; no `code_verifier` or `code_challenge` | Implement RFC 7636 S256 PKCE with temporary HttpOnly cookie storage | **No** *(Mandatory before Phase 6 Frontend)* |
| **SEC-02** | Concurrency Race Condition in Token Rotation (False Positive Reuse Detection) | **High** | [auth_service.go:305](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/services/auth_service.go#L305), [session_repository.go:55](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/repositories/session_repository.go#L55) | Concurrent refresh requests revoke the entire active family, logging out legitimate users | Introduce a 15–30s rotation grace period (leeway window) and assert `RowsAffected == 1` | **No** *(Mandatory before Phase 6 Frontend)* |
| **SEC-03** | Access Token Leaked in URL Query String during OAuth Redirect | **High** | [auth_handler.go:384](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/handlers/auth_handler.go#L384) | Redirects browser to `/login?token=<jwt>` | Use one-time authorization code exchange or redirect to frontend refresh trigger | **No** *(Mandatory before Phase 6 Frontend)* |
| **SEC-04** | Silent Disabling of Turnstile in Production | **Medium** | [config.go:132](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/config/config.go#L132), [turnstile.go:81](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/pkg/turnstile/turnstile.go#L81) | If secret key is unset, Turnstile fails open without warning | Enforce startup assertion in `main.go`/`config.go` requiring Turnstile in production | **No** *(Mandatory before Prod)* |
| **SEC-05** | Test Bypass Token Accepted in Production Code | **Medium** | [turnstile.go:91](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/pkg/turnstile/turnstile.go#L91) | `"dummy-test-token"` bypasses verification unconditionally | Restrict dummy token bypass strictly to `APP_ENV != "production"` | **No** *(Mandatory before Prod)* |
| **SEC-06** | Missing Fiber Trusted Proxy Configuration | **Medium** | [main.go:34](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/cmd/api/main.go#L34), [limiter.go:23](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/middleware/limiter.go#L23) | `c.IP()` returns proxy IP; all clients share single rate limit bucket | Configure `EnableTrustedProxyCheck: true` and specify `TrustedProxies` | **No** *(Mandatory before Deployment)* |
| **SEC-07** | Logout Endpoint Lacks Rate Limiting | **Low** | [routes.go:82](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/routes/routes.go#L82) | `POST /auth/logout` has no rate limiter | Attach `authLimiterNormal` middleware | **No** *(Low risk)* |
| **SEC-08** | In-Memory Limiter Inadequate for Multi-Instance Deployments | **Informational** | [limiter.go:20](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/internal/middleware/limiter.go#L20) | Limits are isolated per server instance | Migrate to Redis-backed storage when scaling horizontally | **No** *(Architecture roadmap)* |
| **SEC-09** | Live Google OAuth Unverified Against Real Cloud Credentials | **Informational** | [google.go](file:///D:/1.%20PROYEK%203/Mindora%20App/Backend/pkg/oauth/google.go) | Tested solely against mock verifiers/exchangers | Provision Google Cloud credentials and perform manual browser verification | **No** *(Awaiting credentials)* |

---

## 9. Final Security Gate Classification

### Overall Audit Verdict: **PASS WITH LIMITATIONS**

### Rationale:
1. **Core Cryptographic & Identity Integrity:** The core Phase 3 implementation successfully implements secure bcrypt password hashing, JWT HS256 access token generation, cryptographically random refresh token generation, SHA-256 token hashing in PostgreSQL, Google OIDC token signature validation, account status enforcement, and email registration collision handling.
2. **Phase 4 Independence:** None of the identified security findings block the implementation of **Phase 4 (RBAC Authorization Middleware & Resource Ownership Scoping)**. Phase 4 concerns validating user roles (`user`, `admin`) and enforcing permissions (`tasks:create`, `tasks:read`, etc.) on incoming requests that already hold a valid JWT access token.
3. **Remediation Timing:**
   - **SEC-01 (PKCE), SEC-02 (Concurrency Grace Period), and SEC-03 (Token in Query Param)** must be remediated prior to or during **Phase 6 (Frontend Integration)** to prevent user session lockouts and token leakage.
   - **SEC-04, SEC-05, and SEC-06** must be remediated prior to staging/production infrastructure deployment.

---

## 10. Next Steps

1. Await explicit user authorization before commencing remediation or transitioning to **Phase 4**.
2. Once authorized, proceed to **Phase 4: RBAC Authorization Middleware, Permission Guards, and Context Utilities**.
