package swipegames

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

var testConfig = ClientConfig{
	CID:               "550e8400-e29b-41d4-a716-446655440000",
	ExtCID:            "test_ext_cid",
	APIKey:            "test-api-key",
	IntegrationAPIKey: "test-integration-api-key",
	Env:               EnvStaging,
}

func TestNewClient(t *testing.T) {
	t.Run("uses staging URL by default", func(t *testing.T) {
		cfg := testConfig
		cfg.Env = ""
		c, err := NewClient(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(c.baseURL, "staging.platform.0.swipegames.io") {
			t.Errorf("expected staging URL, got %s", c.baseURL)
		}
	})

	t.Run("uses production URL when env is production", func(t *testing.T) {
		cfg := testConfig
		cfg.Env = EnvProduction
		c, err := NewClient(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(c.baseURL, "prod.platform.1.swipegames.io") {
			t.Errorf("expected production URL, got %s", c.baseURL)
		}
	})

	t.Run("uses custom baseURL when provided", func(t *testing.T) {
		cfg := testConfig
		cfg.BaseURL = "https://custom.api/v1"
		c, err := NewClient(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if c.baseURL != "https://custom.api/v1" {
			t.Errorf("expected custom URL, got %s", c.baseURL)
		}
	})

	t.Run("returns error on unknown env", func(t *testing.T) {
		cfg := testConfig
		cfg.Env = "unknown"
		_, err := NewClient(cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "unknown env: unknown") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("returns ValidationError when CID is empty", func(t *testing.T) {
		cfg := testConfig
		cfg.CID = ""
		_, err := NewClient(cfg)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected ValidationError, got %T: %v", err, err)
		}
		if valErr.Field != "CID" {
			t.Errorf("field: got %s", valErr.Field)
		}
	})

	t.Run("returns ValidationError when ExtCID is empty", func(t *testing.T) {
		cfg := testConfig
		cfg.ExtCID = ""
		_, err := NewClient(cfg)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected ValidationError, got %T: %v", err, err)
		}
		if valErr.Field != "ExtCID" {
			t.Errorf("field: got %s", valErr.Field)
		}
	})

	t.Run("returns ValidationError when APIKey is empty", func(t *testing.T) {
		cfg := testConfig
		cfg.APIKey = ""
		_, err := NewClient(cfg)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected ValidationError, got %T: %v", err, err)
		}
		if valErr.Field != "APIKey" {
			t.Errorf("field: got %s", valErr.Field)
		}
	})

	t.Run("returns ValidationError when IntegrationAPIKey is empty", func(t *testing.T) {
		cfg := testConfig
		cfg.IntegrationAPIKey = ""
		_, err := NewClient(cfg)
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected ValidationError, got %T: %v", err, err)
		}
		if valErr.Field != "IntegrationAPIKey" {
			t.Errorf("field: got %s", valErr.Field)
		}
	})
}

func TestCreateNewGame(t *testing.T) {
	t.Run("sends signed POST request with correct body", func(t *testing.T) {
		gsID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440099")
		mockResponse := CreateNewGameResponse{GameURL: "https://game.url", GsID: gsID}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if !strings.HasSuffix(r.URL.Path, "/create-new-game") {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected application/json content type")
			}
			if r.Header.Get("X-REQUEST-SIGN") == "" {
				t.Errorf("missing X-REQUEST-SIGN header")
			}

			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["cID"] != testConfig.CID {
				t.Errorf("cID: got %v, want %s", body["cID"], testConfig.CID)
			}
			if body["extCID"] != testConfig.ExtCID {
				t.Errorf("extCID: got %v, want %s", body["extCID"], testConfig.ExtCID)
			}
			if body["gameID"] != "sg_catch_97" {
				t.Errorf("gameID: got %v", body["gameID"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		cfg := testConfig
		cfg.BaseURL = server.URL
		client, _ := NewClient(cfg)

		result, err := client.CreateNewGame(context.Background(), CreateNewGameParams{
			GameID:    "sg_catch_97",
			Demo:      false,
			Platform:  PlatformDesktop,
			Currency:  "USD",
			Locale:    "en_us",
			SessionID: "session-123",
			User:      &User{Id: "user-1"},
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.GameURL != "https://game.url" {
			t.Errorf("gameURL: got %s", result.GameURL)
		}
		if result.GsID != gsID {
			t.Errorf("gsID: got %s", result.GsID)
		}
	})

	t.Run("includes optional fields in body when provided", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["returnURL"] != "https://return.url" {
				t.Errorf("returnURL: got %v", body["returnURL"])
			}
			if body["depositURL"] != "https://deposit.url" {
				t.Errorf("depositURL: got %v", body["depositURL"])
			}
			if body["initDemoBalance"] != "1000" {
				t.Errorf("initDemoBalance: got %v", body["initDemoBalance"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CreateNewGameResponse{GameURL: "https://game.url"})
		}))
		defer server.Close()

		cfg := testConfig
		cfg.BaseURL = server.URL
		client, _ := NewClient(cfg)

		_, err := client.CreateNewGame(context.Background(), CreateNewGameParams{
			GameID:          "sg_catch_97",
			Demo:            true,
			Platform:        PlatformDesktop,
			Currency:        "USD",
			Locale:          "en_us",
			ReturnURL:       "https://return.url",
			DepositURL:      "https://deposit.url",
			InitDemoBalance: "1000",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns APIError on non-200 with status, code, and details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid signature",
				"details": "Token expired at 12:00",
			})
		}))
		defer server.Close()

		cfg := testConfig
		cfg.BaseURL = server.URL
		client, _ := NewClient(cfg)

		_, err := client.CreateNewGame(context.Background(), CreateNewGameParams{
			GameID:   "sg_catch_97",
			Demo:     false,
			Platform: PlatformDesktop,
			Currency: "USD",
			Locale:   "en_us",
		})

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected APIError, got %T: %v", err, err)
		}
		if apiErr.StatusCode != 401 {
			t.Errorf("status: got %d", apiErr.StatusCode)
		}
		if apiErr.Message != "Invalid signature" {
			t.Errorf("message: got %s", apiErr.Message)
		}
		if apiErr.Details != "Token expired at 12:00" {
			t.Errorf("details: got %s", apiErr.Details)
		}
	})

	t.Run("returns APIError with code when error response includes code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Game not found",
				"code":    "game_not_found",
			})
		}))
		defer server.Close()

		cfg := testConfig
		cfg.BaseURL = server.URL
		client, _ := NewClient(cfg)

		_, err := client.CreateNewGame(context.Background(), CreateNewGameParams{
			GameID:   "nonexistent",
			Demo:     false,
			Platform: PlatformDesktop,
			Currency: "USD",
			Locale:   "en_us",
		})

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected APIError, got %T: %v", err, err)
		}
		if apiErr.StatusCode != 400 {
			t.Errorf("status: got %d", apiErr.StatusCode)
		}
		if apiErr.Code != "game_not_found" {
			t.Errorf("code: got %s", apiErr.Code)
		}
		if apiErr.Message != "Game not found" {
			t.Errorf("message: got %s", apiErr.Message)
		}
	})

	t.Run("uses status text when error response is not valid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(502)
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		cfg := testConfig
		cfg.BaseURL = server.URL
		client, _ := NewClient(cfg)

		_, err := client.CreateNewGame(context.Background(), CreateNewGameParams{
			GameID:   "sg_catch_97",
			Demo:     false,
			Platform: PlatformDesktop,
			Currency: "USD",
			Locale:   "en_us",
		})

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected APIError, got %T: %v", err, err)
		}
		if apiErr.StatusCode != 502 {
			t.Errorf("status: got %d", apiErr.StatusCode)
		}
	})
}

