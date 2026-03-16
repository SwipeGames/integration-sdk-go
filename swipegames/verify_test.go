package swipegames

import (
	"encoding/json"
	"testing"
)

const testIntegrationAPIKey = "test-integration-api-key"

func newTestClient() *Client {
	c, _ := NewClient(ClientConfig{
		CID:               "test-cid",
		ExtCID:            "test-ext-cid",
		APIKey:            "test-api-key",
		IntegrationAPIKey: testIntegrationAPIKey,
		Env:               EnvStaging,
	})
	return c
}

func signBody(body string) string {
	sig, _ := createSignatureFromString(body, testIntegrationAPIKey)
	return sig
}

func signQueryParams(params map[string]string) string {
	sig, _ := createQueryParamsSignature(params, testIntegrationAPIKey)
	return sig
}

func TestVerifyBetRequest(t *testing.T) {
	client := newTestClient()

	t.Run("returns true for valid POST body signature", func(t *testing.T) {
		body := `{"type":"regular","sessionID":"s1","amount":"10.00","txID":"550e8400-e29b-41d4-a716-446655440000","roundID":"660e8400-e29b-41d4-a716-446655440000"}`
		sig := signBody(body)
		if !client.VerifyBetRequest(body, sig) {
			t.Error("expected true")
		}
	})

	t.Run("returns true for valid string body signature", func(t *testing.T) {
		bodyStr := `{"amount":"10.00","roundID":"r1","sessionID":"s1","txID":"tx1","type":"regular"}`
		sig := signBody(bodyStr)
		if !client.VerifyBetRequest(bodyStr, sig) {
			t.Error("expected true")
		}
	})

	t.Run("returns false for missing signature header", func(t *testing.T) {
		if client.VerifyBetRequest(`{"test":true}`, "") {
			t.Error("expected false")
		}
	})

	t.Run("returns false for invalid signature", func(t *testing.T) {
		if client.VerifyBetRequest(`{"test":true}`, "0000000000000000000000000000000000000000000000000000000000000000") {
			t.Error("expected false")
		}
	})

	t.Run("returns false for wrong key", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{"test": true})
		sig, _ := createSignatureFromString(string(body), "wrong-key")
		if client.VerifyBetRequest(string(body), sig) {
			t.Error("expected false")
		}
	})
}

func TestVerifyWinRequest(t *testing.T) {
	client := newTestClient()

	t.Run("returns true for valid signature", func(t *testing.T) {
		body := `{"sessionID":"s1","amount":"50.00","txID":"tx2","roundID":"r1"}`
		sig := signBody(body)
		if !client.VerifyWinRequest(body, sig) {
			t.Error("expected true")
		}
	})
}

func TestVerifyRefundRequest(t *testing.T) {
	client := newTestClient()

	t.Run("returns true for valid signature", func(t *testing.T) {
		body := `{"sessionID":"s1","txID":"tx3","origTxID":"tx1","amount":"10.00"}`
		sig := signBody(body)
		if !client.VerifyRefundRequest(body, sig) {
			t.Error("expected true")
		}
	})
}

func TestVerifyBalanceRequest(t *testing.T) {
	client := newTestClient()

	t.Run("returns true for valid query param signature", func(t *testing.T) {
		params := map[string]string{"sessionID": "session-123"}
		sig := signQueryParams(params)
		if !client.VerifyBalanceRequest(params, sig) {
			t.Error("expected true")
		}
	})

	t.Run("returns false for missing signature", func(t *testing.T) {
		if client.VerifyBalanceRequest(map[string]string{"sessionID": "s1"}, "") {
			t.Error("expected false")
		}
	})

	t.Run("returns false for tampered params", func(t *testing.T) {
		params := map[string]string{"sessionID": "session-123"}
		sig := signQueryParams(params)
		if client.VerifyBalanceRequest(map[string]string{"sessionID": "session-456"}, sig) {
			t.Error("expected false")
		}
	})
}

