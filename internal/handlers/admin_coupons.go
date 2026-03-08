package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type AdminCouponHandlers struct {
	db           *gorm.DB
	polarService *services.PolarService
}

func NewAdminCouponHandlers(db *gorm.DB, polarService *services.PolarService) *AdminCouponHandlers {
	return &AdminCouponHandlers{db: db, polarService: polarService}
}

func (h *AdminCouponHandlers) ListCoupons(c *gin.Context) {
	var coupons []models.Coupon
	if err := h.db.Order("created_at desc").Find(&coupons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list coupons", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"coupons": coupons})
}

// buildPolarPayload constructs the Polar CreateDiscountPayload from a coupon.
// Returns nil for unrecognised discount types.
func buildPolarPayload(coupon *models.Coupon) *services.CreateDiscountPayload {
	if coupon.DiscountType != "percentage" && coupon.DiscountType != "fixed" {
		return nil
	}
	code := coupon.Code
	payload := &services.CreateDiscountPayload{
		Name:     coupon.Name,
		Code:     &code,
		Type:     coupon.DiscountType,
		Duration: "once",
	}
	if coupon.DiscountType == "percentage" {
		// DiscountValue is 1-100 (e.g. 10 = 10%); Polar wants basis_points (1000 = 10%)
		bp := coupon.DiscountValue * 100
		payload.BasisPoints = &bp
	} else {
		// DiscountValue is dollar amount; Polar wants cents
		amt := coupon.DiscountValue * 100
		currency := "usd"
		payload.Amount = &amt
		payload.Currency = &currency
	}
	if coupon.MaxRedemptions != nil && *coupon.MaxRedemptions > 0 {
		payload.MaxRedemptions = coupon.MaxRedemptions
	}
	return payload
}

func (h *AdminCouponHandlers) CreateCoupon(c *gin.Context) {
	var coupon models.Coupon
	if err := c.ShouldBindJSON(&coupon); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	// Attempt Polar sync for monetary discount types; surface warning but don't block save
	var polarWarning string
	if payload := buildPolarPayload(&coupon); payload != nil {
		polarDiscount, err := h.polarService.CreateDiscount(*payload)
		if err != nil {
			polarWarning = fmt.Sprintf("Coupon saved locally but Polar sync failed: %s — use 'Sync to Polar' to retry.", err.Error())
			log.Printf("[Coupon] Polar sync failed for %s: %v", coupon.Code, err)
		} else {
			coupon.PolarDiscountID = &polarDiscount.ID
			log.Printf("[Coupon] Polar discount created for %s: %s", coupon.Code, polarDiscount.ID)
		}
	}

	if err := h.db.Create(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon", "details": err.Error()})
		return
	}

	resp := gin.H{"coupon": coupon}
	if polarWarning != "" {
		resp["polar_warning"] = polarWarning
	}
	c.JSON(http.StatusOK, resp)
}

// SyncToPolar retries syncing a coupon to Polar and updates polar_discount_id.
func (h *AdminCouponHandlers) SyncToPolar(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coupon ID"})
		return
	}

	var coupon models.Coupon
	if err := h.db.First(&coupon, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
		return
	}

	payload := buildPolarPayload(&coupon)
	if payload == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon type does not support Polar sync"})
		return
	}

	polarDiscount, err := h.polarService.CreateDiscount(*payload)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Polar sync failed", "details": err.Error()})
		return
	}

	if err := h.db.Model(&coupon).Update("polar_discount_id", polarDiscount.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save polar_discount_id"})
		return
	}

	coupon.PolarDiscountID = &polarDiscount.ID
	c.JSON(http.StatusOK, gin.H{"coupon": coupon, "polar_discount_id": polarDiscount.ID})
}

func (h *AdminCouponHandlers) DeleteCoupon(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coupon ID"})
		return
	}

	if err := h.db.Delete(&models.Coupon{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete coupon", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminCouponHandlers) ToggleCoupon(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coupon ID"})
		return
	}

	var coupon models.Coupon
	if err := h.db.First(&coupon, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
		return
	}

	coupon.IsActive = !coupon.IsActive
	if err := h.db.Save(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle coupon status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"coupon": coupon})
}
