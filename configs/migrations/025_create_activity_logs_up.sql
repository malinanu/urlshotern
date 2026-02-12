-- Create activity logs table for admin audit trail
CREATE TABLE IF NOT EXISTS activity_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    user_name VARCHAR(255) NOT NULL,
    activity_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX idx_activity_logs_activity_type ON activity_logs(activity_type);

-- Comments
COMMENT ON TABLE activity_logs IS 'Audit trail of user and admin actions';
COMMENT ON COLUMN activity_logs.activity_type IS 'Type: user_login, user_registered, url_created, user_updated, etc.';
