# 10 — Token & Session Security Architecture

**Topic:** JWT Claims, Refresh Token Rotation, Cryptographic Storage  
**Compliance Standard:** RFC 7519 (JWT), RFC 6749 (OAuth 2.0), OWASP Session Management  

---

## 1. Access Token Architecture (JWT)

- **Algorithm:** `HMAC-SHA256` (`HS256`).
- **Secret Key:** Injected via `JWT_SECRET` environment variable (minimum 32 characters).
- **Lifespan:** 15 minutes (`JWT_ACCESS_EXPIRY=15m`).
- **Storage:** Frontend in-memory state only. **Never stored in localStorage or sessionStorage**.
- **Claims Structure:**
  ```json
  {
    "sub": "b2c1d3e4-f5a6-4b7c-8d9e-0f1a2b3c4d5e",
    "email": "student@mindora.app",
    "role": "user",
    "iss": "mindora-api",
    "aud": "mindora-client",
    "exp": 1731862500,
    "iat": 1731861600,
    "nbf": 1731861600,
    "jti": "550e8400-e29b-41d4-a716-446655440000"
  }
  ```

---

## 2. Refresh Token Rotation & Reuse Detection

Mindora implements **Refresh Token Rotation (RTR)** with **Family-Wide Revocation**:

```text
Login: Issue (Access_1, Refresh_1, Family_A)
           |
Refresh 1: Present Refresh_1 -> Revoke Refresh_1 -> Issue (Access_2, Refresh_2, Family_A)
           |
Refresh 2: Present Refresh_2 -> Revoke Refresh_2 -> Issue (Access_3, Refresh_3, Family_A)
           |
Attacker:  Replays stolen Refresh_1 (Already Revoked)
           |
           v
REUSE DETECTED!
System executes: UPDATE auth_sessions SET revoked_at = NOW() WHERE family_id = 'Family_A'
All tokens in Family_A (including Access_3 & Refresh_3) are immediately revoked!
Attacker and legitimate user are both forced to re-authenticate.
```

### Storage Security
- **Raw Token:** 32 cryptographically random bytes encoded as a 64-character hex string.
- **Database Storage:** The database only stores `token_hash = SHA256(raw_refresh_token)`. Even with a full database dump, attackers cannot recover plaintext refresh tokens.
- **Cookie Security:**
  - `HttpOnly`: Prevents access via client JavaScript (mitigates XSS).
  - `Secure`: Transmitted only over HTTPS in production.
  - `SameSite=Lax`: Mitigates Cross-Site Request Forgery (CSRF).
  - `Path=/api/v1/auth`: Restricts cookie transmission strictly to auth routes.
