BEGIN;

-- Enums
CREATE TYPE event_status AS ENUM ('draft', 'published', 'cancelled', 'completed');
CREATE TYPE ticket_status AS ENUM ('available', 'reserved', 'paid', 'cancelled', 'used');



CREATE INDEX IF NOT EXISTS ix_users_email ON users(email);

-- Events table
CREATE TABLE IF NOT EXISTS events (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT,
  location TEXT,
  start_time TIMESTAMPTZ NOT NULL,
  end_time TIMESTAMPTZ NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status event_status NOT NULL DEFAULT 'draft',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_events_user_id ON events(user_id);
CREATE INDEX IF NOT EXISTS ix_events_start_time ON events(start_time);

-- Ticket types table
CREATE TABLE IF NOT EXISTS ticket_types (
  id BIGSERIAL PRIMARY KEY,
  event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  price BIGINT NOT NULL DEFAULT 0,       -- price in smallest currency unit
  currency VARCHAR(8) NOT NULL DEFAULT 'USD',
  total_qty INTEGER NOT NULL DEFAULT 0,
  sold_qty INTEGER NOT NULL DEFAULT 0,
  sales_start TIMESTAMPTZ,
  sales_end TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_ticket_types_event_id ON ticket_types(event_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_ticket_types_event_name ON ticket_types(event_id, lower(name));

-- Tickets table
CREATE TABLE IF NOT EXISTS tickets (
  id BIGSERIAL PRIMARY KEY,
  event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  ticket_type_id BIGINT NOT NULL REFERENCES ticket_types(id) ON DELETE CASCADE,
  user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
  status ticket_status NOT NULL DEFAULT 'available',
  paid_at TIMESTAMPTZ,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  buyer_email TEXT,
  buyer_phone TEXT
);

CREATE INDEX IF NOT EXISTS ix_tickets_event_id ON tickets(event_id);
CREATE INDEX IF NOT EXISTS ix_tickets_ticket_type_id ON tickets(ticket_type_id);
CREATE INDEX IF NOT EXISTS ix_tickets_user_id ON tickets(user_id);
CREATE INDEX IF NOT EXISTS ix_tickets_status ON tickets(status);

COMMIT;
