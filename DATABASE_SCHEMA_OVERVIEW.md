# Database Schema Overview

## Complete Table Inventory (70 Tables)

### Core Tables (24 tables) - Already Exist via initTables()

#### 1. URL Management (3 tables)
```
url_mappings (id, short_code, original_url, created_at, expires_at, click_count, is_active, created_by_ip)
click_events (id, short_code, clicked_at, ip_address, user_agent, referrer, country_code)
└─ FK: short_code → url_mappings.short_code
```

#### 2. User Management (1 table)
```
users (id, username, name, email, password_hash, phone, phone_verified, email_verified,
       provider, provider_id, avatar_url, account_type, is_active, is_admin,
       last_login_at, last_login_ip, failed_login_attempts, locked_until,
       created_at, updated_at)
```

#### 3. Authentication & Sessions (2 tables)
```
user_sessions (id, user_id, device_info, ip_address, user_agent, refresh_token_hash, expires_at, created_at)
└─ FK: user_id → users.id

api_keys (id, user_id, key_hash, key_prefix, name, description, permissions, rate_limit,
          last_used_at, last_used_ip, expires_at, is_active, created_at, updated_at)
└─ FK: user_id → users.id
```

#### 4. API Tracking (1 table)
```
api_key_usage (id, api_key_id, endpoint, method, status_code, response_time_ms,
               ip_address, user_agent, request_size, response_size, created_at)
└─ FK: api_key_id → api_keys.id
```

#### 5. Audit & Security (1 table)
```
audit_logs (id, user_id, action, resource_type, resource_id, old_values, new_values,
            metadata, ip_address, user_agent, success, error_message, created_at)
└─ FK: user_id → users.id
```

#### 6. Basic Analytics (8 tables)
```
geographic_analytics (id, short_code, country_code, country_name, region, city,
                      latitude, longitude, clicks, unique_ips, last_click, created_at, updated_at)

device_analytics (id, short_code, device_type, device_brand, device_model, os_name, os_version,
                  browser_name, browser_version, screen_resolution, user_agent_hash,
                  clicks, unique_ips, last_click, created_at, updated_at)

time_analytics (id, short_code, click_date, hour_of_day, day_of_week, day_of_month,
                month, year, clicks, unique_ips, created_at, updated_at)

referrer_analytics (id, short_code, referrer_domain, referrer_url, referrer_type,
                    campaign_source, campaign_medium, campaign_name, campaign_term, campaign_content,
                    clicks, unique_clicks, conversions, last_click, created_at, updated_at)

conversion_goals (id, user_id, goal_name, goal_type, target_url, custom_event_name,
                  goal_value, attribution_window, is_active, created_at, updated_at)
└─ FK: user_id → users.id

conversions (id, short_code, goal_id, conversion_id, conversion_type, conversion_value,
             user_ip, user_agent, referrer, session_id, click_id, conversion_time,
             attribution_model, time_to_conversion, created_at)
└─ FK: goal_id → conversion_goals.id

ab_tests (id, user_id, test_name, test_type, description, traffic_split,
          start_date, end_date, sample_size, confidence, is_active, status,
          winner, confidence_level, min_sample_size, conversion_goal_id, created_at, updated_at)
└─ FK: user_id → users.id
└─ FK: conversion_goal_id → conversion_goals.id

ab_test_results (id, test_id, variant_id, session_id, event_type, conversion_value, timestamp)
└─ FK: test_id → ab_tests.id
```

#### 7. Attribution Tracking (1 table)
```
attribution_touchpoints (id, session_id, short_code, user_ip, user_agent, referrer,
                         campaign_source, campaign_medium, campaign_name,
                         touchpoint_order, touchpoint_time, conversion_id, created_at)
```

#### 8. Email Verification (2 tables)
```
email_verifications (id, user_id, email, token, expires_at, verified_at, created_at)
└─ FK: user_id → users.id

phone_verifications (id, user_id, phone, code, expires_at, verified_at, attempts, created_at)
└─ FK: user_id → users.id
```

#### 9. Password Reset (1 table)
```
password_resets (id, user_id, token, expires_at, used_at, ip_address, created_at)
└─ FK: user_id → users.id
```

#### 10. Polar Integration (3 tables)
```
polar_subscriptions (id, user_id, polar_subscription_id, status, current_period_start,
                     current_period_end, cancel_at_period_end, created_at, updated_at)
└─ FK: user_id → users.id

polar_products (id, polar_product_id, name, description, price, interval, created_at, updated_at)

polar_webhook_events (id, event_id, event_type, payload, processed, error, created_at)
```

