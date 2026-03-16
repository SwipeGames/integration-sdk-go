package swipegames

import "testing"

func TestNewBalanceResponse(t *testing.T) {
	r := NewBalanceResponse("100.50")
	if r.Balance != "100.50" {
		t.Errorf("got %s, want 100.50", r.Balance)
	}
}

func TestNewBetResponse(t *testing.T) {
	r := NewBetResponse("90.50", "tx-1")
	if r.Balance != "90.50" || r.TxID != "tx-1" {
		t.Errorf("got %+v", r)
	}
}

func TestNewWinResponse(t *testing.T) {
	r := NewWinResponse("190.50", "tx-2")
	if r.Balance != "190.50" || r.TxID != "tx-2" {
		t.Errorf("got %+v", r)
	}
}

func TestNewRefundResponse(t *testing.T) {
	r := NewRefundResponse("100.50", "tx-3")
	if r.Balance != "100.50" || r.TxID != "tx-3" {
		t.Errorf("got %+v", r)
	}
}

func TestNewErrorResponse(t *testing.T) {
	t.Run("with all fields", func(t *testing.T) {
		r := NewErrorResponse(ErrorResponseOpts{
			Message:    "Insufficient funds",
			Code:       ErrorCodeInsufficientFunds,
			Action:     ErrorActionRefresh,
			ActionData: "some-data",
			Details:    "Balance: 0.00",
		})
		if r.Message != "Insufficient funds" {
			t.Errorf("message: got %s", r.Message)
		}
		if r.Code == nil || *r.Code != ErrorCodeInsufficientFunds {
			t.Errorf("code: got %v", r.Code)
		}
		if r.Action == nil || *r.Action != ErrorActionRefresh {
			t.Errorf("action: got %v", r.Action)
		}
		if r.ActionData == nil || *r.ActionData != "some-data" {
			t.Errorf("actionData: got %v", r.ActionData)
		}
		if r.Details == nil || *r.Details != "Balance: 0.00" {
			t.Errorf("details: got %v", r.Details)
		}
	})

	t.Run("with minimal fields", func(t *testing.T) {
		r := NewErrorResponse(ErrorResponseOpts{Message: "Server error"})
		if r.Message != "Server error" {
			t.Errorf("message: got %s", r.Message)
		}
		if r.Code != nil {
			t.Errorf("code should be nil, got %v", r.Code)
		}
		if r.Action != nil {
			t.Errorf("action should be nil, got %v", r.Action)
		}
	})
}
