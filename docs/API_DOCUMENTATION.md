# URL Shortener API Documentation

## Overview

This is a comprehensive URL shortener service with advanced analytics, A/B testing, attribution modeling, and real-time insights. The API provides enterprise-level features using only free services and open-source libraries.

**Base URL:** `http://localhost:8080` (development)

**Version:** v1

**Content-Type:** `application/json`

## Authentication

Currently, the API uses simple authentication mechanisms. Enhanced authentication features are available in Phase 4.

## Rate Limiting

- **Standard endpoints:** 1000 requests per hour per IP
- **Analytics endpoints:** 500 requests per hour per IP  
- **Real-time endpoints:** 100 requests per minute per IP

## Core URL Shortening

### Shorten URL

Create a shortened URL from a long URL.

**POST** `/api/v1/shorten`

**Request Body:**
```json
{
  "url": "https://www.example.com/very/long/path",
  "custom_code": "my-custom-code", // optional
  "expires_at": "2024-12-31T23:59:59Z", // optional
  "campaign_source": "newsletter", // optional
  "campaign_medium": "email", // optional
  "campaign_name": "weekly_digest" // optional
}
```

**Response:**
```json
{
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "original_url": "https://www.example.com/very/long/path",
  "expires_at": "2024-12-31T23:59:59Z",
  "created_at": "2023-06-15T10:30:00Z"
}
```

**Status Codes:**
- `201 Created` - URL successfully shortened
- `400 Bad Request` - Invalid URL or parameters
- `409 Conflict` - Custom code already exists
- `429 Too Many Requests` - Rate limit exceeded

### Redirect to Original URL

Redirect to the original URL using the short code.

**GET** `/{shortCode}`

**Response:**
- `302 Found` - Redirects to original URL
- `404 Not Found` - Short code does not exist
- `410 Gone` - Short URL has expired

## Basic Analytics

### Get Basic Analytics

Get basic analytics data for a short URL.

**GET** `/api/v1/analytics/{shortCode}`

**Query Parameters:**
- `days` (optional) - Number of days to include (1-365, default: 30)

**Response:**
```json
{
  "short_code": "abc123",
  "original_url": "https://www.example.com",
  "total_clicks": 1250,
  "unique_clicks": 890,
  "click_rate": 71.2,
  "created_at": "2023-06-15T10:30:00Z",
  "daily_stats": [
    {
      "date": "2023-06-15",
      "clicks": 45,
      "unique_clicks": 32
    }
  ]
}
```

## Advanced Analytics

### Geographic Analytics

Get geographic distribution of clicks.

**GET** `/api/v1/analytics/{shortCode}/geographic`

**Query Parameters:**
- `days` (optional) - Number of days (1-365, default: 30)

**Response:**
```json
{
  "countries": [
    {
      "country": "United States",
      "code": "US",
      "clicks": 450,
      "percentage": 36.0
    },
    {
      "country": "United Kingdom", 
      "code": "UK",
      "clicks": 320,
      "percentage": 25.6
    }
  ],
  "cities": [
    {
      "city": "New York",
      "country": "United States",
      "clicks": 180,
      "percentage": 14.4
    },
    {
      "city": "London",
      "country": "United Kingdom", 
      "clicks": 150,
      "percentage": 12.0
    }
  ],
  "total_clicks": 1250
}
```

### Device Analytics

Get device, browser, and operating system analytics.

**GET** `/api/v1/analytics/{shortCode}/devices`

**Query Parameters:**
- `days` (optional) - Number of days (1-365, default: 30)

**Response:**
```json
{
  "devices": [
    {
      "device_type": "Desktop",
      "clicks": 650,
      "percentage": 52.0
    },
    {
      "device_type": "Mobile",
      "clicks": 480,
      "percentage": 38.4
    },
    {
      "device_type": "Tablet",
      "clicks": 120,
      "percentage": 9.6
    }
  ],
  "browsers": [
    {
      "browser": "Chrome",
      "version": "91.0",
      "clicks": 560,
      "percentage": 44.8
    },
    {
      "browser": "Firefox",
      "version": "89.0",
      "clicks": 280,
      "percentage": 22.4
    }
  ],
  "operating_systems": [
    {
      "os": "Windows",
      "version": "10",
      "clicks": 420,
      "percentage": 33.6
    },
    {
      "os": "macOS",
      "version": "11.0",
      "clicks": 320,
      "percentage": 25.6
    }
  ],
  "total_clicks": 1250
}
```

### Referrer Analytics

Get referrer and campaign source analytics.

**GET** `/api/v1/analytics/{shortCode}/referrers`

**Query Parameters:**
- `days` (optional) - Number of days (1-365, default: 30)

