package unit

import (
	"testing"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAttributionStorage extends MockPostgresStorage for attribution tests
type MockAttributionStorage struct {
	MockPostgresStorage
}

func (m *MockAttributionStorage) RecordTouchpoint(touchpoint *models.AttributionTouchpoint) error {
	args := m.Called(touchpoint)
	return args.Error(0)
}

func (m *MockAttributionStorage) GetConversionJourney(conversionID string) ([]*models.AttributionTouchpoint, error) {
	args := m.Called(conversionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AttributionTouchpoint), args.Error(1)
}

func (m *MockAttributionStorage) SaveAttribution(attribution *models.TouchpointAttribution) error {
	args := m.Called(attribution)
	return args.Error(0)
}

func (m *MockAttributionStorage) GetAttributionReport(shortCode string, model services.AttributionModel, days int) (*models.AttributionReport, error) {
	args := m.Called(shortCode, model, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AttributionReport), args.Error(1)
}

func (m *MockAttributionStorage) RecordConversion(conversion *models.Conversion) error {
	args := m.Called(conversion)
	return args.Error(0)
}

func (m *MockAttributionStorage) GetSessionTouchpoints(sessionID string) ([]*models.AttributionTouchpoint, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AttributionTouchpoint), args.Error(1)
}

// Mock ConversionTrackingService
type MockConversionTrackingService struct {
	mock.Mock
}

func (m *MockConversionTrackingService) TrackConversion(event *models.ConversionEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockConversionTrackingService) GetConversionMetrics(shortCode string, days int) (*models.ConversionMetrics, error) {
	args := m.Called(shortCode, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConversionMetrics), args.Error(1)
}

