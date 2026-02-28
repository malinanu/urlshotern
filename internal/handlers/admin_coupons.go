package handlers

import (
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

func (h *AdminCouponHandlers) CreateCoupon(c *gin.Context) {
	var coupon models.Coupon
	if err := c.ShouldBindJSON(&coupon); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	// Sync with Polar if it's a monetary discount
	if coupon.DiscountType == "percentage" || coupon.DiscountType == "fixed" {
		codeStr := coupon.Code
		payload := services.CreateDiscountPayload{
			Name:     coupon.Name,
			Code:     &codeStr,
			Type:     coupon.DiscountType, // "percentage" or "fixed"
			Duration: "once",
		}

		if coupon.DiscountType == "percentage" {
			bp := int(coupon.DiscountValue * 100)
			payload.BasisPoints = &bp
		} else if coupon.DiscountType == "fixed" {
			amt := int(coupon.DiscountValue * 100) // Assumes USD dollars -> cents
			usd := "usd"
			payload.Amount = &amt
			payload.Currency = &usd
		}

		if coupon.MaxRedemptions > 0 {
			payload.MaxRedemptions = &coupon.MaxRedemptions
		}

		polarDiscount, err := h.polarService.CreateDiscount(payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create discount in Polar dashboard", "details": err.Error()})
			return
		}

		// Save the Polar discount ID locally so we can reference it
		coupon.PolarDiscountID = polarDiscount.ID
	}

	if err := h.db.Create(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"coupon": coupon})
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
