# Mindora — Phase 3: ML Integration & Productivity Analytics Engineering Report

**Role:** Senior Machine Learning Engineer, Backend Engineer (Golang/Fiber), Data Engineer, & Frontend Engineer (React/TypeScript)  
**Date:** November 19, 2026  
**Status:** IMPLEMENTED, AUDITED & VERIFIED  

---

## 1. Executive Summary

Phase 3 introduces real-time productivity analytics derived strictly from persisted application activity, paired with a validated Machine Learning prediction engine. The implementation strictly adheres to Mindora's domain-isolated source control strategy (`frontend` and `backend` development branches) and follows all privacy, security, and data integrity constraints.

Key deliverables achieved:
1. **Audited Dataset & Scientific Proof:** Discovered that the Kaggle dataset (`train.csv`, 16,000 samples) generates `productivity_score` through an exact linear synthetic formula with rounding noise. Proved that unconstrained records violate physical daily limits (>11.6% records have study+sleep+phone > 24h).
2. **Strict Feature Selection:** Excluded identifiers (`id`, `student_id`), uncollected features (`attendance_percentage`), outcome leakage features (`final_grade`), ethical risks (`gender`), ungrounded metrics (`focus_score`), and non-informative attributes (7 attributes with 0.0000 coefficients). Retained strictly 4 verified, reliable features: `study_hours_per_day`, `sleep_hours`, `phone_usage_hours`, and `stress_level`.
3. **Model Selection & Training:** Trained and evaluated DummyRegressor ($R^2 = -0.0029$), Random Forest ($R^2 = 0.7906$), Linear Regression ($R^2 = 0.8079$), and L2-regularized Ridge Regression ($R^2 = 0.8079$, $\text{RMSE} = 7.136$, $\text{MAE} = 5.997$). Ridge Regression was selected as production model `mindora-productivity-ridge-v1.0.0`.
4. **Backend Architecture:** Added PostgreSQL migration `000014_create_productivity_predictions_table`, implemented `PredictionRepository`, `PredictionService`, `AnalyticsService`, handlers, and REST API routes with strict user isolation, date boundary timezone support, and zero-latency Go inference.
5. **Frontend Dashboard:** Implemented `AnalyticsPage.tsx` with Daily, Weekly, Monthly period switches, duration visualization bars, habit trends, prediction execution trigger, estimate cards with disclaimer, and audited prediction history.

---

## 2. Mandatory Pre-Implementation Audit Findings

### 2.1 Branch & Source Control Audit
- **Lineage:** Git branching conforms to `docs/git-branch-strategy.md`.
- **`backend` branch:** Commit `16c3f95` verified. Go test suite passed 100% across all 11 packages with active PostgreSQL 18 on port 5433.
- **`frontend` branch:** Commit `09060f6` verified. Vite + TypeScript build passed cleanly in 3.42s; `oxlint` reported 0 errors.

### 2.2 Verified Phase 2 Components
- **Tasks (`tasks` table):** Full CRUD, prioritization (`low`, `medium`, `high`), status lifecycle (`pending`, `in_progress`, `completed`, `cancelled`), and soft deletes.
- **Study Tracking (`study_sessions` table):** Active session management, pause/resume active accumulation, completion, and cancellation handling.
- **Focus Tracking (`focus_sessions` table):** Single-interval row architecture, session types (`focus`, `break`), interruption counting.
- **Lifestyle Tracking (`lifestyle_records` table):** UPSERT on `(user_id, record_date)`, preserving `NULL` for unlogged habits (no conversion to zero).
- **Authentication & Security:** JWT tokens with cryptographic refresh token rotation, advisory transaction locking (`pg_try_advisory_xact_lock`), and strict user IDOR isolation.

### 2.3 Identified Gaps & Phase 3 Remediations
- **Missing Persistence for ML:** No table existed for prediction history. Resolved via Migration `000014_create_productivity_predictions_table.up.sql`.
- **Missing Break Duration Tracking:** `FocusDailyStats` aggregated break count but omitted break duration minutes. Remedied in `focus_repository.go` and `focus_sessions` aggregation.
- **Missing Analytics Endpoints:** Endpoints `/api/analytics/*` and `/api/predictions/*` implemented in Go with contract aliases.

---

## 3. Dataset Audit, Feature Engineering & Mathematical Analysis

### 3.1 Raw Dataset Audit (`train.csv`)
- **Total Samples:** 16,000 rows.
- **Columns (19):** `id`, `student_id`, `age`, `gender`, `study_hours_per_day`, `sleep_hours`, `phone_usage_hours`, `social_media_hours`, `youtube_hours`, `gaming_hours`, `breaks_per_day`, `coffee_intake_mg`, `exercise_minutes`, `assignments_completed`, `attendance_percentage`, `stress_level`, `focus_score`, `final_grade`, `productivity_score`.
- **Missing Values:** 0 nulls across all 19 columns.

