package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type PaymentHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentHTTPClient(baseURL string, client *http.Client) *PaymentHTTPClient {
	return &PaymentHTTPClient{
		baseURL: baseURL,
		client:  client,
	}
}

type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

func (p *PaymentHTTPClient) AuthorizePayment(ctx context.Context, orderID string, amount int64) (string, string, error) {
	reqBody := paymentRequest{
		OrderID: orderID,
		Amount:  amount,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/payments", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return "", "", fmt.Errorf("payment service unavailable")
	}

	var result paymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	return result.Status, result.TransactionID, nil
}
