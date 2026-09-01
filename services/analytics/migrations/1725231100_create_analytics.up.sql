-- Analytics service aggregates data from other services
-- This service does not own its own tables; it queries cross-service data
-- via the database views below for read-only analytics.
-- In production these would be materialized views or fed by event consumers.

-- Placeholder: analytics is read-only, no own tables needed.
SELECT 1;
