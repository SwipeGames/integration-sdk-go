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
func (c *Client) ParseAndVerifyBetRequest(rawBody string, signatureHeader string) ParsedRequestResult[BetRequest] {
	var body BetRequest
	result := c.parseAndVerifyInboundRequest(rawBody, signatureHeader, &body)
	if !result.OK {
		return ParsedRequestResult[BetRequest]{OK: false, Error: result.Error}
	}

	// validate required fields
	if body.Type == "" || body.SessionID == "" || body.Amount == "" || body.TxID == uuid.Nil || body.RoundID == uuid.Nil {
		return ParsedRequestResult[BetRequest]{
			OK:    false,
			Error: &ErrorResponseWithCodeAndAction{Message: "Invalid request body"},
		}
	}
	if !body.Type.Valid() {
		return ParsedRequestResult[BetRequest]{
			OK:    false,
			Error: &ErrorResponseWithCodeAndAction{Message: "Invalid request body"},
		}
	}

	return ParsedRequestResult[BetRequest]{OK: true, Body: body}
}

// ParseAndVerifyWinRequest parses, verifies, and validates an incoming /win request.
func (c *Client) ParseAndVerifyWinRequest(rawBody string, signatureHeader string) ParsedRequestResult[WinRequest] {
	var body WinRequest
	result := c.parseAndVerifyInboundRequest(rawBody, signatureHeader, &body)
	if !result.OK {
		return ParsedRequestResult[WinRequest]{OK: false, Error: result.Error}
	}

	// validate required fields
	if body.Type == "" || body.SessionID == "" || body.Amount == "" || body.TxID == uuid.Nil || body.RoundID == uuid.Nil {
		return ParsedRequestResult[WinRequest]{
			OK:    false,
			Error: &ErrorResponseWithCodeAndAction{Message: "Invalid request body"},
		}
	}
	if !body.Type.Valid() {
		return ParsedRequestResult[WinRequest]{
			OK:    false,
			Error: &ErrorResponseWithCodeAndAction{Message: "Invalid request body"},
		}
	}

	return ParsedRequestResult[WinRequest]{OK: true, Body: body}
}

// ParseAndVerifyRefundRequest parses, verifies, and validates an incoming /refund request.
func (c *Client) ParseAndVerifyRefundRequest(rawBody string, signatureHeader string) ParsedRequestResult[RefundRequest] {
	var body RefundRequest
	result := c.parseAndVerifyInboundRequest(rawBody, signatureHeader, &body)
	if !result.OK {
		return ParsedRequestResult[RefundRequest]{OK: false, Error: result.Error}
	}

	// validate required fields
	if body.SessionID == "" || body.Amount == "" || body.TxID == uuid.Nil || body.OrigTxID == uuid.Nil {
		return ParsedRequestResult[RefundRequest]{
			OK:    false,
			Error: &ErrorResponseWithCodeAndAction{Message: "Invalid request body"},
		}
	}

	return ParsedRequestResult[RefundRequest]{OK: true, Body: body}
}

// ParseAndVerifyBalanceRequest parses, verifies, and validates an incoming GET /balance request.
func (c *Client) ParseAndVerifyBalanceRequest(queryParams map[string]string, signatureHeader string) ParsedBalanceResult {
	if !c.VerifyBalanceRequest(queryParams, signatureHeader) {
		return ParsedBalanceResult{
			OK:    false,
			Error: &ErrorResponseWithCodeAndAction{Message: "Invalid signature"},
		}
	}

	sessionID, ok := queryParams["sessionID"]
	if !ok || sessionID == "" {
		return ParsedBalanceResult{
			OK:    false,
			Error: &ErrorResponseWithCodeAndAction{Message: "Missing sessionID"},
		}
	}

	return ParsedBalanceResult{
		OK:    true,
		Query: &GetBalanceQuery{SessionID: sessionID},
	}
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

type parseResult struct {
	OK    bool
	Error *ErrorResponseWithCodeAndAction
}

func (c *Client) parseAndVerifyInboundRequest(rawBody string, signatureHeader string, target any) parseResult {
	if !c.verifyInboundSignature(rawBody, signatureHeader) {
		return parseResult{
			OK:    false,
			Error: &ErrorResponseWithCodeAndAction{Message: "Invalid signature"},
		}
	}

	if err := json.Unmarshal([]byte(rawBody), target); err != nil {
		return parseResult{
			OK:    false,
			Error: &ErrorResponseWithCodeAndAction{Message: "Invalid request body"},
		}
	}

	return parseResult{OK: true}
}
