-- Migration: URL Collaboration Tables
-- Category: Collaboration
-- Tables: url_shares, url_comments, url_bookmarks, url_favorites, url_collections, url_collection_items, url_notes, url_activities, user_activity_logs
-- Dependencies: url_mappings, users tables

-- URL shares table
CREATE TABLE IF NOT EXISTS url_shares (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    sharer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with_type VARCHAR(20) NOT NULL,
    shared_with_id BIGINT,
    share_type VARCHAR(20) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- URL comments table
CREATE TABLE IF NOT EXISTS url_comments (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id BIGINT REFERENCES url_comments(id),
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- URL bookmarks table
CREATE TABLE IF NOT EXISTS url_bookmarks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    tags TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, url_id)
);

-- URL favorites table
CREATE TABLE IF NOT EXISTS url_favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, url_id)
);

-- URL collections table
CREATE TABLE IF NOT EXISTS url_collections (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id BIGINT REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#3b82f6',
    icon VARCHAR(50) DEFAULT 'folder',
    is_public BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- URL collection items junction table
CREATE TABLE IF NOT EXISTS url_collection_items (
    id BIGSERIAL PRIMARY KEY,
    collection_id BIGINT NOT NULL REFERENCES url_collections(id) ON DELETE CASCADE,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    added_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(collection_id, url_id)
);

-- URL notes table
CREATE TABLE IF NOT EXISTS url_notes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    title VARCHAR(200),
    content TEXT NOT NULL,
    is_private BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- URL activity table
CREATE TABLE IF NOT EXISTS url_activities (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id),
    activity_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User activity logs table
CREATE TABLE IF NOT EXISTS user_activity_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(128),
    activity_type VARCHAR(50) NOT NULL,
    description TEXT,
    url_id BIGINT REFERENCES url_mappings(id),
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    device_type VARCHAR(20),
    browser VARCHAR(50),
    platform VARCHAR(50),
    location VARCHAR(100),
    duration_ms BIGINT,
    screen_resolution VARCHAR(20),
    timezone VARCHAR(50),
    is_mobile BOOLEAN DEFAULT FALSE,
    is_bot BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_url_shares_url_id ON url_shares(url_id);
CREATE INDEX IF NOT EXISTS idx_url_shares_sharer_id ON url_shares(sharer_id);
CREATE INDEX IF NOT EXISTS idx_url_comments_url_id ON url_comments(url_id);
CREATE INDEX IF NOT EXISTS idx_url_comments_user_id ON url_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_url_bookmarks_user_id ON url_bookmarks(user_id);
CREATE INDEX IF NOT EXISTS idx_url_favorites_user_id ON url_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_url_collection_items_collection_id ON url_collection_items(collection_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_user_id ON user_activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_session_id ON user_activity_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_activity_type ON user_activity_logs(activity_type);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_created_at ON user_activity_logs(created_at);
