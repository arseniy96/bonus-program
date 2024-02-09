package accrual

import "net/url"

const (
	OrderStatusRegistered = "REGISTERED"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusProcessed  = "PROCESSED"
	GetOrderPath          = "/api/orders/"
)

func CheckOrder(host, orderNumber string) (*GetOrderResponse, error) {
	fullURL, err := url.JoinPath(host, GetOrderPath, orderNumber)
	if err != nil {
		return nil, err
	}
	client := InitClient()

	return client.GetOrderRequest(fullURL)
}
