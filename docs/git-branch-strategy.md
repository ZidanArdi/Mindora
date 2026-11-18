# Mindora Source Control Strategy — Git Branch Separation

**Role:** Senior DevOps Engineer & Git Repository Architect  
**Project:** Mindora (Student Productivity Application)  
**Status:** IMPLEMENTED & VERIFIED  
**Date:** November 18, 2026  
**Repository Pattern:** Monorepo with Domain-Isolated Development Branches  

---

## 1. Executive Summary

This document establishes the official Git branching and source control architecture for the Mindora application monorepo. To enable autonomous, conflict-free development across distinct teams while maintaining a unified repository root, the repository has been structured into two dedicated development branches:

1. **`frontend`**: Exclusively responsible for `Frontend/` application code, user interfaces, frontend tests, and frontend documentation. The `Backend/` directory remains strictly locked to the common baseline snapshot.
2. **`backend`**: Exclusively responsible for `Backend/` Go services, database migrations, security policies, backend unit/integration tests, and backend documentation. The `Frontend/` directory remains strictly locked to the common baseline snapshot.

Both branches preserve the complete directory tree structure (`Frontend/` and `Backend/`) so that workspace tools, build scripts, linters, and path references continue to operate seamlessly. Neither branch pollutes the other domain's code.

---

## 2. Repository Architecture & Lineage

### 2.1 Common Baseline Commit

A thorough audit of the Git history identified commit `36a179b` as the common ancestor and baseline for both branches:

```
Commit Hash : 36a179b1897e06574944faee2ec74971c08d9be9 (short: 36a179b)
Author      : Senior Frontend Engineer <frontend@mindora.app>
Date        : Sat Mar 1 2026
Message     : fix(frontend): update web favicon and branding with official Mindora logo
```

**Why this commit was chosen as the baseline:**
- In `36a179b`, the frontend contained the complete, approved atomic design system and branding assets before Phase 1 RBAC modifications began.
- In `36a179b`, the backend code already conformed to the Standard Go Project Layout (`Backend/cmd/`, `Backend/internal/`, `Backend/pkg/`).
- Prior to this commit (specifically commit `58bd9c8`), an incompatible flat directory structure had broken Go package builds. Commit `36a179b` successfully rectified this, serving as the stable common ancestor.

### 2.2 Branch Topology & Tree Representation

```
* 2b0fa08 (frontend) docs: add git branch separation and source control strategy
* 82016dc feat(frontend): implement Mindora authentication interface
|   [Frontend/ active, Backend/ frozen at 36a179b]
|
| * f1d3211 (backend) docs: add git branch separation and source control strategy
| * 5fdaac7 feat(backend): implement Mindora authentication and RBAC
|/    [Backend/ active, Frontend/ frozen at 36a179b]
|
* 36a179b (origin/frontend) fix(frontend): update web favicon and branding with official Mindora logo
|   [Common Stable Baseline]
|
* 88c1b15 feat(frontend): implement nature-inspired redesign, atomic design components
* f87fc4b refactor(frontend): restructure components using atomic design methodology
|
| * 58bd9c8 (origin/backend, legacy/backend-flat) refactor(backend): restructure directory into modular architecture
|/    [Preserved legacy pointer]
|
* bdf3885 (main, origin/main) Initial commit: Initialize Mindora project
```

In addition, an immutable backup branch has been created to preserve the unified Phase 4 state:
- **`backup/phase-4-baseline`** (`71a6786`): Full working snapshot capturing both frontend and backend Phase 4 implementations prior to domain separation.

---

## 3. Dedicated Branch Allocation & Isolation Proof

### 3.1 Branch `frontend`

- **Branch Head:** `2b0fa08` (feature commit: `82016dc`)
- **Parent Commit:** `36a179b`
- **Ownership Scope:**
  - `Frontend/**` (React 18 + Vite + TypeScript + Tailwind CSS)
  - `docs/rbac/**` (shared architecture and frontend integration specifications)
  - `.gitignore` (standardized secret and artifact exclusion rules)