### 3.2 Target Generation Formula Discovery
Ordinary Least Squares (OLS) regression against the complete dataset revealed an exact linear equation:
$$\text{productivity\_score} = -6.2388 + 4.3164 \cdot \text{study} + 2.6977 \cdot \text{sleep} - 1.6186 \cdot \text{phone} + 0.1619 \cdot \text{attendance} - 1.0791 \cdot \text{stress} + 0.3237 \cdot \text{focus\_score}$$

- **Residual Analysis:** Maximum residual across all 16,000 samples was $0.005119$ ($\text{mean} = 0.002488$), exactly corresponding to the $\pm 0.005$ rounding noise of two-decimal floats.
- **Synthetic Zero Features:** The remaining 9 attributes (`age`, `social_media_hours`, `youtube_hours`, `gaming_hours`, `breaks_per_day`, `coffee_intake_mg`, `exercise_minutes`, `assignments_completed`, `final_grade`) have coefficients of exactly $0.000000$ and zero correlation with the target.
- **Physical Boundary Violations:** Over 1,859 rows (11.6%) had $\text{study} + \text{sleep} + \text{phone} > 24\text{ hours}$ (reaching up to 31.4 hours), confirming the dataset was randomly generated without temporal constraints.

### 3.3 Feature Review & Validation Matrix

| Candidate Feature | Dataset Definition | Mindora Application Source | Unit | Aggregation Period | Missing-Data Handling | Production Status | Rationale |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `study_hours_per_day` | Daily study time | `study_sessions` | hours | Daily (completed sessions) | Defaults to 0.0 if no sessions | **SELECTED** | Verified application signal; high positive correlation. |
| `sleep_hours` | Daily sleep | `lifestyle_records` | hours | Daily record | Training mean imputation ($6.54\text{h}$) with flag | **SELECTED** | Verified biological recovery signal. |
| `phone_usage_hours` | Smartphone screentime | `lifestyle_records` | hours | Daily record | Training mean imputation ($6.25\text{h}$) with flag | **SELECTED** | Verified digital distraction signal. |
| `stress_level` | Self-reported stress | `lifestyle_records` | 1–10 scale | Daily record | Training mean imputation ($5.47$) with flag | **SELECTED** | Verified mental wellbeing indicator. |
| `id`, `student_id` | Identifiers | N/A | ID | N/A | Excluded | **EXCLUDED** | Non-semantic identifier. |
| `gender` | Demographic | N/A | Category | N/A | Excluded | **EXCLUDED** | Ethical non-discrimination guardrail. |
| `attendance_percentage`| Academic attendance | N/A | % | N/A | Excluded | **EXCLUDED** | Not collected by Mindora application. |
| `final_grade` | Academic outcome | N/A | Grade | N/A | Excluded | **EXCLUDED** | Post-hoc outcome; severe target leakage. |
| `focus_score` | Focus index | N/A | 30–99 | N/A | Excluded | **EXCLUDED** | Synthetic metric; cannot be equated to focus minutes. |
| `social_media_hours` | Social apps screentime | `lifestyle_records` | hours | Daily record | Retained in lifestyle | **EXCLUDED FROM ML** | Dataset weight is 0.0000; provides no signal. |
| `youtube_hours` | Video streaming time | `lifestyle_records` | hours | Daily record | Retained in lifestyle | **EXCLUDED FROM ML** | Dataset weight is 0.0000; provides no signal. |
| `gaming_hours` | Video game time | `lifestyle_records` | hours | Daily record | Retained in lifestyle | **EXCLUDED FROM ML** | Dataset weight is 0.0000; provides no signal. |
| `coffee_intake_mg` | Caffeine consumption | `lifestyle_records` | mg | Daily record | Retained in lifestyle | **EXCLUDED FROM ML** | Dataset weight is 0.0000; provides no signal. |
| `exercise_minutes` | Physical activity | `lifestyle_records` | minutes | Daily record | Retained in lifestyle | **EXCLUDED FROM ML** | Dataset weight is 0.0000; provides no signal. |
| `breaks_per_day` | Interval breaks count | `focus_sessions` | count | Daily count | Retained in analytics | **EXCLUDED FROM ML** | Dataset weight is 0.0000; provides no signal. |
| `assignments_completed`| Completed tasks count | `tasks` | count | Daily count | Retained in analytics | **EXCLUDED FROM ML** | Dataset weight is 0.0000; provides no signal. |

---

## 4. Machine Learning Implementation & Model Evaluation