func TestRecordTouchpoint_Success(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	touchpoint := &models.AttributionTouchpoint{
		SessionID:       "session123",
		ShortCode:      "abc123",
		UserIP:         "192.168.1.1",
		UserAgent:      "Mozilla/5.0 Chrome/91.0",
		Referrer:       "https://google.com/search?q=test",
		CampaignSource: "google",
		CampaignMedium: "organic",
		CampaignName:   "",
		TouchpointTime: time.Now(),
	}

	mockDB.On("RecordTouchpoint", touchpoint).Return(nil)

	err := service.RecordTouchpoint(touchpoint)

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestRecordTouchpoint_DatabaseError(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	touchpoint := &models.AttributionTouchpoint{
		SessionID: "session123",
		ShortCode: "abc123",
	}

	mockDB.On("RecordTouchpoint", touchpoint).Return(assert.AnError)

	err := service.RecordTouchpoint(touchpoint)

	assert.Error(t, err)
	mockDB.AssertExpectations(t)
}

func TestCalculateAttribution_FirstTouchModel(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	journey := []*models.AttributionTouchpoint{
		{
			ID:             1,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "google",
			CampaignMedium: "organic",
			TouchpointTime: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:             2,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "facebook",
			CampaignMedium: "social",
			TouchpointTime: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:             3,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "newsletter",
			CampaignMedium: "email",
			TouchpointTime: time.Now().Add(-1 * time.Hour),
		},
	}

	mockDB.On("GetConversionJourney", "conv123").Return(journey, nil)

	result, err := service.CalculateAttribution("conv123", services.FirstTouchModel)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// First touch should get 100% attribution
	assert.Equal(t, float64(1.0), result[0].AttributionValue)
	assert.Equal(t, int64(1), result[0].TouchpointID)

	// Others should get 0%
	assert.Equal(t, float64(0.0), result[1].AttributionValue)
	assert.Equal(t, float64(0.0), result[2].AttributionValue)

	mockDB.AssertExpectations(t)
}

func TestCalculateAttribution_LastTouchModel(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	journey := []*models.AttributionTouchpoint{
		{
			ID:             1,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "google",
			CampaignMedium: "organic",
			TouchpointTime: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:             2,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "facebook",
			CampaignMedium: "social",
			TouchpointTime: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:             3,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "newsletter",
			CampaignMedium: "email",
			TouchpointTime: time.Now().Add(-1 * time.Hour),
		},
	}

	mockDB.On("GetConversionJourney", "conv123").Return(journey, nil)

	result, err := service.CalculateAttribution("conv123", services.LastTouchModel)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Last touch should get 100% attribution
	assert.Equal(t, float64(1.0), result[2].AttributionValue)
	assert.Equal(t, int64(3), result[2].TouchpointID)

	// Others should get 0%
	assert.Equal(t, float64(0.0), result[0].AttributionValue)
	assert.Equal(t, float64(0.0), result[1].AttributionValue)

	mockDB.AssertExpectations(t)
}

func TestCalculateAttribution_LinearModel(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	journey := []*models.AttributionTouchpoint{
		{
			ID:             1,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "google",
			CampaignMedium: "organic",
			TouchpointTime: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:             2,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "facebook",
			CampaignMedium: "social",
			TouchpointTime: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:             3,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "newsletter",
			CampaignMedium: "email",
			TouchpointTime: time.Now().Add(-1 * time.Hour),
		},
	}

	mockDB.On("GetConversionJourney", "conv123").Return(journey, nil)

	result, err := service.CalculateAttribution("conv123", services.LinearModel)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Each touchpoint should get equal attribution (1/3)
	expectedValue := 1.0 / 3.0
	for _, touchpoint := range result {
		assert.InDelta(t, expectedValue, touchpoint.AttributionValue, 0.001)
	}

	// Validate total attribution equals 1.0
	totalAttribution := result[0].AttributionValue + result[1].AttributionValue + result[2].AttributionValue
	assert.InDelta(t, 1.0, totalAttribution, 0.001)

	mockDB.AssertExpectations(t)
}

func TestCalculateAttribution_TimeDecayModel(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	now := time.Now()
	journey := []*models.AttributionTouchpoint{
		{
			ID:             1,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "google",
			CampaignMedium: "organic",
			TouchpointTime: now.Add(-7 * 24 * time.Hour), // 7 days ago
		},
		{
			ID:             2,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "facebook",
			CampaignMedium: "social",
			TouchpointTime: now.Add(-3 * 24 * time.Hour), // 3 days ago
		},
		{
			ID:             3,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "newsletter",
			CampaignMedium: "email",
			TouchpointTime: now.Add(-1 * time.Hour), // 1 hour ago
		},
	}

	mockDB.On("GetConversionJourney", "conv123").Return(journey, nil)

	result, err := service.CalculateAttribution("conv123", services.TimeDecayModel)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// More recent touchpoints should have higher attribution
	assert.Greater(t, result[2].AttributionValue, result[1].AttributionValue, "Most recent touchpoint should have highest attribution")
	assert.Greater(t, result[1].AttributionValue, result[0].AttributionValue, "Middle touchpoint should have more attribution than oldest")

	// Total attribution should equal 1.0
	totalAttribution := result[0].AttributionValue + result[1].AttributionValue + result[2].AttributionValue
	assert.InDelta(t, 1.0, totalAttribution, 0.001)

	mockDB.AssertExpectations(t)
}

func TestCalculateAttribution_PositionBasedModel(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	journey := []*models.AttributionTouchpoint{
		{
			ID:             1,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "google",
			CampaignMedium: "organic",
			TouchpointTime: time.Now().Add(-5 * time.Hour),
		},
		{
			ID:             2,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "facebook",
			CampaignMedium: "social",
			TouchpointTime: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:             3,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "twitter",
			CampaignMedium: "social",
			TouchpointTime: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:             4,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "newsletter",
			CampaignMedium: "email",
			TouchpointTime: time.Now().Add(-1 * time.Hour),
		},
	}

	mockDB.On("GetConversionJourney", "conv123").Return(journey, nil)

	result, err := service.CalculateAttribution("conv123", services.PositionBasedModel)

	assert.NoError(t, err)
	assert.Len(t, result, 4)

	// First and last touchpoints should get 40% each (0.4)
	assert.InDelta(t, 0.4, result[0].AttributionValue, 0.001, "First touchpoint should get 40%")
	assert.InDelta(t, 0.4, result[3].AttributionValue, 0.001, "Last touchpoint should get 40%")

	// Middle touchpoints should get 10% each (0.1)
	assert.InDelta(t, 0.1, result[1].AttributionValue, 0.001, "Middle touchpoint should get 10%")
	assert.InDelta(t, 0.1, result[2].AttributionValue, 0.001, "Middle touchpoint should get 10%")

	// Total attribution should equal 1.0
	totalAttribution := 0.0
	for _, tp := range result {
		totalAttribution += tp.AttributionValue
	}
	assert.InDelta(t, 1.0, totalAttribution, 0.001)

	mockDB.AssertExpectations(t)
}

func TestCalculateAttribution_SingleTouchpoint(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	journey := []*models.AttributionTouchpoint{
		{
			ID:             1,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "google",
			CampaignMedium: "organic",
			TouchpointTime: time.Now().Add(-1 * time.Hour),
		},
	}

	mockDB.On("GetConversionJourney", "conv123").Return(journey, nil)

	// Test all models with single touchpoint
	models := []services.AttributionModel{
		services.FirstTouchModel,
		services.LastTouchModel,
		services.LinearModel,
		services.TimeDecayModel,
		services.PositionBasedModel,
	}

	for _, model := range models {
		t.Run("Single touchpoint - "+string(model), func(t *testing.T) {
			result, err := service.CalculateAttribution("conv123", model)

			assert.NoError(t, err)
			assert.Len(t, result, 1)
			assert.Equal(t, float64(1.0), result[0].AttributionValue, "Single touchpoint should get 100% attribution")
		})
	}

	mockDB.AssertExpectations(t)
}

func TestCalculateAttribution_EmptyJourney(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	mockDB.On("GetConversionJourney", "empty").Return([]*models.AttributionTouchpoint{}, nil)

	result, err := service.CalculateAttribution("empty", services.LinearModel)

	assert.NoError(t, err)
	assert.Empty(t, result)

	mockDB.AssertExpectations(t)
}

func TestGetAttributionReport_Success(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	expectedReport := &models.AttributionReport{
		ShortCode:       "abc123",
		AttributionModel: string(services.LinearModel),
		TotalConversions: 150,
		TotalValue:      75000.00,
		ChannelBreakdown: []*models.ChannelAttribution{
			{
				Channel:        "google/organic",
				Conversions:    60,
				ConversionRate: 12.0,
				AttributedValue: 30000.00,
				Percentage:     40.0,
			},
			{
				Channel:        "facebook/social",
				Conversions:    45,
				ConversionRate: 9.0,
				AttributedValue: 22500.00,
				Percentage:     30.0,
			},
			{
				Channel:        "newsletter/email",
				Conversions:    45,
				ConversionRate: 9.0,
				AttributedValue: 22500.00,
				Percentage:     30.0,
			},
		},
		TimeRange: models.DateRange{
			StartDate: time.Now().Add(-30 * 24 * time.Hour),
			EndDate:   time.Now(),
		},
	}

	mockDB.On("GetAttributionReport", "abc123", services.LinearModel, 30).Return(expectedReport, nil)

	result, err := service.GetAttributionReport("abc123", services.LinearModel, 30)

	assert.NoError(t, err)
	assert.Equal(t, expectedReport, result)
	assert.Equal(t, int64(150), result.TotalConversions)
	assert.Equal(t, 75000.00, result.TotalValue)
	assert.Len(t, result.ChannelBreakdown, 3)

	// Validate percentages sum to 100%
	totalPercentage := 0.0
	for _, channel := range result.ChannelBreakdown {
		totalPercentage += channel.Percentage
	}
	assert.InDelta(t, 100.0, totalPercentage, 0.1)

	mockDB.AssertExpectations(t)
}

func TestGetAttributionReport_DatabaseError(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	mockDB.On("GetAttributionReport", "error", services.LinearModel, 30).Return(nil, assert.AnError)

	result, err := service.GetAttributionReport("error", services.LinearModel, 30)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockDB.AssertExpectations(t)
}

func TestRecordConversion_Success(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	conversion := &models.Conversion{
		ID:              "conv123",
		SessionID:       "session123",
		ShortCode:      "abc123",
		ConversionType: "purchase",
		Value:          99.99,
		ConversionTime: time.Now(),
	}

	mockDB.On("RecordConversion", conversion).Return(nil)

	err := service.RecordConversion(conversion)

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestGetSessionTouchpoints_Success(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	expectedTouchpoints := []*models.AttributionTouchpoint{
		{
			ID:             1,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "google",
			CampaignMedium: "organic",
			TouchpointTime: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:             2,
			SessionID:      "session123",
			ShortCode:     "def456",
			CampaignSource: "facebook",
			CampaignMedium: "social",
			TouchpointTime: time.Now().Add(-1 * time.Hour),
		},
	}

	mockDB.On("GetSessionTouchpoints", "session123").Return(expectedTouchpoints, nil)

	result, err := service.GetSessionTouchpoints("session123")

	assert.NoError(t, err)
	assert.Equal(t, expectedTouchpoints, result)
	assert.Len(t, result, 2)

	mockDB.AssertExpectations(t)
}

// Test attribution model validation
func TestCalculateAttribution_InvalidModel(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	journey := []*models.AttributionTouchpoint{
		{
			ID:        1,
			SessionID: "session123",
			ShortCode: "abc123",
		},
	}

	mockDB.On("GetConversionJourney", "conv123").Return(journey, nil)

	// Test with invalid model
	result, err := service.CalculateAttribution("conv123", services.AttributionModel("invalid_model"))

	// Should either handle gracefully or return an error
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, result)
	} else {
		// If no error, should default to a valid model behavior
		assert.NotNil(t, result)
	}
}