---

## New Tables (46 tables) - Created by Migrations 012-020

### Migration 013: RBAC Tables (5 tables)

```
oauth_providers (id, user_id, provider, provider_user_id, provider_email,
                 access_token, refresh_token, expires_at, created_at, updated_at)
└─ FK: user_id → users.id

roles (id, name, display_name, description, is_system, is_active, created_at, updated_at)

permissions (id, name, display_name, description, resource, action, is_system, created_at)

role_permissions (id, role_id, permission_id, created_at)
├─ FK: role_id → roles.id
└─ FK: permission_id → permissions.id

user_roles (id, user_id, role_id, assigned_by, assigned_at, expires_at)
├─ FK: user_id → users.id
├─ FK: role_id → roles.id
└─ FK: assigned_by → users.id
```

**Default Roles:**
- admin (full access)
- user (standard access)
- premium (enhanced features)
- team_admin (team management)
- team_member (team access)

### Migration 014: Team Management (3 tables)

```
teams (id, name, slug, description, logo_url, owner_id, settings,
       is_active, created_at, updated_at)
└─ FK: owner_id → users.id

team_members (id, team_id, user_id, role, permissions, invited_by, joined_at)
├─ FK: team_id → teams.id
├─ FK: user_id → users.id
└─ FK: invited_by → users.id

team_invitations (id, team_id, email, role, permissions, invited_by,
                  invitation_token, expires_at, accepted_at, accepted_by,
                  declined_at, created_at)
├─ FK: team_id → teams.id
├─ FK: invited_by → users.id
└─ FK: accepted_by → users.id
```

### Migration 015: URL Collaboration (9 tables)

```
url_shares (id, url_id, sharer_id, shared_with_type, shared_with_id,
            share_type, expires_at, is_active, created_at, updated_at)
├─ FK: url_id → url_mappings.id
└─ FK: sharer_id → users.id

url_comments (id, url_id, user_id, content, parent_id, is_edited,
              is_deleted, created_at, updated_at, deleted_at)
├─ FK: url_id → url_mappings.id
├─ FK: user_id → users.id
└─ FK: parent_id → url_comments.id (self-reference)

url_bookmarks (id, user_id, url_id, tags, notes, created_at, updated_at)
├─ FK: user_id → users.id
└─ FK: url_id → url_mappings.id

url_favorites (id, user_id, url_id, created_at)
├─ FK: user_id → users.id
└─ FK: url_id → url_mappings.id

url_collections (id, user_id, team_id, name, description, color, icon,
                 is_public, sort_order, created_at, updated_at, deleted_at)
├─ FK: user_id → users.id
└─ FK: team_id → teams.id

url_collection_items (id, collection_id, url_id, sort_order, added_by, created_at)
├─ FK: collection_id → url_collections.id
├─ FK: url_id → url_mappings.id
└─ FK: added_by → users.id

url_notes (id, user_id, url_id, title, content, is_private, created_at, updated_at)
├─ FK: user_id → users.id
└─ FK: url_id → url_mappings.id

url_activities (id, url_id, user_id, activity_type, description, metadata,
                ip_address, user_agent, created_at)
├─ FK: url_id → url_mappings.id
└─ FK: user_id → users.id

user_activity_logs (id, user_id, session_id, activity_type, description, url_id,
                    metadata, ip_address, user_agent, device_type, browser, platform,
                    location, duration_ms, screen_resolution, timezone,
                    is_mobile, is_bot, created_at)
├─ FK: user_id → users.id
└─ FK: url_id → url_mappings.id
```

### Migration 016: Custom Domains (2 tables)

```
custom_domains (id, user_id, team_id, domain, verification_method, verification_token,
                is_verified, verified_at, ssl_enabled, ssl_cert_path, ssl_key_path,
                ssl_expires_at, health_check_url, last_health_check, health_status,
                dns_records, settings, is_active, created_at, updated_at)
├─ FK: user_id → users.id
└─ FK: team_id → teams.id

domain_analytics (id, domain_id, total_urls, total_clicks, unique_visitors,
                  date, created_at)
└─ FK: domain_id → custom_domains.id
```

### Migration 017: CMS Tables (8 tables)

