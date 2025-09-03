package unit

import (
	"testing"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAdvancedAnalyticsStorage extends MockPostgresStorage for advanced analytics tests
type MockAdvancedAnalyticsStorage struct {
	MockPostgresStorage
}

func (m *MockAdvancedAnalyticsStorage) GetGeographicAnalytics(shortCode string, days int) (*models.GeographicAnalytics, error) {
	args := m.Called(shortCode, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GeographicAnalytics), args.Error(1)
}

func (m *MockAdvancedAnalyticsStorage) GetDeviceAnalytics(shortCode string, days int) (*models.DeviceAnalytics, error) {
	args := m.Called(shortCode, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DeviceAnalytics), args.Error(1)
}

func (m *MockAdvancedAnalyticsStorage) GetReferrerAnalytics(shortCode string, days int) ([]*models.ReferrerStat, error) {
	args := m.Called(shortCode, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ReferrerStat), args.Error(1)
}

func (m *MockAdvancedAnalyticsStorage) GetTimeBasedAnalytics(shortCode string, days int) ([]*models.TimeBasedStat, error) {
	args := m.Called(shortCode, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TimeBasedStat), args.Error(1)
}

func (m *MockAdvancedAnalyticsStorage) GetConversionFunnelData(shortCode string, days int) ([]*models.FunnelStep, error) {
	args := m.Called(shortCode, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FunnelStep), args.Error(1)
}

func TestGetGeographicAnalytics_Success(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := &models.GeographicAnalytics{
		Countries: []*models.CountryStat{
			{Country: "United States", Code: "US", Clicks: 150, Percentage: 45.5},
			{Country: "United Kingdom", Code: "UK", Clicks: 100, Percentage: 30.3},
			{Country: "Canada", Code: "CA", Clicks: 80, Percentage: 24.2},
		},
		Cities: []*models.CityStat{
			{City: "New York", Country: "United States", Clicks: 75, Percentage: 22.7},
			{City: "London", Country: "United Kingdom", Clicks: 60, Percentage: 18.2},
			{City: "Toronto", Country: "Canada", Clicks: 40, Percentage: 12.1},
		},
		TotalClicks: 330,
	}

	mockDB.On("GetGeographicAnalytics", "abc123", 30).Return(expectedData, nil)

	result, err := service.GetGeographicAnalytics("abc123", 30)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	assert.Len(t, result.Countries, 3)
	assert.Len(t, result.Cities, 3)
	assert.Equal(t, int64(330), result.TotalClicks)

	mockDB.AssertExpectations(t)
}

func TestGetGeographicAnalytics_EmptyData(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := &models.GeographicAnalytics{
		Countries:   []*models.CountryStat{},
		Cities:      []*models.CityStat{},
		TotalClicks: 0,
	}

	mockDB.On("GetGeographicAnalytics", "empty", 7).Return(expectedData, nil)

	result, err := service.GetGeographicAnalytics("empty", 7)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Countries)
	assert.Empty(t, result.Cities)
	assert.Equal(t, int64(0), result.TotalClicks)

	mockDB.AssertExpectations(t)
}

func TestGetDeviceAnalytics_Success(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := &models.DeviceAnalytics{
		Devices: []*models.DeviceStat{
			{DeviceType: "Desktop", Clicks: 200, Percentage: 50.0},
			{DeviceType: "Mobile", Clicks: 150, Percentage: 37.5},
			{DeviceType: "Tablet", Clicks: 50, Percentage: 12.5},
		},
		Browsers: []*models.BrowserStat{
			{Browser: "Chrome", Version: "91.0", Clicks: 180, Percentage: 45.0},
			{Browser: "Firefox", Version: "89.0", Clicks: 120, Percentage: 30.0},
			{Browser: "Safari", Version: "14.1", Clicks: 100, Percentage: 25.0},
		},
		OperatingSystems: []*models.OStat{
			{OS: "Windows", Version: "10", Clicks: 160, Percentage: 40.0},
			{OS: "macOS", Version: "11.0", Clicks: 120, Percentage: 30.0},
			{OS: "Android", Version: "11", Clicks: 80, Percentage: 20.0},
			{OS: "iOS", Version: "14.6", Clicks: 40, Percentage: 10.0},
		},
		TotalClicks: 400,
	}

	mockDB.On("GetDeviceAnalytics", "abc123", 30).Return(expectedData, nil)

	result, err := service.GetDeviceAnalytics("abc123", 30)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	assert.Len(t, result.Devices, 3)
	assert.Len(t, result.Browsers, 3)
	assert.Len(t, result.OperatingSystems, 4)
	assert.Equal(t, int64(400), result.TotalClicks)

	// Validate percentages sum correctly
	var deviceTotal float64
	for _, device := range result.Devices {
		deviceTotal += device.Percentage
	}
	assert.InDelta(t, 100.0, deviceTotal, 0.1)

	mockDB.AssertExpectations(t)
}

