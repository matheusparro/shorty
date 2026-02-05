-- Drop trigger
DROP TRIGGER IF EXISTS update_short_urls_updated_at ON short_urls;

-- Drop indexes
DROP INDEX IF EXISTS idx_short_urls_expires_at;
DROP INDEX IF EXISTS idx_short_urls_user_id;
DROP INDEX IF EXISTS idx_short_urls_active_expires;
DROP INDEX IF EXISTS idx_short_urls_short_code;

-- Drop table
DROP TABLE IF EXISTS short_urls;