```
media_files (id, user_id, filename, original_name, file_path, file_url,
             file_size, mime_type, file_type, width, height, alt, caption,
             is_public, created_at, updated_at)
└─ FK: user_id → users.id

static_pages (id, slug, title, content, content_blocks, meta_description,
              meta_keywords, meta_image, meta_title, canonical_url, schema_markup,
              is_published, sort_order, template, author_id, published_at,
              scheduled_at, expires_at, featured_image_id, category, tags,
              custom_fields, view_count, is_sticky, allow_comments, language,
              created_at, updated_at)
├─ FK: author_id → users.id
└─ FK: featured_image_id → media_files.id

page_revisions (id, page_id, title, content, meta_description, meta_keywords,
                revision_number, author_id, created_at)
├─ FK: page_id → static_pages.id
└─ FK: author_id → users.id

page_visits (id, page_id, ip_address, user_agent, referrer, country_code, visited_at)
└─ FK: page_id → static_pages.id

global_settings (id, key, value, type, category, display_name, description,
                 is_public, sort_order, created_at, updated_at)

content_blocks (id, name, identifier, content, type, is_active, sort_order,
                created_at, updated_at)

page_templates (id, name, slug, description, content, fields, preview_url,
                is_active, sort_order, created_at, updated_at)

navigation_menus (id, name, location, items, is_active, sort_order,
                  created_at, updated_at)
```

### Migration 018: Billing Tables (7 tables)

```
subscriptions (id, user_id, plan_type, status, stripe_customer_id,
               stripe_subscription_id, current_period_start, current_period_end,
               cancel_at_period_end, trial_end, billing_cycle, metadata,
               created_at, updated_at)
└─ FK: user_id → users.id

billing_events (id, user_id, subscription_id, event_type, stripe_event_id,
                amount, currency, status, metadata, processed_at, created_at)
├─ FK: user_id → users.id
└─ FK: subscription_id → subscriptions.id

usage_tracking (id, user_id, tracking_period, urls_created, urls_active,
                clicks_count, api_calls_count, custom_domains_count, last_updated)
└─ FK: user_id → users.id

payment_methods (id, user_id, stripe_payment_method_id, type, is_default,
                 card_brand, card_last4, card_exp_month, card_exp_year,
                 billing_address, created_at, updated_at)
└─ FK: user_id → users.id

coupons (id, code, name, description, discount_type, discount_value, currency,
         max_redemptions, current_redemptions, valid_from, valid_until,
         applicable_plans, is_active, created_at, updated_at)

coupon_redemptions (id, coupon_id, user_id, subscription_id, redeemed_at)
├─ FK: coupon_id → coupons.id
├─ FK: user_id → users.id
└─ FK: subscription_id → subscriptions.id

billing_analytics (materialized view)
```

**Functions:**
- `update_usage_tracking()` - Update usage metrics
- `get_current_month_usage()` - Get current usage
- `create_default_subscription()` - Auto-create free plan
- `refresh_billing_analytics()` - Refresh analytics view

### Migration 019: Advanced Analytics (14 tables)

```
touchpoint_attributions (id, touchpoint_id, attribution_model, attribution_value,
                         weight, created_at, updated_at)
└─ FK: touchpoint_id → attribution_touchpoints.id

attribution_model_performance (id, short_code, attribution_model, total_conversions,
                               total_attributed_value, avg_attribution_accuracy,
                               model_confidence, date_range_start, date_range_end,
                               created_at, updated_at)

realtime_subscriptions (id, short_code, connection_id, user_id,
                        subscribed_at, last_ping)

hourly_analytics (id, short_code, hour_timestamp, total_clicks, unique_visitors,
                  unique_countries, created_at, updated_at)

daily_analytics (id, short_code, date, total_clicks, unique_visitors,
                 unique_countries, top_country, top_referrer, created_at, updated_at)

weekly_analytics (id, short_code, week_start, week_end, total_clicks,
                  unique_visitors, unique_countries, created_at, updated_at)

monthly_analytics (id, short_code, month_start, month_end, total_clicks,
                   unique_visitors, unique_countries, created_at, updated_at)

country_analytics_summary (id, short_code, country_code, total_clicks,
                           unique_visitors, last_click, created_at, updated_at)

device_type_summary (id, short_code, device_type, total_clicks,
                     unique_visitors, created_at, updated_at)

referrer_summary (id, short_code, referrer_domain, total_clicks,
                  unique_visitors, created_at, updated_at)

analytics_aggregation_state (id, aggregation_type, last_processed_timestamp,
                             status, error_message, created_at, updated_at)

daily_analytics_summary (materialized view)
```