**Response:**
```json
[
  {
    "source": "google.com",
    "medium": "organic",
    "campaign": "",
    "clicks": 420,
    "percentage": 33.6
  },
  {
    "source": "facebook.com",
    "medium": "social",
    "campaign": "summer_promo",
    "clicks": 280,
    "percentage": 22.4
  },
  {
    "source": "direct",
    "medium": "none",
    "campaign": "",
    "clicks": 350,
    "percentage": 28.0
  }
]
```

### Time-based Analytics

Get hourly or daily click patterns.

**GET** `/api/v1/analytics/{shortCode}/timeline`

**Query Parameters:**
- `days` (optional) - Number of days (1-365, default: 30)

**Response:**
```json
[
  {
    "timestamp": "2023-06-15T09:00:00Z",
    "clicks": 25,
    "period": "hour"
  },
  {
    "timestamp": "2023-06-15T10:00:00Z", 
    "clicks": 42,
    "period": "hour"
  },
  {
    "timestamp": "2023-06-15T11:00:00Z",
    "clicks": 38,
    "period": "hour"
  }
]
```

### Conversion Funnel

Get conversion funnel data.

**GET** `/api/v1/analytics/{shortCode}/funnel`

**Query Parameters:**
- `days` (optional) - Number of days (1-365, default: 30)

**Response:**
```json
[
  {
    "step": 1,
    "step_name": "URL Click",
    "users": 1000,
    "conversions": 1000,
    "conversion_rate": 100.0
  },
  {
    "step": 2,
    "step_name": "Page View",
    "users": 800,
    "conversions": 800,
    "conversion_rate": 80.0
  },
  {
    "step": 3,
    "step_name": "Engagement",
    "users": 600,
    "conversions": 600,
    "conversion_rate": 60.0
  },
  {
    "step": 4,
    "step_name": "Conversion",
    "users": 200,
    "conversions": 200,
    "conversion_rate": 20.0
  }
]
```

## Attribution Modeling

### Get Attribution Report

Get multi-touch attribution analysis.

**GET** `/api/v1/attribution/{shortCode}`

**Query Parameters:**
- `model` (optional) - Attribution model: `first_touch`, `last_touch`, `linear`, `time_decay`, `position_based`, `data_driven` (default: `last_touch`)
- `days` (optional) - Number of days (1-365, default: 30)

**Response:**
```json
{
  "short_code": "abc123",
  "attribution_model": "linear",
  "total_conversions": 150,
  "total_value": 75000.00,
  "channel_breakdown": [
    {
      "channel": "google/organic",
      "conversions": 60,
      "conversion_rate": 12.0,
      "attributed_value": 30000.00,
      "percentage": 40.0
    },
    {
      "channel": "facebook/social",
      "conversions": 45,
      "conversion_rate": 9.0,
      "attributed_value": 22500.00,
      "percentage": 30.0
    }
  ],
  "time_range": {
    "start_date": "2023-05-15T00:00:00Z",
    "end_date": "2023-06-15T23:59:59Z"
  }
}
```

### Record Conversion

Record a conversion event for attribution analysis.

**POST** `/api/v1/attribution/conversion`

**Request Body:**
```json
{
  "session_id": "session_123456",
  "short_code": "abc123",
  "conversion_type": "purchase",
  "value": 99.99,
  "event_time": "2023-06-15T15:30:00Z"
}
```

**Response:**
```json
{
  "conversion_id": "conv_789012",
  "status": "recorded",
  "attribution_calculated": true
}
```

## A/B Testing

### Create A/B Test

Create a new A/B test.

**POST** `/api/v1/ab-test`

**Request Body:**
```json
{
  "name": "Button Color Test",
  "description": "Testing red vs blue call-to-action buttons",
  "short_codes": ["test123", "test456"],
  "traffic_split": [0.5, 0.5],
  "metric_name": "click_through_rate",
  "start_date": "2023-06-15T00:00:00Z",
  "end_date": "2023-06-30T23:59:59Z",
  "significance_threshold": 95.0
}
```

**Response:**
```json
{
  "test_id": 12345,
  "status": "active",
  "created_at": "2023-06-15T10:30:00Z",
  "estimated_duration_days": 14
}
```

### Get A/B Test Results

Get current A/B test results.

**GET** `/api/v1/ab-test/{testID}`

