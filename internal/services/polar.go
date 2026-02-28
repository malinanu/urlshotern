package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"gorm.io/gorm"
)

// PolarService handles all interactions with Polar API
type PolarService struct {
	db      *gorm.DB
	config  *configs.Config
	client  *http.Client
	baseURL string
}

// NewPolarService creates a new Polar service instance
func NewPolarService(db *gorm.DB, config *configs.Config) *PolarService {
	return &PolarService{
		db:      db,
		config:  config,
		baseURL: config.PolarAPIURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Polar API Response structures

type PolarCustomer struct {
	ID             string                 `json:"id"`
	Email          string                 `json:"email"`
	Name           *string                `json:"name,omitempty"`
	OrganizationID string                 `json:"organization_id"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

type PolarCheckoutSession struct {
	ID                 string                 `json:"id"`
	Status             string                 `json:"status"` // open, expired, confirmed
	ClientSecret       string                 `json:"client_secret"`
	URL                string                 `json:"url"`
	CustomerID         *string                `json:"customer_id,omitempty"`
	CustomerEmail      *string                `json:"customer_email,omitempty"`
	ProductID          string                 `json:"product_id"`
	ProductPriceID     *string                `json:"product_price_id,omitempty"`
	SuccessURL         string                 `json:"success_url"`
	Amount             *int                   `json:"amount,omitempty"`
	Currency           *string                `json:"currency,omitempty"`
	SubscriptionID     *string                `json:"subscription_id,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	ExpiresAt          time.Time              `json:"expires_at"`
}

type PolarSubscription struct {
	ID                 string                 `json:"id"`
	Status             string                 `json:"status"`
	CurrentPeriodStart *time.Time             `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time             `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd  bool                   `json:"cancel_at_period_end"`
	StartedAt          *time.Time             `json:"started_at,omitempty"`
	EndedAt            *time.Time             `json:"ended_at,omitempty"`
	CustomerID         string                 `json:"customer_id"`
	ProductID          string                 `json:"product_id"`
	PriceID            *string                `json:"price_id,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}

type PolarProduct struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    *string                `json:"description,omitempty"`
	IsRecurring    bool                   `json:"is_recurring"`
	IsArchived     bool                   `json:"is_archived"`
	OrganizationID string                 `json:"organization_id"`
	Prices         []PolarPrice           `json:"prices"`
	Benefits       []PolarBenefit         `json:"benefits"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

type PolarPrice struct {
	ID                string  `json:"id"`
	PriceAmount       int     `json:"price_amount"`
	PriceCurrency     string  `json:"price_currency"`
	Type              string  `json:"type"` // one_time, recurring
	RecurringInterval *string `json:"recurring_interval,omitempty"`
	IsArchived        bool    `json:"is_archived"`
}

type PolarBenefit struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Selectable  bool                   `json:"selectable"`
	Deletable   bool                   `json:"deletable"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// API Methods

// CreateCustomer creates a new customer in Polar
func (ps *PolarService) CreateCustomer(email string, name *string, metadata map[string]interface{}) (*PolarCustomer, error) {
	// Note: organization_id is automatically set by the organization token
	// and should NOT be included in the payload when using organization tokens
	payload := map[string]interface{}{
		"email": email,
	}
	if name != nil {
		payload["name"] = *name
	}
	if metadata != nil {
		payload["metadata"] = metadata
	}

	var customer PolarCustomer
	err := ps.makeRequest("POST", "/customers", payload, &customer)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return &customer, nil
}

// GetCustomer retrieves a customer by ID
func (ps *PolarService) GetCustomer(customerID string) (*PolarCustomer, error) {
	var customer PolarCustomer
	err := ps.makeRequest("GET", fmt.Sprintf("/customers/%s", customerID), nil, &customer)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	return &customer, nil
}

// CreateCheckoutSession creates a new checkout session
func (ps *PolarService) CreateCheckoutSession(productID string, productPriceID *string, discountID *string, customerEmail string, successURL string, metadata map[string]interface{}) (*PolarCheckoutSession, error) {
	var endpoint string
	var payload map[string]interface{}

	if productPriceID != nil && *productPriceID != "" {
		// Use /checkouts/custom endpoint for specific price ID
		endpoint = "/checkouts/custom"
		payload = map[string]interface{}{
			"product_price_id": *productPriceID,
			"customer_email":   customerEmail,
			"success_url":      successURL,
		}
	} else {
		// Use /checkouts endpoint with product ID array
		endpoint = "/checkouts"
		payload = map[string]interface{}{
			"products":       []string{productID}, // Polar expects array of product ID strings
			"customer_email": customerEmail,
			"success_url":    successURL,
		}
	}

	if discountID != nil && *discountID != "" {
		payload["discount_id"] = *discountID
	}

	if metadata != nil {
		payload["metadata"] = metadata
	}

	log.Printf("[Polar] Creating checkout with endpoint: %s, payload: %+v", endpoint, payload)

	var checkout PolarCheckoutSession
	err := ps.makeRequest("POST", endpoint, payload, &checkout)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return &checkout, nil
}

// GetCheckoutSession retrieves a checkout session by ID
func (ps *PolarService) GetCheckoutSession(checkoutID string) (*PolarCheckoutSession, error) {
	var checkout PolarCheckoutSession
	err := ps.makeRequest("GET", fmt.Sprintf("/checkouts/%s", checkoutID), nil, &checkout)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkout session: %w", err)
	}
	return &checkout, nil
}

// GetSubscription retrieves a subscription by ID
func (ps *PolarService) GetSubscription(subscriptionID string) (*PolarSubscription, error) {
	var subscription PolarSubscription
	err := ps.makeRequest("GET", fmt.Sprintf("/subscriptions/%s", subscriptionID), nil, &subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &subscription, nil
}

// CancelSubscription cancels a subscription
func (ps *PolarService) CancelSubscription(subscriptionID string, cancelAtPeriodEnd bool) error {
	payload := map[string]interface{}{
		"cancel_at_period_end": cancelAtPeriodEnd,
	}

	err := ps.makeRequest("PATCH", fmt.Sprintf("/subscriptions/%s", subscriptionID), payload, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}
	return nil
}

// ListProducts retrieves all products for the organization
func (ps *PolarService) ListProducts() ([]PolarProduct, error) {
	type ProductsResponse struct {
		Items []PolarProduct `json:"items"`
	}

	var response ProductsResponse
	// Note: Polar API requires trailing slash on /products/ endpoint
	// The organization_id is inferred from the auth token, no need to pass it
	err := ps.makeRequest("GET", "/products/", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return response.Items, nil
}

// GetProduct retrieves a product by ID
func (ps *PolarService) GetProduct(productID string) (*PolarProduct, error) {
	var product PolarProduct
	err := ps.makeRequest("GET", fmt.Sprintf("/products/%s", productID), nil, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

// PolarOrder represents an order/invoice from Polar
type PolarOrder struct {
	ID             string    `json:"id"`
	Amount         int       `json:"amount"`
	Currency       string    `json:"currency"`
	CustomerID     string    `json:"customer_id"`
	ProductID      string    `json:"product_id"`
	ProductPriceID string    `json:"product_price_id"`
	SubscriptionID *string   `json:"subscription_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// ListOrders retrieves all orders for a customer
func (ps *PolarService) ListOrders(customerID string, limit int) ([]PolarOrder, error) {
	type OrdersResponse struct {
		Items []PolarOrder `json:"items"`
	}

	if limit <= 0 {
		limit = 20
	}

	var response OrdersResponse
	endpoint := fmt.Sprintf("/orders?customer_id=%s&limit=%d", customerID, limit)
	err := ps.makeRequest("GET", endpoint, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	return response.Items, nil
}

// GetOrder retrieves a specific order by ID
func (ps *PolarService) GetOrder(orderID string) (*PolarOrder, error) {
	var order PolarOrder
	err := ps.makeRequest("GET", fmt.Sprintf("/orders/%s", orderID), nil, &order)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

// Database Operations

// GetOrCreatePolarCustomer gets or creates a Polar customer for a user
func (ps *PolarService) GetOrCreatePolarCustomer(userID int64, email string, name *string) (string, error) {
	// Check if user already has a Polar subscription with customer ID
	var subscription models.PolarSubscription
	err := ps.db.Where("user_id = ?", userID).Order("created_at DESC").First(&subscription).Error
	if err == nil && subscription.PolarCustomerID != "" {
		return subscription.PolarCustomerID, nil
	}

	// Create new customer in Polar
	metadata := map[string]interface{}{
		"user_id": userID,
	}
	customer, err := ps.CreateCustomer(email, name, metadata)
	if err != nil {
		return "", err
	}

	return customer.ID, nil
}

// SyncProductsFromPolar syncs products from Polar to local database
func (ps *PolarService) SyncProductsFromPolar() error {
	products, err := ps.ListProducts()
	if err != nil {
		return err
	}

	for _, product := range products {
		// Convert Polar product to local model
		prices := make([]models.PolarPrice, len(product.Prices))
		for i, p := range product.Prices {
			prices[i] = models.PolarPrice{
				ID:                p.ID,
				PriceAmount:       p.PriceAmount,
				PriceCurrency:     p.PriceCurrency,
				RecurringInterval: stringValue(p.RecurringInterval),
				IsArchived:        p.IsArchived,
			}
		}

		benefits := make([]models.PolarBenefit, len(product.Benefits))
		for i, b := range product.Benefits {
			benefitProps, _ := json.Marshal(b.Properties)
			var propsJSON models.JSONB
			json.Unmarshal(benefitProps, &propsJSON)

			benefits[i] = models.PolarBenefit{
				ID:          b.ID,
				Type:        b.Type,
				Description: b.Description,
				Properties:  propsJSON,
			}
		}

		// Determine plan type using multi-strategy detection
		planType := ps.determinePlanType(product)

		localProduct := models.PolarProduct{
			ID:             product.ID,
			PolarProductID: product.ID,
			Name:           product.Name,
			Description:    stringValue(product.Description),
			PlanType:       planType,
			IsRecurring:    product.IsRecurring,
			Prices:         prices,
			Benefits:       benefits,
			OrganizationID: product.OrganizationID,
			IsArchived:     product.IsArchived,
			CreatedAt:      product.CreatedAt,
			LastSyncedAt:   time.Now(),
		}

		// Upsert product
		err = ps.db.Save(&localProduct).Error
		if err != nil {
			return fmt.Errorf("failed to save product %s: %w", product.ID, err)
		}
	}

	return nil
}

// determinePlanType uses multiple strategies to determine the plan type from a Polar product
func (ps *PolarService) determinePlanType(product PolarProduct) string {
	// Strategy 1: Check metadata.plan_type (most reliable if set)
	if product.Metadata != nil {
		if planType, ok := product.Metadata["plan_type"].(string); ok && planType != "" {
			log.Printf("[Polar] Plan type for %s from metadata: %s", product.Name, planType)
			return planType
		}
	}

	// Strategy 2: Match against known product IDs
	switch product.ID {
	case "1216afa1-650c-4dda-95cd-7010dbb03404":
		log.Printf("[Polar] Plan type for %s matched by ID: premium", product.Name)
		return "premium"
	case "f5ffabdf-e116-4aba-99f3-cdebe0f09486":
		log.Printf("[Polar] Plan type for %s matched by ID: enterprise", product.Name)
		return "enterprise"
	}

	// Strategy 3: Infer from product name
	nameLower := strings.ToLower(product.Name)
	if strings.Contains(nameLower, "free") {
		log.Printf("[Polar] Plan type for %s inferred from name: free", product.Name)
		return "free"
	} else if strings.Contains(nameLower, "pro") || strings.Contains(nameLower, "professional") {
		log.Printf("[Polar] Plan type for %s inferred from name: premium", product.Name)
		return "premium"
	} else if strings.Contains(nameLower, "ultra") || strings.Contains(nameLower, "enterprise") {
		log.Printf("[Polar] Plan type for %s inferred from name: enterprise", product.Name)
		return "enterprise"
	}

	// Strategy 4: Infer from price (if prices exist)
	if len(product.Prices) > 0 {
		price := product.Prices[0].PriceAmount
		if price == 0 {
			log.Printf("[Polar] Plan type for %s inferred from price $0: free", product.Name)
			return "free"
		} else if price >= 20000 { // $200+
			log.Printf("[Polar] Plan type for %s inferred from price $%.2f: enterprise", product.Name, float64(price)/100)
			return "enterprise"
		} else if price >= 500 { // $5+
			log.Printf("[Polar] Plan type for %s inferred from price $%.2f: premium", product.Name, float64(price)/100)
			return "premium"
		}
	}

	// Default: free (safest default)
	log.Printf("[Polar] Plan type for %s defaulting to: free", product.Name)
	return "free"
}

// SaveSubscription saves a Polar subscription to the database
func (ps *PolarService) SaveSubscription(subscription *models.PolarSubscription) error {
	return ps.db.Save(subscription).Error
}

// GetUserSubscription retrieves the active subscription for a user
func (ps *PolarService) GetUserSubscription(userID int64) (*models.PolarSubscription, error) {
	var subscription models.PolarSubscription
	err := ps.db.Where("user_id = ? AND (status = ? OR status = ?)", userID, "active", "trialing").
		Order("created_at DESC").
		First(&subscription).Error

	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

// Webhook Validation

// ValidateWebhookSignature validates the signature of a Polar webhook
func (ps *PolarService) ValidateWebhookSignature(payload []byte, signature string, timestamp string) bool {
	if ps.config.PolarWebhookSecret == "" {
		return false
	}

	// Polar uses HMAC SHA256 for webhook signatures
	// Format: timestamp.payload
	signedPayload := timestamp + "." + string(payload)

	mac := hmac.New(sha256.New, []byte(ps.config.PolarWebhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// Helper Methods

// makeRequest makes an HTTP request to the Polar API
func (ps *PolarService) makeRequest(method, endpoint string, payload interface{}, result interface{}) error {
	var body io.Reader
	var payloadStr string
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
		payloadStr = string(jsonData)
	}

	url := ps.baseURL + endpoint

	// Enhanced logging for diagnostics
	log.Printf("[Polar API] Request: %s %s", method, url)
	if payloadStr != "" {
		log.Printf("[Polar API] Payload: %s", payloadStr)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ps.config.PolarAccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := ps.client.Do(req)
	if err != nil {
		log.Printf("[Polar API] Request failed: %v", err)
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log response details
	log.Printf("[Polar API] Response: Status=%d, Body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		log.Printf("[Polar API] Error response: %s", string(respBody))
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		err = json.Unmarshal(respBody, result)
		if err != nil {
			log.Printf("[Polar API] Failed to unmarshal response: %v", err)
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// stringValue safely dereferences a string pointer
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