**Functions:**
- `refresh_daily_analytics_summary()` - Refresh analytics view
- `cleanup_old_analytics_data()` - Data retention cleanup

### Migration 020: URL Mappings Extensions (13 new columns)

```
url_mappings (existing columns + new columns):
  + user_id BIGINT (FK → users.id)
  + team_id BIGINT (FK → teams.id)
  + title VARCHAR(255)
  + description TEXT
  + tags JSONB
  + is_public BOOLEAN
  + password_hash VARCHAR(255)
  + custom_domain VARCHAR(100)
  + utm_source VARCHAR(100)
  + utm_medium VARCHAR(100)
  + utm_campaign VARCHAR(100)
  + utm_term VARCHAR(100)
  + utm_content VARCHAR(100)
```

**View:**
```
urls (view) → url_mappings (alias for backward compatibility)
```

---

## Database Statistics

| Category | Tables | Views | Functions | Triggers |
|----------|--------|-------|-----------|----------|
| Core URL Management | 2 | 1 | 0 | 0 |
| User & Auth | 6 | 0 | 0 | 0 |
| RBAC | 5 | 0 | 0 | 0 |
| Teams | 3 | 0 | 0 | 0 |
| Collaboration | 9 | 0 | 0 | 0 |
| Custom Domains | 2 | 0 | 0 | 0 |
| CMS | 8 | 0 | 0 | 0 |
| Billing | 6 | 1 | 4 | 1 |
| Analytics (Basic) | 8 | 0 | 0 | 0 |
| Analytics (Advanced) | 13 | 1 | 2 | 0 |
| Attribution | 2 | 0 | 0 | 0 |
| Polar Integration | 3 | 0 | 0 | 0 |
| Verification | 2 | 0 | 0 | 0 |
| **TOTAL** | **69** | **2** | **6** | **1** |

## Foreign Key Relationships

### User-Centric Relationships
```
users (1) ──── (∞) user_sessions
users (1) ──── (∞) api_keys
users (1) ──── (∞) oauth_providers
users (1) ──── (∞) user_roles
users (1) ──── (∞) teams (as owner)
users (1) ──── (∞) team_members
users (1) ──── (∞) subscriptions
users (1) ──── (∞) url_mappings
users (1) ──── (∞) custom_domains
users (1) ──── (∞) static_pages (as author)
```

### URL-Centric Relationships
```
url_mappings (1) ──── (∞) click_events
url_mappings (1) ──── (∞) url_shares
url_mappings (1) ──── (∞) url_comments
url_mappings (1) ──── (∞) url_bookmarks
url_mappings (1) ──── (∞) url_favorites
url_mappings (1) ──── (∞) url_collection_items
url_mappings (1) ──── (∞) url_notes
url_mappings (1) ──── (∞) url_activities
```

### Team Relationships
```
teams (1) ──── (∞) team_members
teams (1) ──── (∞) team_invitations
teams (1) ──── (∞) url_mappings
teams (1) ──── (∞) url_collections
teams (1) ──── (∞) custom_domains
```

### Billing Relationships
```
subscriptions (1) ──── (∞) billing_events
subscriptions (1) ──── (∞) coupon_redemptions
users (1) ──── (∞) payment_methods
coupons (1) ──── (∞) coupon_redemptions
```

### Analytics Relationships
```
conversion_goals (1) ──── (∞) conversions
ab_tests (1) ──── (∞) ab_test_results
attribution_touchpoints (1) ──── (∞) touchpoint_attributions
```

---

## Index Summary

**Total Indexes:** ~150+

**Performance-Critical Indexes:**
- `idx_short_code` on url_mappings (UNIQUE)
- `idx_short_code_time` on click_events (composite)
- `idx_users_email` on users (UNIQUE)
- `idx_user_sessions_user_id` on user_sessions
- `idx_api_keys_key_hash` on api_keys
- All foreign key columns
- All lookup fields (country_code, device_type, etc.)
- All timestamp columns for time-series queries

---

## Summary

✅ **70 Total Tables**
✅ **2 Materialized Views**
✅ **6 Functions**
✅ **1 Trigger**
✅ **150+ Indexes**
✅ **Complete RBAC System**
✅ **Team Management**
✅ **Advanced Analytics**
✅ **Billing & Subscriptions**
✅ **CMS Platform**
✅ **Custom Domains**
✅ **Collaboration Features**

This represents a **complete enterprise-grade URL shortening platform** with full feature parity across all modules.
