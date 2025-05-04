CREATE TABLE IF NOT EXISTS telegram_id_map (
    telegram_xid CHAR(64) PRIMARY KEY,
    encrypted_id BYTEA NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

GRANT SELECT, INSERT, UPDATE, DELETE ON telegram_id_map TO murmapp_caster_user;
GRANT USAGE ON SCHEMA public TO murmapp_caster_user;
