package transfer

import "time"

type SubscriptionEvent struct {
	ID        string `json:"id"`
	EventType string `json:"eventType"`
	CreatedAt int64  `json:"created_at"`
	Object    struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Product struct {
			ID                string      `json:"id"`
			Name              string      `json:"name"`
			Description       string      `json:"description"`
			ImageURL          interface{} `json:"image_url"`
			Price             int         `json:"price"`
			Currency          string      `json:"currency"`
			BillingType       string      `json:"billing_type"`
			BillingPeriod     string      `json:"billing_period"`
			Status            string      `json:"status"`
			TaxMode           string      `json:"tax_mode"`
			TaxCategory       string      `json:"tax_category"`
			DefaultSuccessURL string      `json:"default_success_url"`
			CreatedAt         time.Time   `json:"created_at"`
			UpdatedAt         time.Time   `json:"updated_at"`
			Mode              string      `json:"mode"`
		} `json:"product"`
		Customer struct {
			ID        string    `json:"id"`
			Object    string    `json:"object"`
			Email     string    `json:"email"`
			Name      string    `json:"name"`
			Country   string    `json:"country"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Mode      string    `json:"mode"`
		} `json:"customer"`
		CollectionMethod       string      `json:"collection_method"`
		Status                 string      `json:"status"`
		LastTransactionID      string      `json:"last_transaction_id"`
		LastTransactionDate    time.Time   `json:"last_transaction_date"`
		NextTransactionDate    time.Time   `json:"next_transaction_date"`
		CurrentPeriodStartDate time.Time   `json:"current_period_start_date"`
		CurrentPeriodEndDate   time.Time   `json:"current_period_end_date"`
		CanceledAt             interface{} `json:"canceled_at"`
		CreatedAt              time.Time   `json:"created_at"`
		UpdatedAt              time.Time   `json:"updated_at"`
		Metadata               struct {
			CustomData         string `json:"custom_data"`
			InternalCustomerID string `json:"internal_customer_id"`
		} `json:"metadata"`
		Mode string `json:"mode"`
	} `json:"object"`
}