**Response:**
```json
{
  "test_id": 12345,
  "name": "Button Color Test",
  "status": "running",
  "variant_results": [
    {
      "variant_id": 1,
      "variant_name": "Control",
      "short_code": "test123",
      "sessions": 1500,
      "conversions": 75,
      "conversion_rate": 5.0,
      "is_control": true
    },
    {
      "variant_id": 2,
      "variant_name": "Variant A",
      "short_code": "test456",
      "sessions": 1480,
      "conversions": 96,
      "conversion_rate": 6.49,
      "is_control": false
    }
  ],
  "start_date": "2023-06-15T00:00:00Z",
  "end_date": "2023-06-30T23:59:59Z"
}
```

### Get Statistical Significance Analysis

Get detailed statistical analysis of A/B test results.

**GET** `/api/v1/ab-test/{testID}/significance`

**Query Parameters:**
- `confidence` (optional) - Confidence level (90, 95, 99, default: 95)

**Response:**
```json
{
  "is_significant": true,
  "p_value": 0.0234,
  "z_score": 2.28,
  "control_conversion_rate": 5.0,
  "variant_conversion_rate": 6.49,
  "improvement_percentage": 29.8,
  "effect_size": 0.12,
  "confidence_interval": [-0.5, 2.98],
  "minimum_detectable_effect": 1.5,
  "sample_size_recommendation": 2400,
  "confidence_level": 95.0
}
```

### Get Sequential Test Results

Get sequential testing analysis for early stopping decisions.

**GET** `/api/v1/ab-test/{testID}/sequential`

**Response:**
```json
{
  "decision": "continue",
  "log_likelihood_ratio": 1.85,
  "upper_boundary": 2.94,
  "lower_boundary": -2.94,
  "probability_variant_wins": 0.87,
  "recommendation": "Continue test - approaching significance but not yet conclusive",
  "current_power": 0.72,
  "days_running": 8,
  "estimated_days_remaining": 6
}
```

### Stop A/B Test

Stop a running A/B test.

**POST** `/api/v1/ab-test/{testID}/stop`

**Request Body:**
```json
{
  "reason": "Reached statistical significance",
  "winner_variant_id": 2
}
```

**Response:**
```json
{
  "test_id": 12345,
  "status": "stopped",
  "stopped_at": "2023-06-23T14:30:00Z",
  "final_results": {
    "winner": "Variant A",
    "improvement": 29.8,
    "confidence": 95.7
  }
}
```

## Real-time Analytics

### Get Real-time Analytics

Get current real-time statistics.

**GET** `/api/v1/realtime/{shortCode}`

**Response:**
```json
{
  "short_code": "abc123",
  "active_users": 24,
  "clicks_last_hour": 45,
  "clicks_today": 234,
  "top_countries_today": [
    {
      "country": "United States",
      "clicks": 89
    },
    {
      "country": "United Kingdom",
      "clicks": 67
    }
  ],
  "recent_activity": [
    {
      "timestamp": "2023-06-15T15:28:30Z",
      "country": "US",
      "device": "Mobile",
      "referrer": "google.com"
    }
  ],
  "timestamp": "2023-06-15T15:30:00Z"
}
```

### WebSocket Real-time Updates

Connect to real-time analytics via WebSocket.

**WebSocket** `/api/v1/realtime/{shortCode}/websocket`

**Connection Headers:**
```
Connection: Upgrade
Upgrade: websocket
Sec-WebSocket-Version: 13
Sec-WebSocket-Key: <key>
```

**Messages Received:**
```json
{
  "type": "click_event",
  "data": {
    "short_code": "abc123",
    "timestamp": "2023-06-15T15:30:45Z",
    "country": "US",
    "city": "New York",
    "device": "Desktop",
    "browser": "Chrome",
    "referrer": "facebook.com"
  }
}
```

```json
{
  "type": "stats_update",
  "data": {
    "short_code": "abc123",
    "active_users": 26,
    "clicks_last_hour": 47,
    "clicks_today": 236
  }
}
```

## Health Check

### System Health

Check system health and status.