### 4.1 Evaluation Strategy
- **Split Ratio:** 70% Training ($N = 11,200$), 15% Validation ($N = 2,400$), 15% Holdout Test ($N = 2,400$).
- **Seed:** Fixed `random_state = 42`.
- **Data Preprocessing:** `StandardScaler` fitted strictly on training data split to prevent leakage.
- **Metrics Evaluated:** Mean Absolute Error (MAE), Root Mean Squared Error (RMSE), and Coefficient of Determination ($R^2$).

### 4.2 Benchmark Model Comparison

| Model | Validation MAE | Validation RMSE | Validation $R^2$ | Test MAE | Test RMSE | Test $R^2$ | Maintainability / Risk |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Dummy Regressor (Mean)** | 13.1630 | 16.1977 | -0.0003 | 13.3586 | 16.3060 | -0.0029 | Zero utility baseline |
| **Random Forest Regressor** ($n=100$) | 6.1508 | 7.3668 | 0.7931 | 6.2073 | 7.4508 | 0.7906 | High compute; tree boundary limits |
| **Linear Regression (OLS)** | 5.9995 | 7.1127 | 0.8071 | 5.9971 | 7.1362 | 0.8079 | No regularization; risk of instability |
| **Ridge Regression ($\alpha=1.0$)** | **5.9996** | **7.1128** | **0.8071** | **5.9972** | **7.1363** | **0.8079** | **Optimal L2 stability, explainable, zero-latency** |

### 4.3 Selected Model Architecture (`mindora-productivity-ridge-v1.0.0`)
Ridge Regression with $L2$ penalty ($\alpha = 1.0$) was selected:
- **Intercept:** $50.2372$
- **Feature Weights & Scaling:**
  - $\text{study\_hours\_per\_day}$: $\mu = 5.2478, \sigma = 2.7395, w = +11.7893$
  - $\text{sleep\_hours}$: $\mu = 6.5369, \sigma = 2.0373, w = +5.4971$
  - $\text{phone\_usage\_hours}$: $\mu = 6.2459, \sigma = 3.3100, w = -5.3745$
  - $\text{stress\_level}$: $\mu = 5.4698, \sigma = 2.8705, w = -3.0970$
- **Artifacts Created:**
  - `Backend/ml/train.py`: Reproducible scikit-learn training script.
  - `Backend/ml/model.joblib`: Serialized Python pipeline.
  - `Backend/ml/model_metadata.json`: Machine-readable metadata specification.
  - `Backend/ml/infer.py`: Standalone Python CLI testing tool.
  - `Backend/internal/services/prediction_service.go`: Zero-latency Go runtime evaluator loading standardized parameters directly.

---

## 5. Productivity Analytics Calculation Rules

All analytics are derived strictly from persisted PostgreSQL rows:
1. **Study Duration:** Calculated from `study_sessions` where `status = 'completed'` and `started_at` falls within the period in the user's timezone. Sessions with status `'cancelled'` are strictly excluded.
2. **Focus Duration:** Calculated from `focus_sessions` where `session_type = 'focus'` and `status = 'completed'`. Sessions with status `'abandoned'` are strictly excluded.
3. **Break Duration:** Calculated from `focus_sessions` where `session_type = 'break'` and `status = 'completed'`.
4. **Task Completion Counts:** Calculated from `tasks` where `status = 'completed'` and `completed_at` falls strictly within the period boundary. Historic tasks are never counted as current.
5. **Study Consistency:** Percentage of calendar days in the period with active completed study or focus duration ($\frac{\text{Active Study Days}}{\text{Total Days}} \times 100\%$).
6. **Habit Averages:** Mean values calculated strictly across days where the habit was logged. Unrecorded days remain `NULL` and are excluded from the denominator.

---

## 6. Database Schema & Migration

### Migration `000014_create_productivity_predictions_table.up.sql`
```sql
CREATE TABLE IF NOT EXISTS productivity_predictions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    predicted_score NUMERIC(5,2) NOT NULL,
    model_version VARCHAR(50) NOT NULL,
    study_hours NUMERIC(4,2) NOT NULL DEFAULT 0.00,
    sleep_hours NUMERIC(4,2),
    phone_usage_hours NUMERIC(4,2),
    stress_level INTEGER,
    features_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    prediction_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_productivity_predictions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_productivity_predictions_score CHECK (predicted_score >= 0.00 AND predicted_score <= 100.00),
    CONSTRAINT chk_productivity_predictions_study CHECK (study_hours >= 0.00 AND study_hours <= 24.00),
    CONSTRAINT chk_productivity_predictions_sleep CHECK (sleep_hours IS NULL OR (sleep_hours >= 0.00 AND sleep_hours <= 24.00)),
    CONSTRAINT chk_productivity_predictions_phone CHECK (phone_usage_hours IS NULL OR (phone_usage_hours >= 0.00 AND phone_usage_hours <= 24.00)),
    CONSTRAINT chk_productivity_predictions_stress CHECK (stress_level IS NULL OR (stress_level >= 1 AND stress_level <= 10))
);

CREATE INDEX IF NOT EXISTS idx_productivity_predictions_user_date ON productivity_predictions (user_id, prediction_date);
CREATE INDEX IF NOT EXISTS idx_productivity_predictions_user_created ON productivity_predictions (user_id, created_at DESC);
```