func TestGetGames(t *testing.T) {
	t.Run("sends signed GET request with query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if !strings.Contains(r.URL.Path, "/games") {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("cID") != testConfig.CID {
				t.Errorf("missing cID query param")
			}
			if r.URL.Query().Get("extCID") != testConfig.ExtCID {
				t.Errorf("missing extCID query param")
			}
			if r.Header.Get("X-REQUEST-SIGN") == "" {
				t.Errorf("missing X-REQUEST-SIGN header")
			}

			// Verify signature
			expectedSig, _ := createQueryParamsSignature(
				map[string]string{"cID": testConfig.CID, "extCID": testConfig.ExtCID},
				testConfig.APIKey,
			)
			if r.Header.Get("X-REQUEST-SIGN") != expectedSig {
				t.Errorf("signature mismatch: got %s, want %s", r.Header.Get("X-REQUEST-SIGN"), expectedSig)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]string{{"id": "sg_catch_97", "title": "Catch 97"}})
		}))
		defer server.Close()

		cfg := testConfig
		cfg.BaseURL = server.URL
		client, _ := NewClient(cfg)

		result, err := client.GetGames(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 game, got %d", len(result))
		}
		if result[0].Id != "sg_catch_97" {
			t.Errorf("id: got %s", result[0].Id)
		}
	})
}

func TestCreateFreeRounds(t *testing.T) {
	t.Run("sends signed POST to /free-rounds", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}

			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["cID"] != testConfig.CID {
				t.Errorf("cID: got %v", body["cID"])
			}
			if body["extID"] != "ext-fr-1" {
				t.Errorf("extID: got %v", body["extID"])
			}
			if body["quantity"] != float64(10) {
				t.Errorf("quantity: got %v", body["quantity"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"id":    "550e8400-e29b-41d4-a716-446655440077",
				"extID": "ext-fr-1",
			})
		}))
		defer server.Close()

		cfg := testConfig
		cfg.BaseURL = server.URL
		client, _ := NewClient(cfg)

		result, err := client.CreateFreeRounds(context.Background(), CreateFreeRoundsParams{
			ExtID:     "ext-fr-1",
			Currency:  "USD",
			Quantity:  10,
			BetLine:   5,
			ValidFrom: "2026-01-01T00:00:00.000Z",
			GameIDs:   []string{"sg_catch_97"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440077")
		if result.Id != expectedID {
			t.Errorf("id: got %s", result.Id)
		}
		if result.ExtID != "ext-fr-1" {
			t.Errorf("extID: got %s", result.ExtID)
		}
	})
}

func TestCancelFreeRounds(t *testing.T) {
	t.Run("sends signed DELETE to /free-rounds", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}

			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["cID"] != testConfig.CID {
				t.Errorf("cID: got %v", body["cID"])
			}
			if body["extID"] != "ext-fr-1" {
				t.Errorf("extID: got %v", body["extID"])
			}

			w.WriteHeader(200)
		}))
		defer server.Close()

		cfg := testConfig
		cfg.BaseURL = server.URL
		client, _ := NewClient(cfg)

		err := client.CancelFreeRounds(context.Background(), CancelFreeRoundsParams{ExtID: "ext-fr-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns ValidationError when neither id nor extID provided", func(t *testing.T) {
		cfg := testConfig
		cfg.BaseURL = "http://localhost"
		client, _ := NewClient(cfg)

		err := client.CancelFreeRounds(context.Background(), CancelFreeRoundsParams{})
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected ValidationError, got %T: %v", err, err)
		}
	})
}
