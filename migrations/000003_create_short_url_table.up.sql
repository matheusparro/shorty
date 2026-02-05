-- Create short_urls table
CREATE TABLE IF NOT EXISTS short_urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_url TEXT NOT NULL,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    visit_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    
    -- Foreign key to users table
    CONSTRAINT fk_short_urls_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Index for the most important query: redirect by short_code
-- This is the hot path (100:1 read/write ratio)
CREATE UNIQUE INDEX idx_short_urls_short_code ON short_urls(short_code);

-- Index for finding active URLs quickly (redirect + expiration check)
CREATE INDEX idx_short_urls_active_expires ON short_urls(is_active, expires_at) 
    WHERE is_active = true;

-- Index for user's URLs listing
CREATE INDEX idx_short_urls_user_id ON short_urls(user_id);

-- Index for cleanup of expired URLs
CREATE INDEX idx_short_urls_expires_at ON short_urls(expires_at) 
    WHERE expires_at IS NOT NULL;

-- Trigger to auto-update updated_at
CREATE TRIGGER update_short_urls_updated_at
    BEFORE UPDATE ON short_urls
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Optional: Add check constraint for URL format
ALTER TABLE short_urls 
    ADD CONSTRAINT check_original_url_not_empty 
    CHECK (length(original_url) > 0);

ALTER TABLE short_urls 
    ADD CONSTRAINT check_short_code_format 
    CHECK (short_code ~ '^[a-zA-Z0-9_-]+$');