- **Isolation Verification:**
  ```powershell
  git diff 36a179b Backend/
  # Output: 0 bytes (zero modifications to Backend directory)
  ```
- **Key Changes Implemented:**
  - Resolved blank white screen at startup by standardizing `useAuth` hook exports.
  - Implemented real authentication integration (`LoginForm.tsx`, `RegisterForm.tsx`).
  - Added feature flags for graceful fallback: `VITE_ENABLE_GOOGLE_OAUTH=false`, `VITE_ENABLE_TURNSTILE=false`.
  - Configured Axios interceptors with request queueing for refresh token rotation.
  - Passed `npm run build` (3.06s) and `npm run lint` (0 errors, 0 warnings).

### 3.2 Branch `backend`

- **Branch Head:** `f1d3211` (feature commit: `5fdaac7`)
- **Parent Commit:** `36a179b`
- **Ownership Scope:**
  - `Backend/**` (Go 1.23+ Fiber v2, GORM, PostgreSQL 18.6)
  - `docs/rbac/**` (RBAC matrix, security remediation, distributed lock architecture)
  - `.gitignore` (standardized secret and artifact exclusion rules)
- **Isolation Verification:**
  ```powershell
  git diff 36a179b Frontend/
  # Output: 0 bytes (zero modifications to Frontend directory)
  ```
- **Key Changes Implemented:**
  - Robust Role-Based Access Control (RBAC) middleware and permission evaluator.
  - Refresh token rotation with atomic database updates and PostgreSQL transactional advisory locks (`pg_advisory_xact_lock`).
  - Singleflight concurrency deduplication in Go auth service.
  - Strict reuse detection and session family revocation (`RevokeFamily`).
  - 100% test pass rate across all 11 backend packages (`go test -count=1 ./...`).

---

## 4. File Allocation Matrix

| Path / Directory | `frontend` Branch Status | `backend` Branch Status | Policy |
| :--- | :--- | :--- | :--- |
| `Frontend/src/**` | **Active / Modified** | **Frozen** (at `36a179b`) | Domain-isolated |
| `Frontend/package.json` | **Active / Modified** | **Frozen** (at `36a179b`) | Domain-isolated |
| `Frontend/vite.config.ts` | **Active / Modified** | **Frozen** (at `36a179b`) | Domain-isolated |
| `Backend/cmd/**` | **Frozen** (at `36a179b`) | **Active / Modified** | Domain-isolated |
| `Backend/internal/**` | **Frozen** (at `36a179b`) | **Active / Modified** | Domain-isolated |
| `Backend/pkg/**` | **Frozen** (at `36a179b`) | **Active / Modified** | Domain-isolated |
| `Backend/go.mod` | **Frozen** (at `36a179b`) | **Active / Modified** | Domain-isolated |
| `docs/rbac/**` | **Synchronized** | **Synchronized** | Shared Documentation |
| `docs/git-branch-strategy.md` | **Synchronized** | **Synchronized** | DevOps Specification |
| `.gitignore` | **Synchronized** | **Synchronized** | Root Configuration |
| `README.md` | **Preserved** | **Preserved** | Root Documentation |

---

## 5. Security & Secret Exclusion Verification

### 5.1 `.gitignore` Rules
The root `.gitignore` enforces strict exclusions for all sensitive files, build artifacts, and system binaries:
```gitignore
# Environment & Secrets
.env
.env.*
!.env.example

# Dependency directories
node_modules/
vendor/

# Build artifacts
dist/
build/
bin/
*.exe
*.dll

# Logs & temp
*.log
tmp/
```

### 5.2 Verification Commands Executed
1. **Rule Active Check:**
   ```powershell
   git check-ignore -v Backend/.env Frontend/.env
   # Output:
   # .gitignore:1:.env    Backend/.env
   # .gitignore:1:.env    Frontend/.env
   ```
2. **Repository History Secret Audit:**
   ```powershell
   git log --all --full-history -- "**/.*env*" "**.env*"
   # Output: Only .env.example files committed across all branches.
   # No live credentials, private keys, or API tokens were ever committed.
   ```

