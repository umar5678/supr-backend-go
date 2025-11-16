-- Drop trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop table
DROP TABLE IF EXISTS users;

-- Drop types
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;