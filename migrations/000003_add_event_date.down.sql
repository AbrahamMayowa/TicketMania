BEGIN;

-- 1) Add back sales_start/sales_end as timestamptz (will be NULL for existing rows)
ALTER TABLE ticket_types ADD COLUMN IF NOT EXISTS sales_start TIMESTAMPTZ;
ALTER TABLE ticket_types ADD COLUMN IF NOT EXISTS sales_end TIMESTAMPTZ;

-- 2) convert start_time and end_time from time -> timestamptz
-- Reconstruct a timestamptz using the date column when available, otherwise use current_date as fallback.
-- This keeps behaviour deterministic; adjust timezone logic if you need a specific zone.
ALTER TABLE events
  ALTER COLUMN start_time TYPE timestamptz USING ((COALESCE(date, now()::date) + start_time)::timestamptz),
  ALTER COLUMN end_time   TYPE timestamptz USING ((COALESCE(date, now()::date) + end_time)::timestamptz);

-- 3) drop date column
ALTER TABLE events DROP COLUMN IF EXISTS date;

COMMIT;