---

## 6. Build & Test Validation Matrix

| Domain | Command | Result | Metrics / Observations |
| :--- | :--- | :--- | :--- |
| **Frontend** | `npm.cmd run build --prefix Frontend` | **PASSED** | 3.06s build time; generated chunk assets cleanly |
| **Frontend** | `npm.cmd run lint --prefix Frontend` | **PASSED** | 0 errors, 0 warnings across 83 files |
| **Frontend UI** | Playwright / Headless Browser Audit | **PASSED** | Login (`/login`) & Register (`/register`) render cleanly |
| **Backend** | `go test -count=1 ./...` | **PASSED** | 11 of 11 packages passed (100% test suite success) |
| **Backend** | Session Concurrency Test | **PASSED** | Multi-goroutine distributed token race resolved atomically |
| **Backend** | RBAC Permission Verification | **PASSED** | Role matrix and policy engine tested against PostgreSQL 18 |

---

## 7. Developer Workflow Guidelines

### 7.1 For Frontend Developers
1. **Work from `frontend` branch:**
   ```powershell
   git checkout frontend
   git pull origin frontend # when remote tracking is enabled by lead
   ```
2. **Create feature branches from `frontend`:**
   ```powershell
   git checkout -b feat/frontend-task-name frontend
   ```
3. **Commit exclusively within `Frontend/` and frontend docs:**
   ```powershell
   git add Frontend/ docs/
   git commit -m "feat(frontend): add task management view"
   ```
4. **Merge back to `frontend`:**
   Submit Pull Requests targeting `frontend`. Do NOT merge directly into `main` or `backend`.

### 7.2 For Backend Developers
1. **Work from `backend` branch:**
   ```powershell
   git checkout backend
   git pull origin backend # when remote tracking is enabled by lead
   ```
2. **Create feature branches from `backend`:**
   ```powershell
   git checkout -b feat/backend-task-name backend
   ```
3. **Commit exclusively within `Backend/` and backend docs:**
   ```powershell
   git add Backend/ docs/
   git commit -m "feat(backend): add task CRUD handlers"
   ```
4. **Merge back to `backend`:**
   Submit Pull Requests targeting `backend`. Do NOT merge directly into `main` or `frontend`.

### 7.3 Integration & Release to `main`
When both branches reach a verified release milestone (e.g., Phase 5 completion):
1. Create a release integration branch from `backend`:
   ```powershell
   git checkout -b release/v1.0 backend
   ```
2. Merge `frontend` into the release branch:
   ```powershell
   git merge frontend --no-ff -m "chore(release): integrate frontend and backend for v1.0"
   ```
   *Because file trees are orthogonal (`Frontend/` modified on `frontend`, `Backend/` modified on `backend`), the merge executes cleanly with zero file collisions.*
3. Perform end-to-end integration tests.
4. Fast-forward or merge `release/v1.0` into `main` and tag the release:
   ```powershell
   git checkout main
   git merge release/v1.0 --no-ff -m "release: Mindora v1.0"
   git tag -a v1.0.0 -m "Mindora Release v1.0.0"
   ```

---

## 8. Summary of Completed Actions

1. Baseline commit `36a179b` audited, tested, and validated as common root.
2. Legacy flat structure preserved at `legacy/backend-flat` (`58bd9c8`).
3. Complete Phase 4 snapshot secured at `backup/phase-4-baseline` (`71a6786`).
4. `frontend` branch created and verified at `82016dc` (`Backend/` identical to `36a179b`).
5. `backend` branch created and verified at `5fdaac7` (`Frontend/` identical to `36a179b`).
6. Full test suites validated on both branches:
   - Frontend: `npm run build` and `npm run lint` pass with 0 errors.
   - Backend: `go test -count=1 ./...` passes 100% across all 11 packages.
7. Secrets, credentials, and build artifacts excluded via verified `.gitignore`.
8. No automatic push or force push executed; all Git operations remain local and safe.