func TestParseAndVerifyBetRequest(t *testing.T) {
	client := newTestClient()

	t.Run("returns ok with typed body on valid signature", func(t *testing.T) {
		body := map[string]any{
			"type": "regular", "sessionID": "s1", "amount": "10.00",
			"txID": "550e8400-e29b-41d4-a716-446655440000", "roundID": "660e8400-e29b-41d4-a716-446655440000",
		}
		rawBody, _ := json.Marshal(body)
		sig := signBody(string(rawBody))

		result := client.ParseAndVerifyBetRequest(string(rawBody), sig)
		if !result.OK {
			t.Fatalf("expected ok, got error: %v", result.Error)
		}
		if result.Body.Type != BetRequestTypeRegular {
			t.Errorf("type: got %s", result.Body.Type)
		}
		if result.Body.SessionID != "s1" {
			t.Errorf("sessionID: got %s", result.Body.SessionID)
		}
		if result.Body.Amount != "10.00" {
			t.Errorf("amount: got %s", result.Body.Amount)
		}
	})

	t.Run("rejects invalid signature", func(t *testing.T) {
		result := client.ParseAndVerifyBetRequest(`{"sessionID":"s1"}`, "bad-sig")
		if result.OK {
			t.Fatal("expected error")
		}
		if result.Error.Message != "Invalid signature" {
			t.Errorf("message: got %s", result.Error.Message)
		}
	})

	t.Run("rejects missing signature", func(t *testing.T) {
		result := client.ParseAndVerifyBetRequest(`{"sessionID":"s1"}`, "")
		if result.OK {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects invalid JSON body", func(t *testing.T) {
		result := client.ParseAndVerifyBetRequest("not-json", "some-sig")
		if result.OK {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects when body fails validation", func(t *testing.T) {
		invalidBody := `{"type":"invalid_type","sessionID":"s1"}`
		sig := signBody(invalidBody)
		result := client.ParseAndVerifyBetRequest(invalidBody, sig)
		if result.OK {
			t.Fatal("expected error")
		}
		if result.Error.Message != "Invalid request body" {
			t.Errorf("message: got %s", result.Error.Message)
		}
	})

	t.Run("rejects when missing txID and roundID", func(t *testing.T) {
		body := `{"type":"regular","sessionID":"s1","amount":"10.00"}`
		sig := signBody(body)
		result := client.ParseAndVerifyBetRequest(body, sig)
		if result.OK {
			t.Fatal("expected error")
		}
		if result.Error.Message != "Invalid request body" {
			t.Errorf("message: got %s", result.Error.Message)
		}
	})
}

func TestParseAndVerifyWinRequest(t *testing.T) {
	client := newTestClient()

	t.Run("returns ok with typed body", func(t *testing.T) {
		body := map[string]any{
			"type": "regular", "sessionID": "s1", "amount": "50.00",
			"txID": "550e8400-e29b-41d4-a716-446655440002", "roundID": "660e8400-e29b-41d4-a716-446655440000",
		}
		rawBody, _ := json.Marshal(body)
		sig := signBody(string(rawBody))

		result := client.ParseAndVerifyWinRequest(string(rawBody), sig)
		if !result.OK {
			t.Fatalf("expected ok, got error: %v", result.Error)
		}
		if result.Body.Amount != "50.00" {
			t.Errorf("amount: got %s", result.Body.Amount)
		}
	})

	t.Run("rejects when missing required fields", func(t *testing.T) {
		body := `{"type":"regular","sessionID":"s1"}`
		sig := signBody(body)
		result := client.ParseAndVerifyWinRequest(body, sig)
		if result.OK {
			t.Fatal("expected error")
		}
		if result.Error.Message != "Invalid request body" {
			t.Errorf("message: got %s", result.Error.Message)
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		body := `{"type":"invalid_type","sessionID":"s1","amount":"10.00","txID":"550e8400-e29b-41d4-a716-446655440002","roundID":"660e8400-e29b-41d4-a716-446655440000"}`
		sig := signBody(body)
		result := client.ParseAndVerifyWinRequest(body, sig)
		if result.OK {
			t.Fatal("expected error")
		}
		if result.Error.Message != "Invalid request body" {
			t.Errorf("message: got %s", result.Error.Message)
		}
	})

	t.Run("rejects when missing txID and roundID", func(t *testing.T) {
		body := `{"type":"regular","sessionID":"s1","amount":"50.00"}`
		sig := signBody(body)
		result := client.ParseAndVerifyWinRequest(body, sig)
		if result.OK {
			t.Fatal("expected error")
		}
		if result.Error.Message != "Invalid request body" {
			t.Errorf("message: got %s", result.Error.Message)
		}
	})
}

func TestParseAndVerifyRefundRequest(t *testing.T) {
	client := newTestClient()

	t.Run("returns ok with typed refund body", func(t *testing.T) {
		body := map[string]any{
			"sessionID": "s1",
			"txID":      "550e8400-e29b-41d4-a716-446655440001",
			"origTxID":  "550e8400-e29b-41d4-a716-446655440000",
			"amount":    "10.00",
		}
		rawBody, _ := json.Marshal(body)
		sig := signBody(string(rawBody))

		result := client.ParseAndVerifyRefundRequest(string(rawBody), sig)
		if !result.OK {
			t.Fatalf("expected ok, got error: %v", result.Error)
		}
	})

	t.Run("rejects when missing required fields", func(t *testing.T) {
		body := `{"sessionID":"s1","amount":"10.00"}`
		sig := signBody(body)
		result := client.ParseAndVerifyRefundRequest(body, sig)
		if result.OK {
			t.Fatal("expected error")
		}
		if result.Error.Message != "Invalid request body" {
			t.Errorf("message: got %s", result.Error.Message)
		}
	})
}

func TestParseAndVerifyBalanceRequest(t *testing.T) {
	client := newTestClient()

	t.Run("returns ok with typed query on valid signature", func(t *testing.T) {
		params := map[string]string{"sessionID": "session-abc"}
		sig := signQueryParams(params)

		result := client.ParseAndVerifyBalanceRequest(params, sig)
		if !result.OK {
			t.Fatalf("expected ok, got error: %v", result.Error)
		}
		if result.Query.SessionID != "session-abc" {
			t.Errorf("sessionID: got %s", result.Query.SessionID)
		}
	})

	t.Run("rejects invalid signature", func(t *testing.T) {
		result := client.ParseAndVerifyBalanceRequest(map[string]string{"sessionID": "s1"}, "bad-sig")
		if result.OK {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects missing signature", func(t *testing.T) {
		result := client.ParseAndVerifyBalanceRequest(map[string]string{"sessionID": "s1"}, "")
		if result.OK {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects missing sessionID", func(t *testing.T) {
		params := map[string]string{"other": "value"}
		sig := signQueryParams(params)

		result := client.ParseAndVerifyBalanceRequest(params, sig)
		if result.OK {
			t.Fatal("expected error")
		}
		if result.Error.Message != "Missing sessionID" {
			t.Errorf("message: got %s", result.Error.Message)
		}
	})
}
