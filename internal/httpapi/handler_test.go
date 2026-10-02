package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nba-stats-api/internal/httpapi"
)

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		databaseErr error
		wantCode    int
		wantStatus  string
		wantCalls   int
	}{
		{
			name:        "live with database unavailable",
			method:      http.MethodGet,
			path:        "/health/live",
			databaseErr: errors.New("database unavailable"),
			wantCode:    http.StatusOK,
			wantStatus:  "ok",
		},
		{
			name:       "ready with database available",
			method:     http.MethodGet,
			path:       "/health/ready",
			wantCode:   http.StatusOK,
			wantStatus: "ok",
			wantCalls:  1,
		},
		{
			name:        "ready with database unavailable",
			method:      http.MethodGet,
			path:        "/health/ready",
			databaseErr: errors.New("database unavailable"),
			wantCode:    http.StatusServiceUnavailable,
			wantStatus:  "unavailable",
			wantCalls:   1,
		},
		{
			name:     "live rejects POST",
			method:   http.MethodPost,
			path:     "/health/live",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "ready rejects POST",
			method:   http.MethodPost,
			path:     "/health/ready",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "unknown route",
			method:   http.MethodGet,
			path:     "/missing",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0

			handler := httpapi.NewHandler(func(ctx context.Context) error {
				calls++
				return tt.databaseErr
			}, teamReaderStub{})

			request := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != tt.wantCode {
				t.Fatalf("expected status %d, got %d",
					tt.wantCode, response.StatusCode)
			}

			if calls != tt.wantCalls {
				t.Fatalf("expected %d database checks, got %d",
					tt.wantCalls, calls)
			}

			if tt.wantStatus == "" {
				return
			}

			if response.Header.Get("Content-Type") != "application/json" {
				t.Fatal("expected application/json")
			}

			if response.Header.Get("Cache-Control") != "no-store" {
				t.Fatal("expected no-store")
			}

			var body map[string]string
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}

			if len(body) != 1 || body["status"] != tt.wantStatus {
				t.Fatalf("expected status %q, got %v", tt.wantStatus, body)
			}
		})
	}
}

func TestHealthReadyUsesRequestContext(t *testing.T) {
	var databaseErr error
	var hasDeadline bool

	handler := httpapi.NewHandler(func(ctx context.Context) error {
		_, hasDeadline = ctx.Deadline()
		databaseErr = ctx.Err()
		return databaseErr
	}, teamReaderStub{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	request := httptest.NewRequest(
		http.MethodGet,
		"/health/ready",
		nil,
	).WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if !errors.Is(databaseErr, context.Canceled) {
		t.Fatalf("expected canceled context, got %v", databaseErr)
	}

	if !hasDeadline {
		t.Fatal("expected a deadline for the database check")
	}

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
}
