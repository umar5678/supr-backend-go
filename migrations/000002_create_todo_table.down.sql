DROP TRIGGER IF EXISTS update_todos_updated_at ON todos;
DROP INDEX IF EXISTS idx_todos_deleted_at;
DROP INDEX IF EXISTS idx_todos_completed;
DROP INDEX IF EXISTS idx_todos_user_id;
DROP TABLE IF EXISTS todos;