# 08 — Google Cloud Console & OAuth 2.0 Setup Guide

**Protocol:** Google OAuth 2.0 / OpenID Connect  
**Flow:** Authorization Code Flow with Server-Side Verification  
**Target Environment:** Mindora Backend REST API  

---

## 1. Google Cloud Project Configuration

To enable live Google authentication in staging and production environments, configure Google Cloud Console:

### Step 1: Create or Select Google Cloud Project
1. Navigate to [Google Cloud Console](https://console.cloud.google.com/).
2. Create a new project named `Mindora Application` (or select an existing organization project).

### Step 2: Configure OAuth Consent Screen
1. Go to **APIs & Services** > **OAuth consent screen**.
2. Select User Type:
   - **Internal:** If testing within a Google Workspace organization.
   - **External:** For general public student user access.
3. Fill in Application Details:
   - **App name:** `Mindora`
   - **User support email:** Support email address.
   - **Developer contact information:** Security contact.
4. Add Scopes:
   - Click **Add or Remove Scopes**.
   - Select strictly:
     - `.../auth/userinfo.email`
     - `.../auth/userinfo.profile`
     - `openid`
   - *Do not request Google Drive, Gmail, or Calendar scopes.*

### Step 3: Create OAuth 2.0 Client Credentials
1. Go to **APIs & Services** > **Credentials**.
2. Click **Create Credentials** > **OAuth client ID**.
3. Application Type: **Web application**.
4. Name: `Mindora Web API Client`.
5. Authorized JavaScript origins:
   - Development: `http://localhost:5173`, `http://localhost:8080`
   - Production: `https://mindora.app`
6. Authorized redirect URIs:
   - Development: `http://localhost:8080/api/v1/auth/google/callback`
   - Production: `https://api.mindora.app/api/v1/auth/google/callback`
7. Click **Create** and securely copy:
   - `Client ID`
   - `Client Secret`

---

## 2. Environment Variables Configuration

Add the generated credentials to `Backend/.env`:

```env
GOOGLE_CLIENT_ID=your_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
GOOGLE_OAUTH_ENABLED=true
```

> [!WARNING]
> Never commit `GOOGLE_CLIENT_SECRET` to Git or expose it to frontend clients.

---

## 3. Graceful Fallback When Unconfigured

If `GOOGLE_OAUTH_ENABLED=false` or credentials are missing:
- Mindora backend continues running without crash.
- Endpoints `/api/v1/auth/google` and `/api/v1/auth/google/callback` return HTTP 503:
  `"Google OAuth is disabled or not configured in this environment"`.
- Unit and integration tests use the internal `MockGoogleVerifier` and `MockTokenExchanger` to verify 100% of the code path without live credentials.
