"""
Mindora Standalone Inference CLI
Usage:
    python infer.py --study 4.5 --sleep 7.5 --phone 2.0 --stress 5
"""

import os
import sys
import json
import argparse
import joblib
import numpy as np

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DEFAULT_META = os.path.abspath(os.path.join(SCRIPT_DIR, "..", "models", "model_metadata.json"))
DEFAULT_MODEL = os.path.abspath(os.path.join(SCRIPT_DIR, "..", "models", "model.joblib"))

METADATA_PATH = DEFAULT_META if os.path.exists(DEFAULT_META) else os.path.join(SCRIPT_DIR, "model_metadata.json")
MODEL_PATH = DEFAULT_MODEL if os.path.exists(DEFAULT_MODEL) else os.path.join(SCRIPT_DIR, "model.joblib")


def predict(study_hours: float, sleep_hours: float, phone_hours: float, stress_level: float):
    if not os.path.exists(METADATA_PATH):
        raise FileNotFoundError(f"Metadata file not found: {METADATA_PATH}")

    with open(METADATA_PATH, "r", encoding="utf-8") as f:
        meta = json.load(f)

    # Calculate using metadata weights directly
    feats = meta["features"]
    scaled_feats = []
    inputs = [study_hours, sleep_hours, phone_hours, stress_level]

    dot_product = 0.0
    for i, val in enumerate(inputs):
        f_meta = feats[i]
        # Impute if None
        val_used = f_meta["imputation_default"] if val is None else float(val)
        scaled_val = (val_used - f_meta["scaler_mean"]) / f_meta["scaler_scale"]
        dot_product += scaled_val * f_meta["coefficient"]

    raw_score = meta["intercept"] + dot_product
    clamped_score = max(0.0, min(100.0, raw_score))

    return {
        "model_version": meta["model_version"],
        "algorithm": meta["algorithm"],
        "predicted_score": round(clamped_score, 2),
        "raw_score": round(raw_score, 4),
        "inputs": {
            "study_hours_per_day": study_hours,
            "sleep_hours": sleep_hours,
            "phone_usage_hours": phone_hours,
            "stress_level": stress_level
        },
        "disclaimer": meta["disclaimer"]
    }


def main():
    parser = argparse.ArgumentParser(description="Mindora Productivity Predictor")
    parser.add_argument("--study", type=float, default=4.0, help="Study hours per day")
    parser.add_argument("--sleep", type=float, default=7.0, help="Sleep hours")
    parser.add_argument("--phone", type=float, default=2.5, help="Phone usage hours")
    parser.add_argument("--stress", type=float, default=4.0, help="Stress level (1-10)")
    args = parser.parse_args()

    res = predict(args.study, args.sleep, args.phone, args.stress)
    print(json.dumps(res, indent=2))


if __name__ == "__main__":
    main()