// Test touchpoint ordering
func TestCalculateAttribution_TouchpointOrdering(t *testing.T) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	now := time.Now()
	
	// Journey with touchpoints not in chronological order
	journey := []*models.AttributionTouchpoint{
		{
			ID:             2,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "facebook",
			TouchpointTime: now.Add(-1 * time.Hour), // Most recent
		},
		{
			ID:             1,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "google",
			TouchpointTime: now.Add(-3 * time.Hour), // Oldest
		},
		{
			ID:             3,
			SessionID:      "session123",
			ShortCode:     "abc123",
			CampaignSource: "email",
			TouchpointTime: now.Add(-2 * time.Hour), // Middle
		},
	}

	mockDB.On("GetConversionJourney", "conv123").Return(journey, nil)

	result, err := service.CalculateAttribution("conv123", services.FirstTouchModel)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Find the touchpoint with 100% attribution (should be the chronologically first one)
	var firstTouchpointFound bool
	for _, tp := range result {
		if tp.AttributionValue == 1.0 {
			assert.Equal(t, int64(1), tp.TouchpointID, "First touch attribution should go to chronologically first touchpoint")
			firstTouchpointFound = true
			break
		}
	}
	assert.True(t, firstTouchpointFound, "Should find the first touchpoint with 100% attribution")

	mockDB.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkCalculateAttribution_Linear_SmallJourney(b *testing.B) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	journey := []*models.AttributionTouchpoint{
		{ID: 1, SessionID: "session123", TouchpointTime: time.Now().Add(-3 * time.Hour)},
		{ID: 2, SessionID: "session123", TouchpointTime: time.Now().Add(-2 * time.Hour)},
		{ID: 3, SessionID: "session123", TouchpointTime: time.Now().Add(-1 * time.Hour)},
	}

	mockDB.On("GetConversionJourney", mock.AnythingOfType("string")).Return(journey, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CalculateAttribution("conv123", services.LinearModel)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkCalculateAttribution_Linear_LargeJourney(b *testing.B) {
	mockDB := &MockAttributionStorage{}
	mockConversion := &MockConversionTrackingService{}
	service := services.NewAttributionService(mockDB, mockConversion)

	// Create a large journey with 50 touchpoints
	journey := make([]*models.AttributionTouchpoint, 50)
	for i := 0; i < 50; i++ {
		journey[i] = &models.AttributionTouchpoint{
			ID:             int64(i + 1),
			SessionID:      "session123",
			TouchpointTime: time.Now().Add(time.Duration(-50+i) * time.Hour),
		}
	}

	mockDB.On("GetConversionJourney", mock.AnythingOfType("string")).Return(journey, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CalculateAttribution("conv123", services.LinearModel)
		if err != nil {
			b.Error(err)
		}
	}
}