# Mindora Machine Learning Subsystem

This directory contains the machine learning assets, training pipeline, and inference CLI for the **Mindora** student productivity prediction engine.

## Model Overview

- **Model Version:** `mindora-productivity-ridge-v1.0.0`
- **Algorithm:** Ridge Regression (L2-Regularized Ordinary Least Squares, $\alpha = 1.0$)
- **Target:** `productivity_score` (continuous scale, clamped to $[0.0, 100.0]$)
- **Framework:** `scikit-learn` / Python 3.10+
- **Production Execution:** Natively evaluated in Golang (`Backend/internal/services/prediction_service.go`). Python runtime is **NOT** required at production runtime.

## Feature Schema & Standard Scaling Parameters

Fitted on the 70% training split of `datasets/train.csv`:

| Feature | Unit | Scaler Mean ($\mu$) | Scaler Scale ($\sigma$) | Ridge Coeff ($\beta$) |
|---|---|---|---|---|
| `study_hours_per_day` | hours | 5.247772321428595 | 2.7395364656044996 | +11.789324586911583 |
| `sleep_hours` | hours | 6.536850892857139 | 2.0372977820552127 | +5.497139844559539 |
| `phone_usage_hours` | hours | 6.245928571428571 | 3.3099733887425202 | -5.374465254832190 |
| `stress_level` | scale (1-10) | 5.469821428571429 | 2.8705057985555413 | -3.0970200605794975 |

- **Model Intercept ($\beta_0$):** `50.23716964285723`

## Directory Structure

```text
ML/
├── datasets/           # Reference datasets (train.csv, test.csv, sample_submission.csv)
├── models/             # Serialized artifacts (model.joblib, model_metadata.json)
├── notebooks/          # Exploratory Jupyter notebooks
├── outputs/            # Training evaluation figures and benchmark logs
├── scripts/
│   ├── train.py        # Reproducible training pipeline & metadata exporter
│   └── infer.py        # Standalone CLI inference tool
├── requirements.txt    # Python dependencies (scikit-learn, numpy, joblib)
└── README.md           # Documentation
```

## Quick Start (Offline CLI Inference)

```bash
python scripts/infer.py --study 4.0 --sleep 7.0 --phone 2.5 --stress 4.0
```

## Retraining / Model Reproduction

```bash
pip install -r requirements.txt
python scripts/train.py
```
Outputs `models/model.joblib` and `models/model_metadata.json`.
