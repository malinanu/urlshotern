package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminBillingHandlers handles admin billing operations
type AdminBillingHandlers struct {
	billingService *services.BillingService
	auditService   *services.AdminAuditService
}

// NewAdminBillingHandlers creates a new admin billing handlers instance
func NewAdminBillingHandlers(billingService *services.BillingService, auditService *services.AdminAuditService) *AdminBillingHandlers {
	return &AdminBillingHandlers{
		billingService: billingService,
		auditService:   auditService,
	}
}

// AdminListSubscriptions lists all subscriptions with filtering
// GET /api/v1/admin/billing/subscriptions
func (h *AdminBillingHandlers) AdminListSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse user_id filter
	var userID *int64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if uid, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = &uid
		}
	}

	// Build filter
	filter := models.SubscriptionFilter{
		UserID:    userID,
		Status:    c.Query("status"),
		PlanType:  c.Query("plan_type"),
		Search:    c.Query("search"),
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	subscriptions, total, err := h.billingService.AdminListSubscriptions(filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list subscriptions",
		})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"total":         total,
		"page":          page,
		"limit":         limit,
		"total_pages":   totalPages,
	})
}

// AdminGetSubscription retrieves a specific subscription
// GET /api/v1/admin/billing/subscriptions/:id
func (h *AdminBillingHandlers) AdminGetSubscription(c *gin.Context) {
	subscriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_subscription_id",
			Message: "Invalid subscription ID",
		})
		return
	}

	subscription, err := h.billingService.AdminGetSubscription(subscriptionID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "subscription_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": subscription,
	})
}

// AdminUpdateSubscription updates a subscription
// PUT /api/v1/admin/billing/subscriptions/:id
func (h *AdminBillingHandlers) AdminUpdateSubscription(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	subscriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_subscription_id",
			Message: "Invalid subscription ID",
		})
		return
	}

	var req models.AdminUpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.billingService.AdminUpdateSubscription(subscriptionID, &req, adminID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "subscription not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	subIDPtr := &subscriptionID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "subscription.update", "subscription", subIDPtr, map[string]interface{}{
		"updates": req,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription updated successfully",
	})
}

// AdminCancelSubscription cancels a subscription
// POST /api/v1/admin/billing/subscriptions/:id/cancel
func (h *AdminBillingHandlers) AdminCancelSubscription(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	subscriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_subscription_id",
			Message: "Invalid subscription ID",
		})
		return
	}

	var req struct {
		Reason    string `json:"reason"`
		Immediate bool   `json:"immediate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = ""
		req.Immediate = false
	}

	if err := h.billingService.AdminCancelSubscription(subscriptionID, req.Reason, req.Immediate, adminID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "subscription not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "cancellation_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	subIDPtr := &subscriptionID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "subscription.cancel", "subscription", subIDPtr, map[string]interface{}{
		"reason":    req.Reason,
		"immediate": req.Immediate,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription cancelled successfully",
	})
}

// AdminRefundSubscription issues a refund
// POST /api/v1/admin/billing/subscriptions/:id/refund
func (h *AdminBillingHandlers) AdminRefundSubscription(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	subscriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_subscription_id",
			Message: "Invalid subscription ID",
		})
		return
	}

	var req models.RefundSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.billingService.AdminRefundSubscription(subscriptionID, &req, adminID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "subscription not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "refund_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	subIDPtr := &subscriptionID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "subscription.refund", "subscription", subIDPtr, map[string]interface{}{
		"amount": req.Amount,
		"reason": req.Reason,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Refund issued successfully",
	})
}

// AdminListPlans lists all billing plans
// GET /api/v1/admin/billing/plans
func (h *AdminBillingHandlers) AdminListPlans(c *gin.Context) {
	includeArchived := c.DefaultQuery("include_archived", "false") == "true"

	products, err := h.billingService.AdminListPlans(includeArchived)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list plans",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": products,
		"total": len(products),
	})
}

// AdminCreatePlan creates a new billing plan
// POST /api/v1/admin/billing/plans
func (h *AdminBillingHandlers) AdminCreatePlan(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req models.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// TODO: Implement Polar API product creation
	// This requires calling PolarService.CreateProduct() which is not yet implemented
	_ = adminID

	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Plan creation via Polar API not yet implemented",
	})
}

// AdminUpdatePlan updates a billing plan
// PUT /api/v1/admin/billing/plans/:id
func (h *AdminBillingHandlers) AdminUpdatePlan(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	planID := c.Param("id")
	var req models.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// TODO: Implement Polar API product update
	// This requires calling PolarService.UpdateProduct() which is not yet implemented
	_ = adminID
	_ = planID

	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Plan update via Polar API not yet implemented",
	})
}

// AdminDeletePlan archives a billing plan
// DELETE /api/v1/admin/billing/plans/:id
func (h *AdminBillingHandlers) AdminDeletePlan(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	planID := c.Param("id")

	if err := h.billingService.AdminArchivePlan(planID, adminID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "plan not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "archive_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "plan.archive", "plan", nil, map[string]interface{}{
		"plan_id": planID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Plan archived successfully",
	})
}

// AdminListWebhookEvents lists webhook events
// GET /api/v1/admin/billing/webhooks
func (h *AdminBillingHandlers) AdminListWebhookEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build filter
	filter := make(map[string]interface{})
	if eventType := c.Query("event_type"); eventType != "" {
		filter["event_type"] = eventType
	}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	events, total, err := h.billingService.AdminListWebhookEvents(filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list webhook events",
		})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{
		"events":      events,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// AdminRetryWebhook retries a failed webhook
// POST /api/v1/admin/billing/webhooks/:id/retry
func (h *AdminBillingHandlers) AdminRetryWebhook(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	webhookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_webhook_id",
			Message: "Invalid webhook ID",
		})
		return
	}

	if err := h.billingService.AdminRetryWebhook(webhookID, adminID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "webhook event not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "retry_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	webhookIDPtr := &webhookID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "webhook.retry", "webhook", webhookIDPtr, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook retry initiated",
	})
}

// AdminSyncPolar manually syncs with Polar API
// POST /api/v1/admin/billing/polar/sync
func (h *AdminBillingHandlers) AdminSyncPolar(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	result, err := h.billingService.AdminSyncPolar(adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "sync_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "polar.sync", "system", nil, result)

	c.JSON(http.StatusOK, gin.H{
		"message": "Polar sync completed successfully",
		"result":  result,
	})
}

// AdminGetRevenue retrieves revenue analytics
// GET /api/v1/admin/billing/revenue
func (h *AdminBillingHandlers) AdminGetRevenue(c *gin.Context) {
	// Parse date range (default to current month)
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	revenue, err := h.billingService.AdminGetRevenue(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to calculate revenue",
		})
		return
	}

	c.JSON(http.StatusOK, revenue)
}