func TestGetReferrerAnalytics_Success(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := []*models.ReferrerStat{
		{
			Source:     "google.com",
			Medium:     "organic",
			Campaign:   "",
			Clicks:     150,
			Percentage: 37.5,
		},
		{
			Source:     "facebook.com",
			Medium:     "social",
			Campaign:   "summer_promo",
			Clicks:     100,
			Percentage: 25.0,
		},
		{
			Source:     "direct",
			Medium:     "none",
			Campaign:   "",
			Clicks:     80,
			Percentage: 20.0,
		},
		{
			Source:     "newsletter",
			Medium:     "email",
			Campaign:   "weekly_digest",
			Clicks:     70,
			Percentage: 17.5,
		},
	}

	mockDB.On("GetReferrerAnalytics", "abc123", 30).Return(expectedData, nil)

	result, err := service.GetReferrerAnalytics("abc123", 30)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	assert.Len(t, result, 4)

	// Test sorting (should be by clicks descending)
	for i := 1; i < len(result); i++ {
		assert.GreaterOrEqual(t, result[i-1].Clicks, result[i].Clicks, "Results should be sorted by clicks descending")
	}

	// Validate total percentages
	var totalPercentage float64
	for _, stat := range result {
		totalPercentage += stat.Percentage
	}
	assert.InDelta(t, 100.0, totalPercentage, 0.1)

	mockDB.AssertExpectations(t)
}

func TestGetTimeBasedAnalytics_HourlyData(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := []*models.TimeBasedStat{
		{Timestamp: time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC), Clicks: 25, Period: "hour"},
		{Timestamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), Clicks: 45, Period: "hour"},
		{Timestamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC), Clicks: 60, Period: "hour"},
		{Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Clicks: 55, Period: "hour"},
		{Timestamp: time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC), Clicks: 40, Period: "hour"},
	}

	mockDB.On("GetTimeBasedAnalytics", "abc123", 1).Return(expectedData, nil)

	result, err := service.GetTimeBasedAnalytics("abc123", 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	assert.Len(t, result, 5)

	// Validate time ordering
	for i := 1; i < len(result); i++ {
		assert.True(t, result[i-1].Timestamp.Before(result[i].Timestamp), "Results should be ordered chronologically")
	}

	// All periods should be "hour" for daily data
	for _, stat := range result {
		assert.Equal(t, "hour", stat.Period)
	}

	mockDB.AssertExpectations(t)
}

func TestGetTimeBasedAnalytics_DailyData(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := []*models.TimeBasedStat{
		{Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Clicks: 150, Period: "day"},
		{Timestamp: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), Clicks: 200, Period: "day"},
		{Timestamp: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), Clicks: 180, Period: "day"},
		{Timestamp: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), Clicks: 220, Period: "day"},
		{Timestamp: time.Date(2023, 1, 5, 0, 0, 0, 0, time.UTC), Clicks: 175, Period: "day"},
	}

	mockDB.On("GetTimeBasedAnalytics", "abc123", 30).Return(expectedData, nil)

	result, err := service.GetTimeBasedAnalytics("abc123", 30)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	assert.Len(t, result, 5)

	// All periods should be "day" for monthly data
	for _, stat := range result {
		assert.Equal(t, "day", stat.Period)
	}

	mockDB.AssertExpectations(t)
}

func TestGetConversionFunnelData_Success(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := []*models.FunnelStep{
		{
			Step:         1,
			StepName:     "URL Click",
			Users:        1000,
			Conversions:  1000,
			ConversionRate: 100.0,
		},
		{
			Step:         2,
			StepName:     "Page View",
			Users:        800,
			Conversions:  800,
			ConversionRate: 80.0,
		},
		{
			Step:         3,
			StepName:     "Engagement",
			Users:        600,
			Conversions:  600,
			ConversionRate: 60.0,
		},
		{
			Step:         4,
			StepName:     "Conversion",
			Users:        200,
			Conversions:  200,
			ConversionRate: 20.0,
		},
	}

	mockDB.On("GetConversionFunnelData", "abc123", 30).Return(expectedData, nil)

	result, err := service.GetConversionFunnelData("abc123", 30)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, result)
	assert.Len(t, result, 4)

	// Validate funnel logic - each step should have equal or fewer users
	for i := 1; i < len(result); i++ {
		assert.GreaterOrEqual(t, result[i-1].Users, result[i].Users, "Funnel steps should show decreasing user counts")
		assert.GreaterOrEqual(t, result[i-1].Conversions, result[i].Conversions, "Funnel steps should show decreasing conversion counts")
	}

	// Validate conversion rates are reasonable
	for _, step := range result {
		assert.GreaterOrEqual(t, step.ConversionRate, 0.0, "Conversion rate should not be negative")
		assert.LessOrEqual(t, step.ConversionRate, 100.0, "Conversion rate should not exceed 100%")
	}

	mockDB.AssertExpectations(t)
}

