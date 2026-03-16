package swipegames

// ErrorResponseOpts contains options for creating an error response.
type ErrorResponseOpts struct {
	Message    string
	Code       ErrorResponseWithCodeAndActionCode
	Action     ErrorResponseWithCodeAndActionAction
	ActionData string
	Details    string
}

// NewBalanceResponse creates a balance response.
func NewBalanceResponse(balance string) BalanceResponse {
	return BalanceResponse{Balance: balance}
}

// NewBetResponse creates a bet response.
func NewBetResponse(balance, txID string) BetResponse {
	return BetResponse{Balance: balance, TxID: txID}
}

// NewWinResponse creates a win response.
func NewWinResponse(balance, txID string) WinResponse {
	return WinResponse{Balance: balance, TxID: txID}
}

// NewRefundResponse creates a refund response.
func NewRefundResponse(balance, txID string) RefundResponse {
	return RefundResponse{Balance: balance, TxID: txID}
}

// NewErrorResponse creates an error response with optional code and action.
func NewErrorResponse(opts ErrorResponseOpts) ErrorResponseWithCodeAndAction {
	resp := ErrorResponseWithCodeAndAction{
		Message: opts.Message,
	}
	if opts.Code != "" {
		resp.Code = &opts.Code
	}
	if opts.Action != "" {
		resp.Action = &opts.Action
	}
	if opts.ActionData != "" {
		resp.ActionData = &opts.ActionData
	}
	if opts.Details != "" {
		resp.Details = &opts.Details
	}
	return resp
}
