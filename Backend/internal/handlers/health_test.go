package handlers_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/routes"
	"mindora/backend/pkg/response"
)

func TestHealthEndpoints(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database for health test: %v", err)
	}
	defer db.Close()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	// Test 1: GET /api/health (Standard system health check)
	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, err := app.Test(req, 2000)
	if err != nil {
		t.Fatalf("failed to execute request /api/health: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	var healthResp response.Response
	if err := json.Unmarshal(bodyBytes, &healthResp); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
	if !healthResp.Success || healthResp.Message != "Mindora API is running" {
		t.Errorf("unexpected response content: %+v", healthResp)
	}

	// Test 2: GET /api/health/db (Database connectivity check)
	reqDB := httptest.NewRequest("GET", "/api/health/db", nil)
	respDB, err := app.Test(reqDB, 2000)
	if err != nil {
		t.Fatalf("failed to execute request /api/health/db: %v", err)
	}
	if respDB.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", respDB.StatusCode)
	}

	bodyDBBytes, _ := io.ReadAll(respDB.Body)
	var dbHealthResp response.Response
	if err := json.Unmarshal(bodyDBBytes, &dbHealthResp); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
	if !dbHealthResp.Success || dbHealthResp.Message != "Database is connected and healthy" {
		t.Errorf("unexpected DB health response content: %+v", dbHealthResp)
	}
}
