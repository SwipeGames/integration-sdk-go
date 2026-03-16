package swipegames

import "testing"

func TestCreateSignature(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		apiKey   string
		expected string
	}{
		{
			name:     "signs JSON object with JCS canonicalization",
			data:     `{"user_id": 123, "amount": 100.50}`,
			apiKey:   "secret-key",
			expected: "9876ed3affd6596f3ddb9102a396718452cf83069904f3d001a2e91e164adc01",
		},
		{
			name:     "produces same signature regardless of key order",
			data:     `{"amount":100.5,"user_id":123}`,
			apiKey:   "secret-key",
			expected: "9876ed3affd6596f3ddb9102a396718452cf83069904f3d001a2e91e164adc01",
		},
		{
			name:     "signs empty object",
			data:     "{}",
			apiKey:   "secret-key",
			expected: "99922a0dbb1fe95624c93c7204445c2eff8a014b0c9b585ddf2da0c21083a34e",
		},
		{
			name:     "handles whitespace in JSON input",
			data:     `{  "user_id"  :  123  ,  "amount"  :  100.50  }`,
			apiKey:   "secret-key",
			expected: "9876ed3affd6596f3ddb9102a396718452cf83069904f3d001a2e91e164adc01",
		},
		{
			name:     "different key produces different signature",
			data:     `{"user_id": 123, "amount": 100.50}`,
			apiKey:   "different-secret-key",
			expected: "d86208a306f6562c80c0a8894a1294a63e5e3bb4e2fd2b9b031b3c3c65cb1847",
		},
		{
			name:     "works with empty key",
			data:     `{"test": "value"}`,
			apiKey:   "",
			expected: "6c0e6084444acce7905532fd7c3871c33cfbc5f52a36d27704ffa02b1bb4df78",
		},
		{
			name:     "accepts map input (not just string)",
			data:     map[string]any{"user_id": float64(123), "amount": 100.5},
			apiKey:   "secret-key",
			expected: "9876ed3affd6596f3ddb9102a396718452cf83069904f3d001a2e91e164adc01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := createSignature(tt.data, tt.apiKey)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sig != tt.expected {
				t.Errorf("got %s, want %s", sig, tt.expected)
			}
		})
	}
}

func TestCreateQueryParamsSignature(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]string
		apiKey   string
		expected string
	}{
		{
			name: "signs single query param",
			params: map[string]string{
				"sessionID": "7eaac66f751bcdb758877004b0a1c0063bdfb615ee0c20a464ae76edc67db324113f1ca8bd62b13dd1c7a43f85a20ea3",
			},
			apiKey:   "secret-key",
			expected: "23b02858e21abd151a4e48ed33e451cae4ad1b7cb267ef75d01c694ea2960e6d",
		},
		{
			name:     "signs empty query params",
			params:   map[string]string{},
			apiKey:   "secret-key",
			expected: "99922a0dbb1fe95624c93c7204445c2eff8a014b0c9b585ddf2da0c21083a34e",
		},
		{
			name:     "signs multiple query params with special characters",
			params:   map[string]string{"message": "hello world!", "data": "test@example.com"},
			apiKey:   "secret-key",
			expected: "0825b42e92c46887f194252fda8b871c3c42aafa3833783d63b2005407000c02",
		},
		{
			name:     "handles empty param value",
			params:   map[string]string{"empty": "", "data": "value"},
			apiKey:   "secret-key",
			expected: "8cf8644bfb7004cd21ad8512923169bb652d836183c07497797ef1ca313d88cc",
		},
		{
			name:     "works with empty key",
			params:   map[string]string{"test": "value"},
			apiKey:   "",
			expected: "6c0e6084444acce7905532fd7c3871c33cfbc5f52a36d27704ffa02b1bb4df78",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := createQueryParamsSignature(tt.params, tt.apiKey)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sig != tt.expected {
				t.Errorf("got %s, want %s", sig, tt.expected)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	t.Run("returns true for valid signature", func(t *testing.T) {
		ok, err := verifySignature(
			`{"user_id": 123, "amount": 100.50}`,
			"9876ed3affd6596f3ddb9102a396718452cf83069904f3d001a2e91e164adc01",
			"secret-key",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for invalid signature", func(t *testing.T) {
		ok, err := verifySignature(
			`{"user_id": 123, "amount": 100.50}`,
			"0000000000000000000000000000000000000000000000000000000000000000",
			"secret-key",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected false, got true")
		}
	})

	t.Run("returns false for wrong key", func(t *testing.T) {
		ok, err := verifySignature(
			`{"user_id": 123, "amount": 100.50}`,
			"9876ed3affd6596f3ddb9102a396718452cf83069904f3d001a2e91e164adc01",
			"wrong-key",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected false, got true")
		}
	})
}

func TestCanonicalizeNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{name: "zero", input: 0, expected: "0"},
		{name: "positive integer", input: 42, expected: "42"},
		{name: "negative integer", input: -42, expected: "-42"},
		{name: "decimal", input: 1.5, expected: "1.5"},
		{name: "very small number uses negative exponent", input: 1.5e-7, expected: "1.5e-7"},
		{name: "very large number uses positive exponent", input: 1.5e21, expected: "1.5e+21"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canonicalizeNumber(tt.input)
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestVerifyQueryParamsSignature(t *testing.T) {
	t.Run("returns true for valid query param signature", func(t *testing.T) {
		ok, err := verifyQueryParamsSignature(
			map[string]string{
				"sessionID": "7eaac66f751bcdb758877004b0a1c0063bdfb615ee0c20a464ae76edc67db324113f1ca8bd62b13dd1c7a43f85a20ea3",
			},
			"23b02858e21abd151a4e48ed33e451cae4ad1b7cb267ef75d01c694ea2960e6d",
			"secret-key",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false for invalid signature", func(t *testing.T) {
		ok, err := verifyQueryParamsSignature(
			map[string]string{"sessionID": "test"},
			"0000000000000000000000000000000000000000000000000000000000000000",
			"secret-key",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected false, got true")
		}
	})
}
