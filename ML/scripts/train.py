"""
Mindora Machine Learning Training Pipeline
Model: Ridge Regression for Student Productivity Estimation
Features: study_hours_per_day, sleep_hours, phone_usage_hours, stress_level
Target: productivity_score (0 - 100)
"""

import os
import json
import csv
import joblib
import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.dummy import DummyRegressor
from sklearn.linear_model import Ridge, LinearRegression
from sklearn.ensemble import RandomForestRegressor
from sklearn.metrics import mean_absolute_error, root_mean_squared_error, r2_score

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DEFAULT_DATASET = os.path.abspath(os.path.join(SCRIPT_DIR, "..", "datasets", "train.csv"))
FALLBACK_DATASET = r"D:\1. PROYEK 3\Dataset ML\student prediction challenge\train.csv"

DATASET_PATH = DEFAULT_DATASET if os.path.exists(DEFAULT_DATASET) else FALLBACK_DATASET
OUTPUT_DIR = os.path.abspath(os.path.join(SCRIPT_DIR, "..", "models"))

MODEL_VERSION = "mindora-productivity-ridge-v1.0.0"

# Strictly validated Mindora candidate features (application-compatible)
FEATURE_COLUMNS = [
    "study_hours_per_day",
    "sleep_hours",
    "phone_usage_hours",
    "stress_level"
]

TARGET_COLUMN = "productivity_score"


