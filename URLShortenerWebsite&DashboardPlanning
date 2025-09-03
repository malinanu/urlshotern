# URL Shortener Website & Dashboard Planning
*Complete Frontend and Authentication System Design*

## Table of Contents
1. [Website Structure](#website-structure)
2. [Authentication System](#authentication-system)
3. [User Dashboard](#user-dashboard)
4. [Admin Dashboard](#admin-dashboard)
5. [Database Schema](#database-schema)
6. [SMS Verification Integration](#sms-verification-integration)
7. [UI/UX Design Recommendations](#uiux-design-recommendations)
8. [Technical Implementation](#technical-implementation)
9. [Security Considerations](#security-considerations)

---

## Website Structure

### Public Website Pages

#### 1. Landing Page (`/`)
**Purpose**: Convert visitors to users
**Key Sections**:
- Hero section with URL shortening form
- Value proposition and benefits
- Feature highlights (analytics, custom links, security)
- Pricing tiers (free, premium)
- Customer testimonials/stats
- Call-to-action buttons

#### 2. URL Shortening Tool (`/shorten`)
**Features**:
- Quick URL shortening without registration
- Real-time URL validation
- Instant short URL generation
- Copy to clipboard functionality
- Preview of analytics (limited for anonymous users)
- Registration prompt for advanced features

#### 3. Authentication Pages
- Login (`/login`)
- Register (`/register`) 
- Forgot Password (`/forgot-password`)
- Reset Password (`/reset-password`)
- Phone Verification (`/verify-phone`)
- Email Verification (`/verify-email`)

#### 4. Information Pages
- About Us (`/about`)
- Features (`/features`)
- Pricing (`/pricing`)
- API Documentation (`/docs`)
- Privacy Policy (`/privacy`)
- Terms of Service (`/terms`)
- Contact (`/contact`)

#### 5. Public Analytics (`/stats/{shortCode}`)
**Features**:
- Public view of click statistics (if enabled by creator)
- Geographic distribution
- Click timeline
- No sensitive data exposure

---

## Authentication System

### Registration Flow

#### Step 1: Basic Information
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "phone": "94712345678",
  "terms_accepted": true,
  "marketing_consent": false
}
```

#### Step 2: Email Verification
- Send verification email with token
- User clicks link to verify email
- Account activated but limited until phone verified

#### Step 3: Phone Verification (Text.lk Integration)
- Generate 6-digit OTP
- Send via Text.lk SMS API
- User enters OTP to complete registration
- Full account access granted

### Login Options

#### 1. Email/Password Login
- Standard form with remember me option
- Rate limiting (5 attempts per 15 minutes)
- Account lockout protection
- Password strength requirements

#### 2. Google OAuth Integration
- Google Sign-In button
- Automatic account creation if new user
- Still require phone verification for new Google users
- Link existing accounts option

### Password Requirements
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter  
- At least one number
- At least one special character
- Cannot contain common passwords
- Cannot contain personal information

### Session Management
- JWT tokens for API authentication
- Refresh token rotation
- Session timeout (24 hours for regular, 7 days for "remember me")
- Device tracking for security

---

## User Dashboard

### Dashboard Layout

#### Navigation Structure
```
┌─────────────────────────────────────────────────────────────┐
│ Logo    [Dashboard] [Analytics] [Settings]     [Profile ▼] │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Main Content Area                                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 1. Main Dashboard (`/dashboard`)

**Quick Stats Cards:**
- Total URLs created
- Total clicks this month
- Most popular URL
- Click-through rate

**Quick Actions:**
- URL shortening widget
- Recent URLs table (last 10)
- Quick access to popular features

**Recent Activity Feed:**
- New clicks on your URLs
- Recently created URLs
- System notifications

### 2. URL Management (`/dashboard/urls`)

**Features:**
- Searchable URL list with filters
- Bulk operations (delete, export)
- URL status (active, expired, disabled)
- Quick edit for custom aliases
- QR code generation
- Social sharing buttons

**Table Columns:**
- Short URL (clickable)
- Original URL (truncated with tooltip)
- Created Date
- Clicks
- Status
- Actions (Edit, Delete, Analytics, QR)

### 3. Analytics (`/dashboard/analytics`)

**Overview Page:**
- Date range selector
- Key metrics cards
- Click trends chart
- Geographic heatmap
- Top performing URLs

**Detailed Analytics:**
- Individual URL performance
- Referrer analysis  
- Device/browser breakdown
- Time-based patterns
- Conversion tracking

**Export Options:**
- CSV export
- PDF reports
- API access for data

### 4. Custom Domains (`/dashboard/domains`)
**Features:**
- Add custom domain (premium feature)
- Domain verification process
- SSL certificate management
- Usage statistics per domain

### 5. API Management (`/dashboard/api`)
**Features:**
- API key generation and management
- Usage statistics and quotas
- Rate limit monitoring
- Documentation links
- Code examples

### 6. Settings (`/dashboard/settings`)

**Account Settings:**
- Profile information editing
- Password change
- Phone number update (with re-verification)
- Email preferences
- Account deletion

**Preferences:**
- Default URL expiration
- Analytics privacy settings
- Notification preferences
- Time zone settings

**Billing (Premium Features):**
- Subscription management
- Usage monitoring
- Payment methods
- Invoice history

---

## Admin Dashboard

### Admin Navigation
```
[Overview] [Users] [URLs] [Analytics] [System] [Security] [Settings]
```

### 1. Admin Overview (`/admin`)

**System Health Cards:**
- System uptime
- Total users (active/inactive)
- URLs created today/total
- Server performance metrics

**Recent Activity:**
- New user registrations
- High-traffic URLs
- System alerts
- Failed login attempts

### 2. User Management (`/admin/users`)

**User List Features:**
- Search and filter users
- User status (active, suspended, banned)
- Registration date and last activity
- Account type (free, premium)
- Bulk actions

**User Details:**
- Profile information
- URL creation history
- Usage statistics
- Payment history (if applicable)
- Action logs

### 3. URL Management (`/admin/urls`)

**Features:**
- Global URL search
- Moderation tools (disable malicious URLs)
- Bulk URL operations
- URL categorization
- Suspicious URL flagging

### 4. System Analytics (`/admin/analytics`)

**Metrics:**
- System performance
- Usage trends
- Revenue analytics
- Geographic distribution
- Popular domains and patterns

### 5. Security Center (`/admin/security`)

**Features:**
- Failed login monitoring
- Suspicious activity alerts
- IP blacklist management
- Rate limiting configuration
- Security event logs

### 6. System Configuration (`/admin/system`)

**Settings:**
- System-wide settings
- Feature flags
- Rate limits
- SMS provider configuration
- Email templates
- Maintenance mode

---

## Database Schema

### Extended Schema for Website and Auth

```sql
-- Users table
CREATE TABLE users (
    id BIGINT PRIMARY KEY DEFAULT generate_snowflake_id(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- NULL for OAuth-only users
    phone VARCHAR(20),
    phone_verified BOOLEAN DEFAULT FALSE,
    email_verified BOOLEAN DEFAULT FALSE,
    provider VARCHAR(50) DEFAULT 'email', -- 'email', 'google'
    provider_id VARCHAR(255), -- Google ID if OAuth
    avatar_url TEXT,
    account_type VARCHAR(20) DEFAULT 'free', -- 'free', 'premium', 'enterprise'
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User sessions
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    device_info JSONB,
    ip_address INET,
    user_agent TEXT,
    refresh_token_hash VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Email verification tokens
CREATE TABLE email_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Phone verification (OTP)
CREATE TABLE phone_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    phone VARCHAR(20) NOT NULL,
    otp_code VARCHAR(6) NOT NULL,
    attempts INTEGER DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Password reset tokens
CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Update URL mappings to include user ownership
ALTER TABLE url_mappings ADD COLUMN user_id BIGINT;
ALTER TABLE url_mappings ADD COLUMN is_public BOOLEAN DEFAULT TRUE;
ALTER TABLE url_mappings ADD COLUMN custom_alias VARCHAR(50);
ALTER TABLE url_mappings ADD COLUMN title VARCHAR(200);
ALTER TABLE url_mappings ADD COLUMN description TEXT;
ALTER TABLE url_mappings ADD FOREIGN KEY (user_id) REFERENCES users(id);

-- API keys for programmatic access
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    permissions JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User preferences
CREATE TABLE user_preferences (
    user_id BIGINT PRIMARY KEY,
    default_expiration INTEGER, -- days
    analytics_public BOOLEAN DEFAULT FALSE,
    email_notifications BOOLEAN DEFAULT TRUE,
    marketing_emails BOOLEAN DEFAULT FALSE,
    timezone VARCHAR(50) DEFAULT 'UTC',
    theme VARCHAR(20) DEFAULT 'light',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Audit logs
CREATE TABLE audit_logs (
    id BIGINT PRIMARY KEY DEFAULT generate_snowflake_id(),
    user_id BIGINT,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);
```

---

## SMS Verification Integration

### Go Implementation for Text.lk API

#### SMS Service Structure

```go
// internal/services/sms.go
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type TextLKConfig struct {
    APIKey    string
    BaseURL   string
    SenderID  string
    Timeout   time.Duration
}

type SMSService struct {
    config TextLKConfig
    client *http.Client
}

type SMSRequest struct {
    Recipient string `json:"recipient"`
    Message   string `json:"message"`
    SenderID  string `json:"sender_id"`
    Type      string `json:"type"`
}

type SMSResponse struct {
    Status  string      `json:"status"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

func NewSMSService(config TextLKConfig) *SMSService {
    return &SMSService{
        config: config,
        client: &http.Client{
            Timeout: config.Timeout,
        },
    }
}

func (s *SMSService) SendOTP(phone, otp string) error {
    message := fmt.Sprintf("Your verification code is: %s. Valid for 5 minutes.", otp)
    
    smsReq := SMSRequest{
        Recipient: phone,
        Message:   message,
        SenderID:  s.config.SenderID,
        Type:      "plain",
    }
    
    return s.sendSMS(smsReq)
}

func (s *SMSService) sendSMS(req SMSRequest) error {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return err
    }
    
    httpReq, err := http.NewRequest("POST", s.config.BaseURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    
    httpReq.Header.Set("Authorization", "Bearer "+s.config.APIKey)
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Accept", "application/json")
    
    resp, err := s.client.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    var smsResp SMSResponse
    if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
        return err
    }
    
    if resp.StatusCode != 200 || smsResp.Status != "success" {
        return fmt.Errorf("SMS failed: %s", smsResp.Message)
    }
    
    return nil
}
```

#### OTP Generation and Validation

```go
// internal/services/otp.go
package services

import (
    "crypto/rand"
    "fmt"
    "time"
)

type OTPService struct {
    storage OTPStorage
    sms     *SMSService
}

type OTP struct {
    UserID    int64
    Phone     string
    Code      string
    ExpiresAt time.Time
    Attempts  int
}

func (o *OTPService) GenerateAndSend(userID int64, phone string) error {
    // Generate 6-digit OTP
    code := generateOTPCode()
    
    otp := OTP{
        UserID:    userID,
        Phone:     phone,
        Code:      code,
        ExpiresAt: time.Now().Add(5 * time.Minute),
        Attempts:  0,
    }
    
    // Store OTP
    if err := o.storage.SaveOTP(otp); err != nil {
        return err
    }
    
    // Send SMS
    return o.sms.SendOTP(phone, code)
}

func (o *OTPService) VerifyOTP(userID int64, code string) error {
    otp, err := o.storage.GetOTP(userID)
    if err != nil {
        return err
    }
    
    // Check expiration
    if time.Now().After(otp.ExpiresAt) {
        return fmt.Errorf("OTP expired")
    }
    
    // Check attempts
    if otp.Attempts >= 3 {
        return fmt.Errorf("too many attempts")
    }
    
    // Verify code
    if otp.Code != code {
        o.storage.IncrementAttempts(userID)
        return fmt.Errorf("invalid OTP")
    }
    
    // Mark as verified
    return o.storage.MarkVerified(userID)
}

func generateOTPCode() string {
    bytes := make([]byte, 3)
    rand.Read(bytes)
    return fmt.Sprintf("%06d", int(bytes[0])<<16|int(bytes[1])<<8|int(bytes[2]))%1000000)
}
```

---

## User Dashboard

### 1. Dashboard Home (`/dashboard`)

**Quick Actions Widget:**
```html
┌─────────────────────────────────────────┐
│ Shorten New URL                         │
│ ┌─────────────────────┐ [Shorten]      │
│ │ Enter URL here...   │                │
│ └─────────────────────┘                │
│ ☐ Custom alias  ☐ Set expiration       │
└─────────────────────────────────────────┘
```

**Statistics Overview:**
- Total URLs: 127
- Total Clicks: 3,456
- This Month: 234 clicks
- Top Performer: abc123 (234 clicks)

**Recent URLs Table:**
| Short URL | Original URL | Clicks | Created |
|-----------|--------------|--------|---------|
| abc123 | google.com | 234 | 2 days ago |
| def456 | facebook.com | 89 | 5 days ago |

### 2. URL Manager (`/dashboard/urls`)

**Filter Options:**
- Date range picker
- Status filter (Active, Expired, Disabled)
- Search by URL or alias
- Sort by clicks, date, alphabetical

**Bulk Actions:**
- Select all/none checkboxes
- Export selected to CSV
- Delete selected
- Update expiration dates

**Individual URL Actions:**
- Edit alias and metadata
- View detailed analytics
- Generate QR code
- Copy to clipboard
- Share via social media
- Disable/enable URL

### 3. Analytics Dashboard (`/dashboard/analytics`)

**Charts and Visualizations:**
- Line chart: Clicks over time
- Bar chart: Top 10 URLs by clicks
- Pie chart: Traffic sources
- World map: Geographic distribution
- Heatmap: Click patterns by hour/day

**Filters:**
- Date range (last 7 days, 30 days, custom)
- URL selection (all URLs or specific)
- Geographic filtering
- Device type filtering

### 4. Account Settings (`/dashboard/settings`)

**Profile Section:**
- Edit name and email
- Upload profile picture
- Change password
- Update phone number (requires re-verification)

**Preferences:**
- Default URL settings
- Notification preferences
- Privacy settings
- Theme selection (light/dark)

**Security:**
- Two-factor authentication setup
- Active sessions management
- Login history
- API key management

**Billing (Premium):**
- Current plan details
- Usage statistics
- Upgrade/downgrade options
- Payment method management
- Invoice downloads

---

## Admin Dashboard

### 1. System Overview (`/admin`)

**Real-time Metrics:**
- Active users online
- URLs created per hour
- System response times
- Error rates
- Server resource usage

**System Health:**
- Database connection status
- Redis connection status
- SMS service status
- External API status

### 2. User Management (`/admin/users`)

**User Table:**
| ID | Name | Email | Status | Plan | Created | Actions |
|----|------|-------|--------|------|---------|---------|
| 123 | John Doe | john@email.com | Active | Premium | 2024-01-15 | View, Edit, Suspend |

**User Actions:**
- View detailed user profile
- Suspend/unsuspend account
- Reset user password
- View user's URLs and analytics
- Send notifications to user

**Bulk Operations:**
- Export user data
- Send bulk notifications
- Apply bulk status changes

### 3. Content Moderation (`/admin/moderation`)

**URL Review:**
- Recently created URLs
- Flagged URLs (reported by users)
- Suspicious pattern detection
- Manual review interface

**Actions:**
- Approve/reject URLs
- Add to blacklist
- Contact URL creator
- Generate moderation reports

### 4. Analytics & Reports (`/admin/analytics`)

**Business Metrics:**
- Revenue tracking
- User growth trends
- Feature usage statistics
- Conversion funnels

**System Performance:**
- Response time trends
- Error rate analysis
- Resource utilization
- Capacity planning data

### 5. System Configuration (`/admin/config`)

**Feature Flags:**
- Enable/disable features
- A/B testing controls
- Maintenance mode
- Rate limiting adjustments

**Integrations:**
- SMS provider settings
- Email service configuration
- OAuth provider setup
- Analytics service config

---

## UI/UX Design Recommendations

### Design System

**Color Palette:**
- Primary: Modern blue (#2563eb)
- Secondary: Emerald green (#10b981)
- Success: Green (#22c55e)
- Warning: Amber (#f59e0b)
- Error: Red (#ef4444)
- Neutral: Gray scale (#f8fafc to #1e293b)

**Typography:**
- Headings: Inter or Poppins
- Body: Inter or system fonts
- Code: JetBrains Mono or Fira Code

**Components:**
- Consistent button styles
- Form input standardization
- Loading states and animations
- Toast notifications
- Modal dialogs
- Data tables with sorting

### Responsive Design

**Breakpoints:**
- Mobile: 320px - 768px
- Tablet: 768px - 1024px
- Desktop: 1024px+

**Mobile-First Features:**
- Touch-friendly buttons (44px minimum)
- Swipe gestures for table actions
- Collapsible navigation
- Progressive disclosure

### Accessibility

**Requirements:**
- WCAG 2.1 AA compliance
- Keyboard navigation support
- Screen reader compatibility
- High contrast mode option
- Font size adjustment

---

## Technical Implementation

### Frontend Tech Stack Recommendations

#### Option 1: React with Next.js (Recommended)
**Pros:**
- Server-side rendering for better SEO
- Great ecosystem and component libraries
- TypeScript support
- Vercel deployment optimization

**Libraries:**
- **UI**: Tailwind CSS + Headless UI
- **Charts**: Recharts or Chart.js
- **Forms**: React Hook Form + Zod validation
- **State**: Zustand or React Query
- **Auth**: NextAuth.js for OAuth integration

#### Option 2: Vue.js with Nuxt.js
**Pros:**
- Easier learning curve
- Excellent developer experience
- Great documentation

**Libraries:**
- **UI**: Tailwind CSS + Headless UI Vue
- **Charts**: Vue-Chart.js
- **Forms**: VeeValidate
- **State**: Pinia

#### Option 3: Svelte with SvelteKit
**Pros:**
- Smallest bundle size
- Great performance
- Modern approach

### Backend Integration Points

**API Endpoints Needed:**

```go
// Authentication routes
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
POST   /api/v1/auth/refresh
POST   /api/v1/auth/google
POST   /api/v1/auth/verify-email
POST   /api/v1/auth/verify-phone
POST   /api/v1/auth/resend-verification
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password

// User management
GET    /api/v1/user/profile
PUT    /api/v1/user/profile
GET    /api/v1/user/urls
DELETE /api/v1/user/account

// URL operations
POST   /api/v1/urls
GET    /api/v1/urls
GET    /api/v1/urls/{id}
PUT    /api/v1/urls/{id}
DELETE /api/v1/urls/{id}
GET    /api/v1/urls/{id}/analytics

// Admin routes
GET    /api/v1/admin/users
GET    /api/v1/admin/urls
GET    /api/v1/admin/analytics
PUT    /api/v1/admin/users/{id}/status
DELETE /api/v1/admin/urls/{id}
```

### Development Workflow with Claude Code

**Step-by-step Implementation:**

1. **Authentication System:**
```bash
claude code "Implement complete user authentication system in Go with JWT tokens, Google OAuth, email verification, and phone verification using Text.lk SMS API"
```

2. **User Dashboard Backend:**
```bash
claude code "Create Go API endpoints for user dashboard including URL management, analytics data, user preferences, and profile management"
```

3. **Frontend Components:**
```bash
claude code "Create React components for URL shortener dashboard including URL table, analytics charts, and user settings with Tailwind CSS"
```

4. **Admin Panel:**
```bash
claude code "Implement admin dashboard API and React components for user management, system monitoring, and content moderation"
```

---

## Security Considerations

### Authentication Security

**Password Security:**
- Bcrypt hashing with salt rounds ≥ 12
- Password strength validation
- Password history (prevent reuse of last 5)
- Account lockout after failed attempts

**Session Security:**
- HTTP-only cookies for session tokens
- CSRF protection
- Secure flag for HTTPS
- Session invalidation on password change

**OAuth Security:**
- State parameter validation
- PKCE for mobile apps
- Scope limitation
- Token refresh security

### SMS Security

**OTP Security:**
- 6-digit codes with 5-minute expiration
- Maximum 3 attempts per phone number
- Rate limiting: 1 SMS per minute per phone
- Secure random number generation

**Phone Verification:**
- Validate phone number format
- Check against international format
- Block disposable/virtual numbers
- Log all verification attempts

### API Security

**Rate Limiting:**
- Per-user: 1000 requests/hour
- Per-IP: 100 requests/hour (anonymous)
- SMS: 5 per hour per phone number
- Progressive backoff for violations

**Input Validation:**
- URL validation and sanitization
- SQL injection prevention
- XSS protection
- File upload restrictions

---

## Implementation Timeline

### Phase 1: Authentication & Core Website (4-6 weeks)

**Week 1-2: Authentication Backend**
- User registration/login API
- Email verification system
- SMS verification with Text.lk
- JWT token management

**Week 3-4: Frontend Authentication**
- Registration and login forms
- Email verification flow
- Phone verification interface
- Google OAuth integration

**Week 5-6: Basic Dashboard**
- User dashboard layout
- URL creation interface
- Basic URL management
- Profile settings

### Phase 2: Advanced Features (4-5 weeks)

**Week 7-8: Analytics System**
- Click tracking enhancement
- Analytics API development
- Dashboard charts and graphs
- Export functionality

**Week 9-10: Admin Panel**
- Admin authentication
- User management interface
- System monitoring dashboard
- Content moderation tools

**Week 11: Polish & Testing**
- UI/UX improvements
- Security testing
- Performance optimization
- Bug fixes

### Phase 3: Advanced Features (3-4 weeks)

**Week 12-13: Premium Features**
- Custom domains
- Advanced analytics
- API key management
- Billing integration

**Week 14-15: Mobile Optimization**
- Mobile app (optional)
- PWA features
- Mobile-specific optimizations

---

## Recommended Development Approach

### 1. Start with Claude Code

**Initial Implementation:**
```bash
# Generate authentication system
claude code "Create complete Go authentication system with user registration, email verification, phone verification using Text.lk SMS, Google OAuth, and JWT tokens. Include all necessary handlers, middleware, and database operations."

# Generate React dashboard
claude code "Create React dashboard for URL shortener with user authentication, URL management table, analytics charts, and settings panel using Tailwind CSS and modern React patterns."
```

### 2. Incremental Development

**Week by Week:**
- Week 1: Authentication backend with Claude Code
- Week 2: Frontend authentication with Claude Code  
- Week 3: URL management features
- Week 4: Analytics implementation
- Continue iteratively...

### 3. Testing Strategy

**Local Testing:**
- Unit tests for all services
- Integration tests for API endpoints
- Frontend component testing
- End-to-end testing with Cypress

**Production Testing:**
- Load testing with k6 or Artillery
- Security testing with OWASP ZAP
- SMS delivery testing
- Cross-browser compatibility

This comprehensive plan provides everything needed to build a professional URL shortener with full authentication, dashboards, and SMS verification. The modular approach allows you to build incrementally while maintaining high quality standards.