func TestGetConversionFunnelData_EmptyFunnel(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := []*models.FunnelStep{}

	mockDB.On("GetConversionFunnelData", "empty", 7).Return(expectedData, nil)

	result, err := service.GetConversionFunnelData("empty", 7)

	assert.NoError(t, err)
	assert.Empty(t, result)

	mockDB.AssertExpectations(t)
}

// Test error handling scenarios
func TestGetGeographicAnalytics_DatabaseError(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	mockDB.On("GetGeographicAnalytics", "error", 30).Return(nil, assert.AnError)

	result, err := service.GetGeographicAnalytics("error", 30)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockDB.AssertExpectations(t)
}

func TestGetDeviceAnalytics_DatabaseError(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	mockDB.On("GetDeviceAnalytics", "error", 30).Return(nil, assert.AnError)

	result, err := service.GetDeviceAnalytics("error", 30)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockDB.AssertExpectations(t)
}

func TestGetReferrerAnalytics_DatabaseError(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	mockDB.On("GetReferrerAnalytics", "error", 30).Return(nil, assert.AnError)

	result, err := service.GetReferrerAnalytics("error", 30)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockDB.AssertExpectations(t)
}

func TestGetTimeBasedAnalytics_DatabaseError(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	mockDB.On("GetTimeBasedAnalytics", "error", 30).Return(nil, assert.AnError)

	result, err := service.GetTimeBasedAnalytics("error", 30)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockDB.AssertExpectations(t)
}

// Test parameter validation
func TestAdvancedAnalytics_InvalidShortCode(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	invalidCodes := []string{"", "  ", "a", "very-long-invalid-short-code-that-exceeds-limits"}

	for _, code := range invalidCodes {
		t.Run("Invalid code: "+code, func(t *testing.T) {
			// We expect the service to handle these gracefully
			// In practice, the service might validate input or pass through to database
			// For now, we'll test that it doesn't panic

			mockDB.On("GetGeographicAnalytics", code, 30).Return(&models.GeographicAnalytics{
				Countries: []*models.CountryStat{},
				Cities:    []*models.CityStat{},
				TotalClicks: 0,
			}, nil).Maybe()

			result, err := service.GetGeographicAnalytics(code, 30)

			// Should handle gracefully (either return empty data or appropriate error)
			if err == nil {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAdvancedAnalytics_InvalidDays(t *testing.T) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	invalidDays := []int{-1, 0, 366, 1000}

	for _, days := range invalidDays {
		t.Run("Invalid days", func(t *testing.T) {
			// Service should handle invalid day ranges appropriately
			mockDB.On("GetGeographicAnalytics", "test", days).Return(&models.GeographicAnalytics{
				Countries: []*models.CountryStat{},
				Cities:    []*models.CityStat{},
				TotalClicks: 0,
			}, nil).Maybe()

			result, err := service.GetGeographicAnalytics("test", days)

			// Should handle gracefully
			if err == nil {
				assert.NotNil(t, result)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkGetGeographicAnalytics(b *testing.B) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := &models.GeographicAnalytics{
		Countries: []*models.CountryStat{
			{Country: "United States", Code: "US", Clicks: 150, Percentage: 50.0},
			{Country: "United Kingdom", Code: "UK", Clicks: 100, Percentage: 33.3},
			{Country: "Canada", Code: "CA", Clicks: 50, Percentage: 16.7},
		},
		Cities: []*models.CityStat{
			{City: "New York", Country: "United States", Clicks: 75, Percentage: 25.0},
			{City: "London", Country: "United Kingdom", Clicks: 60, Percentage: 20.0},
			{City: "Toronto", Country: "Canada", Clicks: 25, Percentage: 8.3},
		},
		TotalClicks: 300,
	}

	mockDB.On("GetGeographicAnalytics", mock.AnythingOfType("string"), mock.AnythingOfType("int")).Return(expectedData, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetGeographicAnalytics("bench", 30)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkGetDeviceAnalytics(b *testing.B) {
	mockDB := &MockAdvancedAnalyticsStorage{}
	service := services.NewAdvancedAnalyticsService(mockDB)

	expectedData := &models.DeviceAnalytics{
		Devices: []*models.DeviceStat{
			{DeviceType: "Desktop", Clicks: 200, Percentage: 50.0},
			{DeviceType: "Mobile", Clicks: 150, Percentage: 37.5},
			{DeviceType: "Tablet", Clicks: 50, Percentage: 12.5},
		},
		TotalClicks: 400,
	}

	mockDB.On("GetDeviceAnalytics", mock.AnythingOfType("string"), mock.AnythingOfType("int")).Return(expectedData, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetDeviceAnalytics("bench", 30)
		if err != nil {
			b.Error(err)
		}
	}
}