def load_dataset(path: str):
    if not os.path.exists(path):
        raise FileNotFoundError(f"Dataset not found at {path}")
    
    with open(path, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        rows = list(reader)
    
    X = np.array([[float(r[col]) for col in FEATURE_COLUMNS] for r in rows], dtype=np.float64)
    y = np.array([float(r[TARGET_COLUMN]) for r in rows], dtype=np.float64)
    return X, y, len(rows)


def run_training():
    print(f"Loading dataset from: {DATASET_PATH}")
    X, y, n_samples = load_dataset(DATASET_PATH)
    print(f"Loaded {n_samples} samples with {len(FEATURE_COLUMNS)} candidate features.")

    # 70% Train, 15% Validation, 15% Test
    X_train_val, X_test, y_train_val, y_test = train_test_split(
        X, y, test_size=0.15, random_state=42
    )
    # Val is 15% of total -> 0.15 / 0.85 = 0.17647
    X_train, X_val, y_train, y_val = train_test_split(
        X_train_val, y_train_val, test_size=0.17647, random_state=42
    )

    print(f"Dataset splits -> Train: {X_train.shape[0]}, Val: {X_val.shape[0]}, Test: {X_test.shape[0]}")

    # Standardize features using ONLY training set
    scaler = StandardScaler()
    X_train_scaled = scaler.fit_transform(X_train)
    X_val_scaled = scaler.transform(X_val)
    X_test_scaled = scaler.transform(X_test)

    # Models to compare
    candidate_models = {
        "baseline_mean": DummyRegressor(strategy="mean"),
        "linear_regression": LinearRegression(),
        "ridge_regression": Ridge(alpha=1.0, random_state=42),
        "random_forest": RandomForestRegressor(n_estimators=100, max_depth=10, random_state=42, n_jobs=-1)
    }

    results = {}
    print("\n--- MODEL COMPARISON & VALIDATION ---")
    for name, model in candidate_models.items():
        model.fit(X_train_scaled, y_train)

        val_preds = model.predict(X_val_scaled)
        val_mae = float(mean_absolute_error(y_val, val_preds))
        val_rmse = float(root_mean_squared_error(y_val, val_preds))
        val_r2 = float(r2_score(y_val, val_preds))

        test_preds = model.predict(X_test_scaled)
        test_mae = float(mean_absolute_error(y_test, test_preds))
        test_rmse = float(root_mean_squared_error(y_test, test_preds))
        test_r2 = float(r2_score(y_test, test_preds))

        results[name] = {
            "val": {"mae": val_mae, "rmse": val_rmse, "r2": val_r2},
            "test": {"mae": test_mae, "rmse": test_rmse, "r2": test_r2}
        }
        print(f"[{name.upper()}] Val: R2={val_r2:.4f}, RMSE={val_rmse:.4f}, MAE={val_mae:.4f} | Test: R2={test_r2:.4f}, RMSE={test_rmse:.4f}, MAE={test_mae:.4f}")

    # Select Ridge Regression for production deployment
    selected_model: Ridge = candidate_models["ridge_regression"]

    # Save joblib model pipeline
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    model_artifact_path = os.path.join(OUTPUT_DIR, "model.joblib")
    joblib.dump({"scaler": scaler, "model": selected_model}, model_artifact_path)
    print(f"\nSerialized joblib model saved to {model_artifact_path}")

    # Export model metadata JSON
    metadata = {
        "model_version": MODEL_VERSION,
        "algorithm": "Ridge Regression (L2-Regularized OLS)",
        "framework": "scikit-learn 1.9.1",
        "trained_date": "2026-11-19",
        "hyperparameters": {
            "alpha": 1.0,
            "fit_intercept": True,
            "random_state": 42
        },
        "target": {
            "name": TARGET_COLUMN,
            "unit": "Productivity Score (0 - 100)",
            "min_observed": float(np.min(y)),
            "max_observed": float(np.max(y)),
            "mean_observed": float(np.mean(y)),
            "std_observed": float(np.std(y))
        },
        "features": [
            {
                "name": "study_hours_per_day",
                "index": 0,
                "unit": "hours",
                "scaler_mean": float(scaler.mean_[0]),
                "scaler_scale": float(scaler.scale_[0]),
                "coefficient": float(selected_model.coef_[0]),
                "imputation_default": float(scaler.mean_[0]),
                "description": "Daily completed study session hours aggregated from study_sessions."
            },
            {
                "name": "sleep_hours",
                "index": 1,
                "unit": "hours",
                "scaler_mean": float(scaler.mean_[1]),
                "scaler_scale": float(scaler.scale_[1]),
                "coefficient": float(selected_model.coef_[1]),
                "imputation_default": float(scaler.mean_[1]),
                "description": "Daily sleep hours recorded in lifestyle_records."
            },
            {
                "name": "phone_usage_hours",
                "index": 2,
                "unit": "hours",
                "scaler_mean": float(scaler.mean_[2]),
                "scaler_scale": float(scaler.scale_[2]),
                "coefficient": float(selected_model.coef_[2]),
                "imputation_default": float(scaler.mean_[2]),
                "description": "Daily smartphone screen time recorded in lifestyle_records."
            },
            {
                "name": "stress_level",
                "index": 3,
                "unit": "scale (1-10)",
                "scaler_mean": float(scaler.mean_[3]),
                "scaler_scale": float(scaler.scale_[3]),
                "coefficient": float(selected_model.coef_[3]),
                "imputation_default": float(scaler.mean_[3]),
                "description": "Self-reported student stress level (1 to 10) recorded in lifestyle_records."
            }
        ],
        "intercept": float(selected_model.intercept_),
        "evaluation_metrics": results,
        "exclusions": [
            {"feature": "id", "reason": "Arbitrary identifier without semantic meaning."},
            {"feature": "student_id", "reason": "Arbitrary identifier without semantic meaning."},
            {"feature": "gender", "reason": "Ethical and fairness restriction; no demographic discrimination."},
            {"feature": "attendance_percentage", "reason": "Mindora application does not collect attendance."},
            {"feature": "final_grade", "reason": "Target leakage prevention; academic grade is not available at daily prediction time."},
            {"feature": "focus_score", "reason": "Synthetic 30-99 score with ungrounded formula; equating focus duration to focus score is invalid."},
            {"feature": "social_media_hours", "reason": "Dataset coefficient is 0.0000; provides no predictive signal."},
            {"feature": "youtube_hours", "reason": "Dataset coefficient is 0.0000; provides no predictive signal."},
            {"feature": "gaming_hours", "reason": "Dataset coefficient is 0.0000; provides no predictive signal."},
            {"feature": "coffee_intake_mg", "reason": "Dataset coefficient is 0.0000; provides no predictive signal."},
            {"feature": "exercise_minutes", "reason": "Dataset coefficient is 0.0000; provides no predictive signal."},
            {"feature": "breaks_per_day", "reason": "Dataset coefficient is 0.0000; provides no predictive signal."},
            {"feature": "assignments_completed", "reason": "Dataset coefficient is 0.0000; provides no predictive signal."}
        ],
        "disclaimer": "Predictions are mathematical estimates based on daily behavioral habits and must NOT be interpreted as actual measured productivity."
    }

    metadata_path = os.path.join(OUTPUT_DIR, "model_metadata.json")
    with open(metadata_path, "w", encoding="utf-8") as f:
        json.dump(metadata, f, indent=2)
    print(f"Model metadata JSON saved to {metadata_path}")


if __name__ == "__main__":
    run_training()
