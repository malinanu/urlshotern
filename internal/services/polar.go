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
	"net/url"
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
	ID             string                 `json:"id"`
	Status         string                 `json:"status"` // open, expired, confirmed
	ClientSecret   string                 `json:"client_secret"`
	URL            string                 `json:"url"`
	CustomerID     *string                `json:"customer_id,omitempty"`
	CustomerEmail  *string                `json:"customer_email,omitempty"`
	ProductID      string                 `json:"product_id"`
	ProductPriceID *string                `json:"product_price_id,omitempty"`
	SuccessURL     string                 `json:"success_url"`
	Amount         *int                   `json:"amount,omitempty"`
	Currency       *string                `json:"currency,omitempty"`
	SubscriptionID *string                `json:"subscription_id,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	ExpiresAt      time.Time              `json:"expires_at"`
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

	endpoint = "/checkouts"
	payload = map[string]interface{}{
		"customer_email": customerEmail,
		"success_url":    successURL,
	}

	if productPriceID != nil && *productPriceID != "" {
		// Pass specific price ID, Polar API v1 allows product_price_id at the root
		payload["product_price_id"] = *productPriceID
	} else {
		// Use product ID
		payload["product_id"] = productID
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

// ListSubscriptionsByCustomer fetches active subscriptions for a Polar customer ID.
func (ps *PolarService) ListSubscriptionsByCustomer(customerID string) ([]PolarSubscription, error) {
	type SubsResponse struct {
		Items []PolarSubscription `json:"items"`
	}
	var response SubsResponse
	endpoint := fmt.Sprintf("/subscriptions?customer_id=%s&status=active&limit=5", customerID)
	if err := ps.makeRequest("GET", endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by customer: %w", err)
	}
	return response.Items, nil
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

// PolarDiscount represents a created discount in Polar
type PolarDiscount struct {
	ID               string     `json:"id"`
	OrganizationID   string     `json:"organization_id"`
	Name             string     `json:"name"`
	Code             *string    `json:"code,omitempty"`
	Type             string     `json:"type"`                   // "percentage" or "fixed"
	Amount           *int       `json:"amount,omitempty"`       // for fixed type
	Currency         *string    `json:"currency,omitempty"`     // for fixed type
	BasisPoints      *int       `json:"basis_points,omitempty"` // for percentage type
	Duration         string     `json:"duration"`               // "once", "forever", "repeating"
	DurationInMonths *int       `json:"duration_in_months,omitempty"`
	MaxRedemptions   *int       `json:"max_redemptions,omitempty"`
	StartsAt         *time.Time `json:"starts_at,omitempty"`
	EndsAt           *time.Time `json:"ends_at,omitempty"`
}

// CreateDiscountPayload represents the required fields to create a discount in Polar
type CreateDiscountPayload struct {
	Name             string     `json:"name"`
	Code             *string    `json:"code,omitempty"`
	Type             string     `json:"type"`                   // "percentage" or "fixed"
	BasisPoints      *int       `json:"basis_points,omitempty"` // 1000 = 10%
	Amount           *int       `json:"amount,omitempty"`       // 1000 = $10.00
	Currency         *string    `json:"currency,omitempty"`
	Duration         string     `json:"duration"` // "once", "forever", "repeating"
	DurationInMonths *int       `json:"duration_in_months,omitempty"`
	MaxRedemptions   *int       `json:"max_redemptions,omitempty"`
	Products         []string   `json:"products,omitempty"` // Make sure this is an array of IDs
	StartsAt         *time.Time `json:"starts_at,omitempty"`
	EndsAt           *time.Time `json:"ends_at,omitempty"`
}

// CreateDiscount creates a new discount code in Polar Organization
func (ps *PolarService) CreateDiscount(payload CreateDiscountPayload) (*PolarDiscount, error) {
	// For API requests to /v1/discounts/, the payload shape must match exactly.
	endpoint := "/discounts/"

	log.Printf("[Polar] Creating discount: %+v", payload)

	var discount PolarDiscount
	err := ps.makeRequest("POST", endpoint, payload, &discount)
	if err != nil {
		return nil, fmt.Errorf("failed to create discount in polar: %w", err)
	}

	return &discount, nil
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

// FindCustomerByEmail looks up a Polar customer by email address
func (ps *PolarService) FindCustomerByEmail(email string) (*PolarCustomer, error) {
	type Resp struct {
		Items []PolarCustomer `json:"items"`
	}
	var r Resp
	if err := ps.makeRequest("GET", "/customers?email="+url.QueryEscape(email)+"&limit=1", nil, &r); err != nil {
		return nil, err
	}
	if len(r.Items) == 0 {
		return nil, fmt.Errorf("no Polar customer for email %s", email)
	}
	return &r.Items[0], nil
}

// Database Operations

// GetOrCreatePolarCustomer gets or creates a Polar customer for a user
func (ps *PolarService) GetOrCreatePolarCustomer(userID int64, email string, name *string) (string, error) {
	// Check if user already has a Polar subscription with customer ID
	var subscription models.PolarSubscription
	err := ps.db.Where("user_id = ?", userID).Order("created_at DESC").First(&subscription).Error
	if err == nil && subscription.PolarCustomerID != "" && !strings.HasPrefix(subscription.PolarCustomerID, "free_user_") {
		return subscription.PolarCustomerID, nil
	}

	// Create new customer in Polar
	metadata := map[string]interface{}{
		"user_id": userID,
	}
	customer, err := ps.CreateCustomer(email, name, metadata)
	if err != nil {
		// If the customer already exists in Polar (e.g. after switching orgs),
		// fall back to looking them up by email.
		if strings.Contains(err.Error(), "already exists") {
			existing, lookupErr := ps.FindCustomerByEmail(email)
			if lookupErr != nil {
				return "", fmt.Errorf("failed to get/create customer: %w (lookup also failed: %v)", err, lookupErr)
			}
			return existing.ID, nil
		}
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
		planType := ps.DeterminePlanType(product)

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

// DeterminePlanType uses multiple strategies to determine the plan type from a Polar product
func (ps *PolarService) DeterminePlanType(product PolarProduct) string {
	// Strategy 1: Check metadata.plan_type (most reliable if set)
	if product.Metadata != nil {
		if planType, ok := product.Metadata["plan_type"].(string); ok && planType != "" {
			log.Printf("[Polar] Plan type for %s from metadata: %s", product.Name, planType)
			return planType
		}
	}

	// Strategy 2: Match against known product IDs
	switch product.ID {
	case "01a91192-1240-4735-9ac8-55a41988ba2c", // Premium (monthly)
		"c4009092-2fae-40f3-bf93-ab4eec754d72": // Premium Yearly
		log.Printf("[Polar] Plan type for %s matched by ID: premium", product.Name)
		return "premium"
	case "9b551ce6-f962-4b54-ae74-15b3c5b39cb6", // Growth (monthly)
		"3525a2c9-b4b9-4912-8a62-74125b1fd7ea": // Growth Yearly
		log.Printf("[Polar] Plan type for %s matched by ID: growth", product.Name)
		return "growth"
	case "ba93fc32-a360-4929-aaeb-d35fef1ae6ed", // Enterprise (monthly)
		"ece4b9ae-fe9b-4021-b207-9d063d0b3c52": // Enterprise Yearly
		log.Printf("[Polar] Plan type for %s matched by ID: enterprise", product.Name)
		return "enterprise"
	case "8a0a6d12-cc54-4af6-a5f5-74098d49fd6a", // Free (monthly)
		"9f4fa80f-969d-4546-a9ad-86b0457fcf84": // Free Yearly
		log.Printf("[Polar] Plan type for %s matched by ID: free", product.Name)
		return "free"
	}

	// Strategy 3: Infer from product name
	nameLower := strings.ToLower(product.Name)
	if strings.Contains(nameLower, "free") {
		log.Printf("[Polar] Plan type for %s inferred from name: free", product.Name)
		return "free"
	} else if strings.Contains(nameLower, "growth") || strings.Contains(nameLower, "team") {
		log.Printf("[Polar] Plan type for %s inferred from name: growth", product.Name)
		return "growth"
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

// planTier maps plan types to a numeric tier for comparison.
var planTier = map[string]int{"free": 0, "premium": 1, "growth": 2, "enterprise": 3}

// SyncUserSubscriptionFromPolar fetches the active subscriptions from Polar for the user's
// customer ID, picks the highest-tier one (handles simultaneous active subscriptions during
// an upgrade), upserts it locally, and syncs users.account_type.
func (ps *PolarService) SyncUserSubscriptionFromPolar(userID int64) error {
	// 1. Get polar_customer_id from any local row (newest first)
	var anyLocalSub models.PolarSubscription
	if err := ps.db.Where("user_id = ?", userID).Order("created_at DESC").First(&anyLocalSub).Error; err != nil {
		return fmt.Errorf("no local subscription row for user %d: %w", userID, err)
	}

	customerID := anyLocalSub.PolarCustomerID
	if customerID == "" || strings.HasPrefix(customerID, "free_user_") {
		// Look up real Polar customer by the user's email
		var user models.User
		if err := ps.db.Where("id = ?", userID).First(&user).Error; err != nil {
			return fmt.Errorf("user %d not found: %w", userID, err)
		}
		customer, err := ps.FindCustomerByEmail(user.Email)
		if err != nil {
			log.Printf("[SyncPolar] No real Polar customer for user %d (%s): %v", userID, user.Email, err)
			return nil // genuinely free, nothing to sync
		}
		customerID = customer.ID
	}

	// 2. Fetch all active subscriptions for this customer from Polar
	subs, err := ps.ListSubscriptionsByCustomer(customerID)
	if err != nil || len(subs) == 0 {
		// Polar unreachable or no active subs — fall back to syncing by stored subscription ID
		log.Printf("[SyncPolar] Customer lookup returned 0 subs or error for user %d: %v — trying stored ID", userID, err)
		return ps.syncByStoredID(userID, &anyLocalSub)
	}

	// 3. Among all active subscriptions, pick the highest-tier one.
	// After a plan upgrade Polar may keep both the old and new subscription active
	// simultaneously until the old billing period ends.
	bestSub := subs[0]
	bestPlan := "free"
	if product, err := ps.GetProduct(bestSub.ProductID); err == nil {
		bestPlan = ps.DeterminePlanType(*product)
	}
	for _, s := range subs[1:] {
		product, err := ps.GetProduct(s.ProductID)
		if err != nil {
			continue
		}
		pt := ps.DeterminePlanType(*product)
		if planTier[pt] > planTier[bestPlan] {
			bestSub = s
			bestPlan = pt
		}
	}

	log.Printf("[SyncPolar] Best active subscription for user %d: sub=%s plan=%s", userID, bestSub.ID, bestPlan)

	// 4. Derive billing_cycle
	var billingCycle *string
	if bestSub.PriceID != nil {
		if product, err := ps.GetProduct(bestSub.ProductID); err == nil {
			for _, price := range product.Prices {
				if price.ID == *bestSub.PriceID {
					billingCycle = price.RecurringInterval
					break
				}
			}
		}
	}

	// 5. Upsert: find or create a local row for this Polar subscription ID
	subID := bestSub.ID
	var existing models.PolarSubscription
	findErr := ps.db.Where("polar_subscription_id = ?", subID).First(&existing).Error
	if findErr != nil {
		newSub := models.PolarSubscription{
			UserID:              userID,
			PolarCustomerID:     customerID,
			PolarSubscriptionID: &subID,
			PolarProductID:      bestSub.ProductID,
			PolarPriceID:        bestSub.PriceID,
			Status:              bestSub.Status,
			PlanType:            bestPlan,
			BillingCycle:        billingCycle,
			CurrentPeriodStart:  bestSub.CurrentPeriodStart,
			CurrentPeriodEnd:    bestSub.CurrentPeriodEnd,
			CancelAtPeriodEnd:   bestSub.CancelAtPeriodEnd,
		}
		if err := ps.db.Create(&newSub).Error; err != nil {
			return fmt.Errorf("failed to create subscription row: %w", err)
		}
		log.Printf("[SyncPolar] Created new row for user %d (sub %s, plan %s)", userID, subID, bestPlan)
	} else {
		if err := ps.db.Model(&existing).Updates(map[string]interface{}{
			"polar_product_id":     bestSub.ProductID,
			"polar_price_id":       bestSub.PriceID,
			"status":               bestSub.Status,
			"plan_type":            bestPlan,
			"billing_cycle":        billingCycle,
			"current_period_start": bestSub.CurrentPeriodStart,
			"current_period_end":   bestSub.CurrentPeriodEnd,
			"updated_at":           time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("failed to update subscription row: %w", err)
		}
		log.Printf("[SyncPolar] Updated row for user %d (sub %s, plan %s)", userID, subID, bestPlan)
	}

	// 6. Sync account_type
	ps.db.Exec("UPDATE users SET account_type = $1, updated_at = NOW() WHERE id = $2", bestPlan, userID)
	return nil
}

// syncByStoredID is the fallback: sync using the locally stored polar_subscription_id.
func (ps *PolarService) syncByStoredID(userID int64, localSub *models.PolarSubscription) error {
	if localSub.PolarSubscriptionID == nil {
		return fmt.Errorf("no polar_subscription_id stored for user %d", userID)
	}
	polarSub, err := ps.GetSubscription(*localSub.PolarSubscriptionID)
	if err != nil {
		return fmt.Errorf("GetSubscription failed for user %d: %w", userID, err)
	}

	var billingCycle *string
	if polarSub.PriceID != nil {
		if product, err := ps.GetProduct(polarSub.ProductID); err == nil {
			for _, price := range product.Prices {
				if price.ID == *polarSub.PriceID {
					billingCycle = price.RecurringInterval
					break
				}
			}
		}
	}

	planType := localSub.PlanType
	if product, err := ps.GetProduct(polarSub.ProductID); err == nil {
		planType = ps.DeterminePlanType(*product)
	}

	if err := ps.db.Model(localSub).Updates(map[string]interface{}{
		"polar_product_id":     polarSub.ProductID,
		"polar_price_id":       polarSub.PriceID,
		"status":               polarSub.Status,
		"plan_type":            planType,
		"billing_cycle":        billingCycle,
		"current_period_start": polarSub.CurrentPeriodStart,
		"current_period_end":   polarSub.CurrentPeriodEnd,
		"updated_at":           time.Now(),
	}).Error; err != nil {
		return err
	}
	ps.db.Exec("UPDATE users SET account_type = $1, updated_at = NOW() WHERE id = $2", planType, userID)
	return nil
}
