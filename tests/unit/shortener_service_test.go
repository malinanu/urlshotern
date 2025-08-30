package unit

import (
	"testing"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPostgresStorage is a mock implementation of PostgresStorage
type MockPostgresStorage struct {
	mock.Mock
}

func (m *MockPostgresStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPostgresStorage) SaveURLMapping(mapping *models.URLMapping) error {
	args := m.Called(mapping)
	return args.Error(0)
}

func (m *MockPostgresStorage) GetURLMappingByShortCode(shortCode string) (*models.URLMapping, error) {
	args := m.Called(shortCode)
	return args.Get(0).(*models.URLMapping), args.Error(1)
}

func (m *MockPostgresStorage) ShortCodeExists(shortCode string) (bool, error) {
	args := m.Called(shortCode)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostgresStorage) IncrementClickCount(shortCode string) error {
	args := m.Called(shortCode)
	return args.Error(0)
}

func (m *MockPostgresStorage) SaveClickEvent(event *models.ClickEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockPostgresStorage) GetAnalytics(shortCode string, days int) (*models.AnalyticsResponse, error) {
	args := m.Called(shortCode, days)
	return args.Get(0).(*models.AnalyticsResponse), args.Error(1)
}

// MockRedisStorage is a mock implementation of RedisStorage
type MockRedisStorage struct {
	mock.Mock
}

func (m *MockRedisStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisStorage) SetURLMapping(shortCode string, mapping *models.URLMapping, ttl time.Duration) error {
	args := m.Called(shortCode, mapping, ttl)
	return args.Error(0)
}

func (m *MockRedisStorage) GetURLMapping(shortCode string) (*models.URLMapping, error) {
	args := m.Called(shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.URLMapping), args.Error(1)
}

func (m *MockRedisStorage) DeleteURLMapping(shortCode string) error {
	args := m.Called(shortCode)
	return args.Error(0)
}

func (m *MockRedisStorage) IncrementClickCount(shortCode string) (int64, error) {
	args := m.Called(shortCode)
	return args.Get(0).(int64), args.Error(1)
}

func TestShortenURL_ValidURL(t *testing.T) {
	// Initialize Snowflake for testing
	utils.InitializeSnowflake(1)

	mockDB := new(MockPostgresStorage)
	mockRedis := new(MockRedisStorage)
	
	config := &configs.Config{
		BaseURL:    "http://localhost:8080",
		ServerHost: "localhost",
	}

	service := services.NewShortenerService(mockDB, mockRedis, config)

	// Setup mocks
	mockDB.On("ShortCodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockDB.On("SaveURLMapping", mock.AnythingOfType("*models.URLMapping")).Return(nil)
	mockRedis.On("SetURLMapping", mock.AnythingOfType("string"), mock.AnythingOfType("*models.URLMapping"), mock.AnythingOfType("time.Duration")).Return(nil)

	request := &models.ShortenRequest{
		URL: "https://www.example.com",
	}

	response, err := service.ShortenURL(request, "127.0.0.1")

	assert.NoError(t, err)
	assert.NotEmpty(t, response.ShortCode)
	assert.Equal(t, "https://www.example.com", response.OriginalURL)
	assert.Contains(t, response.ShortURL, config.BaseURL)

	mockDB.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestShortenURL_InvalidURL(t *testing.T) {
	mockDB := new(MockPostgresStorage)
	mockRedis := new(MockRedisStorage)
	config := &configs.Config{BaseURL: "http://localhost:8080"}

	service := services.NewShortenerService(mockDB, mockRedis, config)

	tests := []struct {
		name string
		url  string
	}{
		{"Empty URL", ""},
		{"Invalid scheme", "ftp://example.com"},
		{"No scheme", "example.com"},
		{"Invalid format", "not-a-url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &models.ShortenRequest{URL: tt.url}
			_, err := service.ShortenURL(request, "127.0.0.1")
			assert.Error(t, err)
		})
	}
}

func TestShortenURL_CustomCode(t *testing.T) {
	utils.InitializeSnowflake(1)

	mockDB := new(MockPostgresStorage)
	mockRedis := new(MockRedisStorage)
	config := &configs.Config{BaseURL: "http://localhost:8080"}

	service := services.NewShortenerService(mockDB, mockRedis, config)

	// Setup mocks
	mockDB.On("ShortCodeExists", "custom123").Return(false, nil)
	mockDB.On("SaveURLMapping", mock.AnythingOfType("*models.URLMapping")).Return(nil)
	mockRedis.On("SetURLMapping", "custom123", mock.AnythingOfType("*models.URLMapping"), mock.AnythingOfType("time.Duration")).Return(nil)

	request := &models.ShortenRequest{
		URL:        "https://www.example.com",
		CustomCode: "custom123",
	}

	response, err := service.ShortenURL(request, "127.0.0.1")

	assert.NoError(t, err)
	assert.Equal(t, "custom123", response.ShortCode)

	mockDB.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestShortenURL_CustomCodeAlreadyExists(t *testing.T) {
	mockDB := new(MockPostgresStorage)
	mockRedis := new(MockRedisStorage)
	config := &configs.Config{BaseURL: "http://localhost:8080"}

	service := services.NewShortenerService(mockDB, mockRedis, config)

	// Setup mock to return that code already exists
	mockDB.On("ShortCodeExists", "existing").Return(true, nil)

	request := &models.ShortenRequest{
		URL:        "https://www.example.com",
		CustomCode: "existing",
	}

	_, err := service.ShortenURL(request, "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, services.ErrCustomCodeAlreadyExists, err)

	mockDB.AssertExpectations(t)
}

func TestGetOriginalURL_CacheHit(t *testing.T) {
	mockDB := new(MockPostgresStorage)
	mockRedis := new(MockRedisStorage)
	config := &configs.Config{}

	service := services.NewShortenerService(mockDB, mockRedis, config)

	expectedMapping := &models.URLMapping{
		ShortCode:   "abc123",
		OriginalURL: "https://www.example.com",
		IsActive:    true,
	}

	// Setup mock to return from cache
	mockRedis.On("GetURLMapping", "abc123").Return(expectedMapping, nil)

	result, err := service.GetOriginalURL("abc123")

	assert.NoError(t, err)
	assert.Equal(t, expectedMapping, result)

	// Should not call database if cache hit
	mockDB.AssertNotCalled(t, "GetURLMappingByShortCode")
	mockRedis.AssertExpectations(t)
}

func TestValidateCustomCode(t *testing.T) {
	mockDB := new(MockPostgresStorage)
	mockRedis := new(MockRedisStorage)
	config := &configs.Config{}

	service := services.NewShortenerService(mockDB, mockRedis, config)

	tests := []struct {
		name      string
		code      string
		shouldErr bool
	}{
		{"Valid code", "abc123", false},
		{"Too short", "ab", true},
		{"Too long", "abcdefghijk", true},
		{"Invalid characters", "abc-123", true},
		{"Reserved word", "api", true},
		{"Reserved word case insensitive", "API", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is testing internal validation logic
			// In a real implementation, you'd expose this method or test through public methods
			
			// For now, we'll test through ShortenURL which calls validateCustomCode internally
			mockDB.On("ShortCodeExists", mock.AnythingOfType("string")).Return(false, nil).Maybe()
			mockDB.On("SaveURLMapping", mock.AnythingOfType("*models.URLMapping")).Return(nil).Maybe()
			mockRedis.On("SetURLMapping", mock.AnythingOfType("string"), mock.AnythingOfType("*models.URLMapping"), mock.AnythingOfType("time.Duration")).Return(nil).Maybe()

			request := &models.ShortenRequest{
				URL:        "https://www.example.com",
				CustomCode: tt.code,
			}

			_, err := service.ShortenURL(request, "127.0.0.1")

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}