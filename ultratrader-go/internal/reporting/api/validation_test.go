package api_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/reporting/api"
)

type TradeRequest struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
}

func (tr *TradeRequest) Validate() error {
	if tr.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if tr.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	return nil
}

func TestDecodeAndValidate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		expectErr  bool
		errMessage string
	}{
		{
			name:      "valid json",
			body:      `{"symbol":"BTC","quantity":1.5}`,
			expectErr: false,
		},
		{
			name:       "missing required field",
			body:       `{"symbol":"","quantity":1.5}`,
			expectErr:  true,
			errMessage: "symbol is required",
		},
		{
			name:       "invalid value",
			body:       `{"symbol":"BTC","quantity":-1.0}`,
			expectErr:  true,
			errMessage: "quantity must be greater than zero",
		},
		{
			name:       "unknown field mass assignment attempt",
			body:       `{"symbol":"BTC","quantity":1.5,"isAdmin":true}`,
			expectErr:  true,
			errMessage: "unknown field",
		},
		{
			name:       "badly formed json",
			body:       `{"symbol":"BTC","quantity":1.5,}`,
			expectErr:  true,
			errMessage: "badly-formed",
		},
		{
			name:       "empty body",
			body:       ``,
			expectErr:  true,
			errMessage: "must not be empty",
		},
		{
			name:       "multiple json objects",
			body:       `{"symbol":"BTC","quantity":1.5}{"symbol":"ETH","quantity":2.0}`,
			expectErr:  true,
			errMessage: "single JSON object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			var v TradeRequest
			err := api.DecodeAndValidate(w, req, &v)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if !strings.Contains(err.Error(), tt.errMessage) {
					t.Errorf("expected error containing %q, got %q", tt.errMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if v.Symbol != "BTC" || v.Quantity != 1.5 {
					t.Errorf("decoded values incorrect: %+v", v)
				}
			}
		})
	}
}
