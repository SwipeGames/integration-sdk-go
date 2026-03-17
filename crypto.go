package swipegames

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// canonicalizeJSON implements RFC 8785 JSON Canonicalization Scheme.
// It produces a deterministic JSON string with sorted keys and no extra whitespace.
func canonicalizeJSON(data interface{}) (string, error) {
	var obj interface{}

	switch v := data.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &obj); err != nil {
			return "", fmt.Errorf("failed to parse JSON string: %w", err)
		}
	case []byte:
		if err := json.Unmarshal(v, &obj); err != nil {
			return "", fmt.Errorf("failed to parse JSON bytes: %w", err)
		}
	default:
		// marshal and unmarshal to get a generic interface{} representation
		b, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal data: %w", err)
		}
		if err := json.Unmarshal(b, &obj); err != nil {
			return "", fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return serializeCanonical(obj)
}

func serializeCanonical(v interface{}) (string, error) {
	switch val := v.(type) {
	case nil:
		return "null", nil
	case bool:
		if val {
			return "true", nil
		}
		return "false", nil
	case float64:
		return canonicalizeNumber(val), nil
	case string:
		b, err := json.Marshal(val)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case []interface{}:
		var parts []string
		for _, item := range val {
			s, err := serializeCanonical(item)
			if err != nil {
				return "", err
			}
			parts = append(parts, s)
		}
		return fmt.Sprintf("[%s]", strings.Join(parts, ",")), nil
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var parts []string
		for _, k := range keys {
			keyStr, err := json.Marshal(k)
			if err != nil {
				return "", err
			}
			valStr, err := serializeCanonical(val[k])
			if err != nil {
				return "", err
			}
			parts = append(parts, fmt.Sprintf("%s:%s", string(keyStr), valStr))
		}
		return fmt.Sprintf("{%s}", strings.Join(parts, ",")), nil
	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}

// canonicalizeNumber formats a float64 according to RFC 8785 / ES2024 Number.toString().
func canonicalizeNumber(f float64) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "null"
	}
	if f == 0 {
		return "0"
	}
	// use strconv.FormatFloat with -1 precision to get the shortest representation
	s := strconv.FormatFloat(f, 'f', -1, 64)
	// check if scientific notation is shorter / required (matching ES2024 behavior)
	eStr := strconv.FormatFloat(f, 'e', -1, 64)
	// es2024: use exponential for very large or very small numbers
	if math.Abs(f) >= 1e21 || (math.Abs(f) < 1e-6 && f != 0) {
		// format exponent like es2024: e+XX
		return formatES2024Exponential(eStr)
	}
	return s
}

func formatES2024Exponential(s string) string {
	parts := strings.Split(s, "e")
	if len(parts) != 2 {
		return s
	}
	mantissa := parts[0]
	exponent := parts[1]

	// remove trailing zeros from mantissa
	if strings.Contains(mantissa, ".") {
		mantissa = strings.TrimRight(mantissa, "0")
		mantissa = strings.TrimRight(mantissa, ".")
	}

	// format exponent: remove leading zeros, keep sign
	exp, _ := strconv.Atoi(exponent)
	if exp >= 0 {
		return fmt.Sprintf("%se+%d", mantissa, exp)
	}
	return fmt.Sprintf("%se-%d", mantissa, -exp)
}

// createSignatureFromCanonical creates an HMAC-SHA256 signature from an already-canonicalized JSON string.
func createSignatureFromCanonical(canonical string, apiKey string) (string, error) {
	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write([]byte(canonical))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// createSignature creates an HMAC-SHA256 signature using JCS canonicalization.
func createSignature(data interface{}, apiKey string) (string, error) {
	canonical, err := canonicalizeJSON(data)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalize: %w", err)
	}
	return createSignatureFromCanonical(canonical, apiKey)
}

// createSignatureFromString creates an HMAC-SHA256 signature from a JSON string.
func createSignatureFromString(jsonStr string, apiKey string) (string, error) {
	return createSignature(jsonStr, apiKey)
}

// createQueryParamsSignature creates an HMAC-SHA256 signature from query parameters.
func createQueryParamsSignature(params map[string]string, apiKey string) (string, error) {
	// convert to map[string]interface{} for canonicalization
	obj := make(map[string]interface{}, len(params))
	for k, v := range params {
		obj[k] = v
	}
	return createSignature(obj, apiKey)
}

// verifySignature verifies an HMAC-SHA256 signature using timing-safe comparison.
func verifySignature(data interface{}, signature, apiKey string) (bool, error) {
	expected, err := createSignature(data, apiKey)
	if err != nil {
		return false, err
	}
	return safeCompare(expected, signature), nil
}

// verifySignatureFromString verifies an HMAC-SHA256 signature from a JSON string.
func verifySignatureFromString(jsonStr, signature, apiKey string) (bool, error) {
	return verifySignature(jsonStr, signature, apiKey)
}

// verifyQueryParamsSignature verifies a query params signature using timing-safe comparison.
func verifyQueryParamsSignature(params map[string]string, signature, apiKey string) (bool, error) {
	expected, err := createQueryParamsSignature(params, apiKey)
	if err != nil {
		return false, err
	}
	return safeCompare(expected, signature), nil
}

func safeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