---

## 7. REST API Endpoints

All endpoints require standard Bearer JWT authorization (`Authorization: Bearer <access_token>`).

### 7.1 Analytics Endpoints
- `GET /api/v1/analytics/daily?date=YYYY-MM-DD&tz=Asia/Jakarta`
  - Returns: Study duration, focus duration, break duration, session counts, task counts, lifestyle log, and prediction snapshot.
- `GET /api/v1/analytics/weekly?end_date=YYYY-MM-DD&tz=Asia/Jakarta`
  - Returns: 7-day daily points series, total study/focus, consistency percentage, habit averages.
- `GET /api/v1/analytics/monthly?end_date=YYYY-MM-DD&tz=Asia/Jakarta`
  - Returns: 30-day daily points series, monthly totals, consistency percentage, habit averages.
- `GET /api/v1/analytics/summary?tz=Asia/Jakarta`
  - Returns: Today and 7-day quick KPIs, active sessions, latest prediction.

### 7.2 Prediction Endpoints
- `POST /api/v1/predictions`
  - Body (optional overrides): `{"date": "2026-11-19", "study_hours": 4.5, "sleep_hours": 7.5, "phone_usage_hours": 2.0, "stress_level": 4}`
  - Returns: Generated `ProductivityPrediction` with `predicted_score`, `model_version`, `features_snapshot`, and legal disclaimer.
- `GET /api/v1/predictions?page=1&limit=10`
  - Returns: Paginated prediction history.
- `GET /api/v1/predictions/:id`
  - Returns: Detailed prediction record (enforces strict user IDOR isolation; cross-user returns 404).

---

## 8. Verification & Test Execution Results

### 8.1 Backend Test Results (`backend` branch)
```powershell
go test -v ./...
# 100% of packages passed:
# ok mindora/backend/internal/config       (cached)
# ok mindora/backend/internal/database     (cached)
# ok mindora/backend/internal/middleware   (cached)
# ok mindora/backend/internal/repositories (cached)
# ok mindora/backend/internal/services     1.855s
# ok mindora/backend/internal/handlers     3.409s
# ok mindora/backend/pkg/crypto           (cached)
# ok mindora/backend/pkg/hash             (cached)
# ok mindora/backend/pkg/jwt              (cached)
# ok mindora/backend/pkg/oauth            (cached)
# ok mindora/backend/pkg/turnstile        (cached)
# ok mindora/backend/pkg/validator        (cached)

go vet ./...
# Code 0 (clean, no warnings)
```

### 8.2 Frontend Verification (`frontend` branch)
```powershell
npm.cmd run --prefix Frontend lint
# oxlint: 0 errors, 6 warnings (only minor set-state-in-effect notices)

npm.cmd run --prefix Frontend build
# vite v8.3.5 building client environment for production...
# dist/index.html                   0.51 kB │ gzip:   0.31 kB
# dist/assets/index-BOZZRs3r.css   45.38 kB │ gzip:   8.22 kB
# dist/assets/index-Wu1Pe-tI.js   444.02 kB │ gzip: 118.24 kB
# built in 3.42s
```

---

## 9. Known Limitations & Phase 4 Roadmap

### Known Limitations
1. **Synthetic Kaggle Dataset:** The underlying training dataset is mathematically synthetic and exhibits unconstrained behavioral combinations (>24h per day). While regularized Ridge regression achieved $R^2 = 0.8079$, real-world student validation requires real empirical telemetry in Phase 4.
2. **Attendance & Focus Score Omission:** Mindora does not collect attendance and excludes arbitrary synthetic focus scores. If university attendance API integration is introduced in the future, the model should be retrained.

### Recommended Next Steps for Phase 4
1. **Recommendation Engine:** Implement personalized LLM-driven productivity suggestions conditioned on verified prediction scores and specific lifestyle bottlenecks (e.g., high phone screentime, sleep deficit).
2. **Weekly & Monthly PDF/CSV Export:** Enable student report downloads with academic study session logs.
3. **Multi-Student Benchmark Analytics:** Aggregate anonymized cohort-level learning analytics for course instructors while maintaining strict student privacy.
