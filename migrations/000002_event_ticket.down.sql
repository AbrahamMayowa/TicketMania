BEGIN;
DROP INDEX IF EXISTS ix_tickets_status;
DROP INDEX IF EXISTS ix_tickets_user_id;
DROP INDEX IF EXISTS ix_tickets_ticket_type_id;
DROP INDEX IF EXISTS ix_tickets_event_id;

DROP INDEX IF EXISTS ux_ticket_types_event_name;
DROP INDEX IF EXISTS ix_ticket_types_event_id;

DROP INDEX IF EXISTS ix_events_start_time;
DROP INDEX IF EXISTS ix_events_user_id;

-- Don't drop users table; only remove the users email index if it was created by this migration
DROP INDEX IF EXISTS ix_users_email;

DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS ticket_types;
DROP TABLE IF EXISTS events;

DROP TYPE IF EXISTS ticket_status;
DROP TYPE IF EXISTS event_status;
COMMIT;