package swipegames

import (
	"encoding/json"

	"github.com/google/uuid"
)

// VerifyBetRequest verifies the signature of an incoming /bet request.
func (c *Client) VerifyBetRequest(body string, signatureHeader string) bool {
	return c.verifyInboundSignature(body, signatureHeader)
}

// VerifyWinRequest verifies the signature of an incoming /win request.
func (c *Client) VerifyWinRequest(body string, signatureHeader string) bool {
	return c.verifyInboundSignature(body, signatureHeader)
}

// VerifyRefundRequest verifies the signature of an incoming /refund request.
func (c *Client) VerifyRefundRequest(body string, signatureHeader string) bool {
	return c.verifyInboundSignature(body, signatureHeader)
}

// VerifyBalanceRequest verifies the signature of an incoming GET /balance request.
func (c *Client) VerifyBalanceRequest(queryParams map[string]string, signatureHeader string) bool {
	if signatureHeader == "" {
		return false
	}
	ok, err := verifyQueryParamsSignature(queryParams, signatureHeader, c.integrationAPIKey)
	if err != nil {
		return false
	}
	return ok
}

// ParseAndVerifyBetRequest parses, verifies, and validates an incoming /bet request.
func (c *Client) ParseAndVerifyBetRequest(rawBody string, signatureHeader string) (BetRequest, *VerifyError) {
	var body BetRequest
	if err := c.parseAndVerifyInboundRequest(rawBody, signatureHeader, &body); err != nil {
		return BetRequest{}, err
	}

	// validate required fields
	if body.Type == "" || body.SessionID == "" || body.Amount == "" || body.TxID == uuid.Nil || body.RoundID == uuid.Nil {
		return BetRequest{}, newVerifyError("Invalid request body")
	}
	if !body.Type.Valid() {
		return BetRequest{}, newVerifyError("Invalid request body")
	}

	return body, nil
}

// ParseAndVerifyWinRequest parses, verifies, and validates an incoming /win request.
func (c *Client) ParseAndVerifyWinRequest(rawBody string, signatureHeader string) (WinRequest, *VerifyError) {
	var body WinRequest
	if err := c.parseAndVerifyInboundRequest(rawBody, signatureHeader, &body); err != nil {
		return WinRequest{}, err
	}

	// validate required fields
	if body.Type == "" || body.SessionID == "" || body.Amount == "" || body.TxID == uuid.Nil || body.RoundID == uuid.Nil {
		return WinRequest{}, newVerifyError("Invalid request body")
	}
	if !body.Type.Valid() {
		return WinRequest{}, newVerifyError("Invalid request body")
	}

	return body, nil
}

// ParseAndVerifyRefundRequest parses, verifies, and validates an incoming /refund request.
func (c *Client) ParseAndVerifyRefundRequest(rawBody string, signatureHeader string) (RefundRequest, *VerifyError) {
	var body RefundRequest
	if err := c.parseAndVerifyInboundRequest(rawBody, signatureHeader, &body); err != nil {
		return RefundRequest{}, err
	}

	// validate required fields
	if body.SessionID == "" || body.Amount == "" || body.TxID == uuid.Nil || body.OrigTxID == uuid.Nil {
		return RefundRequest{}, newVerifyError("Invalid request body")
	}

	return body, nil
}

// ParseAndVerifyBalanceRequest parses, verifies, and validates an incoming GET /balance request.
func (c *Client) ParseAndVerifyBalanceRequest(queryParams map[string]string, signatureHeader string) (*GetBalanceQuery, *VerifyError) {
	if !c.VerifyBalanceRequest(queryParams, signatureHeader) {
		return nil, newVerifyError("Invalid signature")
	}

	sessionID, ok := queryParams["sessionID"]
	if !ok || sessionID == "" {
		return nil, newVerifyError("Missing sessionID")
	}

	return &GetBalanceQuery{SessionID: sessionID}, nil
}

func (c *Client) verifyInboundSignature(body string, signatureHeader string) bool {
	if signatureHeader == "" {
		return false
	}
	ok, err := verifySignatureFromString(body, signatureHeader, c.integrationAPIKey)
	if err != nil {
		return false
	}
	return ok
}

func (c *Client) parseAndVerifyInboundRequest(rawBody string, signatureHeader string, target interface{}) *VerifyError {
	if !c.verifyInboundSignature(rawBody, signatureHeader) {
		return newVerifyError("Invalid signature")
	}

	if err := json.Unmarshal([]byte(rawBody), target); err != nil {
		return newVerifyError("Invalid request body")
	}

	return nil
}