**GET** `/health`

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2023-06-15T15:30:00Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "external_apis": "healthy"
  },
  "version": "1.0.0"
}
```

## Error Responses

All endpoints may return these standard error responses:

### 400 Bad Request
```json
{
  "error": "invalid_request",
  "message": "The request is invalid or malformed",
  "details": {
    "field": "url",
    "issue": "Invalid URL format"
  }
}
```

### 401 Unauthorized
```json
{
  "error": "unauthorized",
  "message": "Authentication required"
}
```

### 403 Forbidden
```json
{
  "error": "forbidden",
  "message": "Access denied"
}
```

### 404 Not Found
```json
{
  "error": "not_found",
  "message": "Short code not found"
}
```

### 409 Conflict
```json
{
  "error": "custom_code_already_exists",
  "message": "The custom code is already in use"
}
```

### 429 Too Many Requests
```json
{
  "error": "rate_limit_exceeded",
  "message": "Too many requests",
  "retry_after": 3600
}
```

### 500 Internal Server Error
```json
{
  "error": "internal_server_error",
  "message": "An unexpected error occurred"
}
```

## Common Error Codes

- `invalid_request` - Malformed request body or parameters
- `invalid_url` - URL format is invalid
- `invalid_short_code` - Short code format is invalid
- `invalid_days_parameter` - Days parameter out of range (1-365)
- `invalid_confidence_parameter` - Confidence level not supported (90, 95, 99)
- `invalid_attribution_model` - Attribution model not supported
- `custom_code_already_exists` - Custom short code is already taken
- `short_code_not_found` - Short code does not exist
- `test_not_found` - A/B test ID does not exist
- `unauthorized` - Authentication required
- `forbidden` - Access denied
- `rate_limit_exceeded` - Too many requests
- `internal_server_error` - Unexpected server error

## SDKs and Libraries

### JavaScript/Node.js
```bash
npm install url-shortener-sdk
```

```javascript
const URLShortener = require('url-shortener-sdk');

const client = new URLShortener({
  baseURL: 'http://localhost:8080',
  apiKey: 'your-api-key' // when authentication is enabled
});

// Shorten URL
const result = await client.shorten({
  url: 'https://www.example.com',
  customCode: 'my-link'
});

// Get analytics
const analytics = await client.getAnalytics('abc123', { days: 30 });

// Get geographic analytics
const geographic = await client.getGeographicAnalytics('abc123');
```

### Python
```bash
pip install url-shortener-python
```

```python
from url_shortener import URLShortenerClient

client = URLShortenerClient(
    base_url='http://localhost:8080',
    api_key='your-api-key'  # when authentication is enabled
)

# Shorten URL
result = client.shorten(
    url='https://www.example.com',
    custom_code='my-link'
)

# Get analytics
analytics = client.get_analytics('abc123', days=30)

# Get attribution report
attribution = client.get_attribution_report('abc123', model='linear')
```

### cURL Examples

#### Shorten URL
```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.example.com",
    "custom_code": "my-link"
  }'
```

#### Get Geographic Analytics
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/abc123/geographic?days=30"
```

#### Create A/B Test
```bash
curl -X POST http://localhost:8080/api/v1/ab-test \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Button Color Test",
    "short_codes": ["test123", "test456"],
    "traffic_split": [0.5, 0.5],
    "metric_name": "click_through_rate"
  }'
```

## Webhooks

Configure webhooks to receive real-time notifications about events.

### Webhook Events

- `url.clicked` - When a short URL is clicked
- `conversion.recorded` - When a conversion is recorded
- `ab_test.completed` - When an A/B test reaches significance
- `ab_test.stopped` - When an A/B test is manually stopped

### Webhook Payload Example

```json
{
  "event": "url.clicked",
  "timestamp": "2023-06-15T15:30:00Z",
  "data": {
    "short_code": "abc123",
    "original_url": "https://www.example.com",
    "user_ip": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "referrer": "https://google.com",
    "country": "US",
    "city": "New York"
  }
}
```

## Best Practices

### URL Shortening
- Use HTTPS URLs for better security and analytics
- Include UTM parameters for better campaign tracking
- Set appropriate expiration dates for time-sensitive links

### Analytics
- Use appropriate time ranges (days parameter) for your needs
- Cache analytics results when possible to reduce API calls
- Monitor rate limits to avoid service interruption

### A/B Testing
- Ensure adequate sample sizes before starting tests
- Use statistical significance analysis before making decisions  
- Consider sequential testing for faster results
- Document test hypotheses and results

### Attribution Modeling
- Record conversion events consistently
- Use appropriate attribution models for your business
- Consider customer journey length when choosing models

### Performance
- Use WebSocket connections for real-time requirements
- Implement client-side caching for frequently accessed data
- Use compression (gzip) for large analytical datasets

## Support

- **Documentation:** [Full documentation](http://localhost:8080/docs)
- **Issues:** [GitHub Issues](https://github.com/your-username/url-shortener/issues)
- **Email:** support@yourcompany.com

## Changelog

### v1.0.0 (Phase 3 - Current)
- Advanced Analytics Dashboard
- Multi-touch Attribution Modeling  
- A/B Testing with Statistical Significance
- Real-time Analytics via WebSocket
- Geographic and Device Analytics
- Conversion Funnel Analysis

### Coming Soon (Phase 4)
- User Authentication & Authorization
- Team Management & Collaboration
- Custom Domains
- Link Management Dashboard

### Future (Phase 5)
- API Rate Limiting & Quotas
- Advanced Security Features
- Enterprise Integration
- Bulk Operations