-- ============================================================================
-- PRODUCTION DELETE SCRIPT - Account with phone: +923701653058
-- ============================================================================
-- WARNING: This will permanently delete the user and all associated data!
-- BACKUP YOUR DATABASE BEFORE RUNNING THIS SCRIPT!
-- ============================================================================

-- Set timezone to match your system
SET TIME ZONE 'UTC';

-- Store the user ID for reference
DO $$
DECLARE
    v_user_id UUID;
    v_phone VARCHAR(20) := '+923701653058';
BEGIN
    -- Find the user
    SELECT id INTO v_user_id FROM users WHERE phone = v_phone;
    
    IF v_user_id IS NULL THEN
        RAISE NOTICE 'User with phone % not found', v_phone;
        RETURN;
    END IF;
    
    RAISE NOTICE 'Found user: % with ID: %', v_phone, v_user_id;
    RAISE NOTICE 'Starting deletion cascade...';
    
    -- The DELETE FROM users will cascade to all related tables due to foreign key constraints
    -- The order doesn't matter because of CASCADE on delete
    
    DELETE FROM users WHERE id = v_user_id;
    
    RAISE NOTICE 'User deleted successfully!';
    RAISE NOTICE 'All related data has been deleted due to CASCADE constraints';
END $$;

-- Verify the user is gone
SELECT 'Verification: User count with phone +923701653058:' as result, 
       COUNT(*) as count 
FROM users 
WHERE phone = '+923701653058';

-- Show deletion summary
SELECT 'Account deletion completed successfully' as status;
