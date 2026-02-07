BEGIN;

-- 1) add date column (nullable), populate from existing start_time
ALTER TABLE events ADD COLUMN IF NOT EXISTS date DATE;

UPDATE events
SET date = start_time::date
WHERE date IS NULL AND start_time IS NOT NULL;

-- set NOT NULL only when it's safe (no remaining NULLs)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM events WHERE date IS NULL) THEN
    ALTER TABLE events ALTER COLUMN date SET NOT NULL;
  END IF;
END$$;

-- 2) convert start_time and end_time from timestamptz -> time (store only time-of-day)
-- use USING to preserve the time-of-day portion
ALTER TABLE events
  ALTER COLUMN start_time TYPE time USING (start_time::time),
  ALTER COLUMN end_time   TYPE time USING (end_time::time);

-- 3) remove sales_start/sales_end from ticket_types
ALTER TABLE ticket_types DROP COLUMN IF EXISTS sales_start;
ALTER TABLE ticket_types DROP COLUMN IF EXISTS sales_end;

COMMIT;