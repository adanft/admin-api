CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT NOT NULL,
    avatar TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    CONSTRAINT users_name_not_blank_chk CHECK (btrim(name) <> ''),
    CONSTRAINT users_last_name_not_blank_chk CHECK (btrim(last_name) <> ''),
    CONSTRAINT users_username_not_blank_chk CHECK (btrim(username) <> ''),
    CONSTRAINT users_email_not_blank_chk CHECK (btrim(email) <> ''),
    CONSTRAINT users_name_length_chk CHECK (char_length(name) <= 80),
    CONSTRAINT users_last_name_length_chk CHECK (char_length(last_name) <= 80),
    CONSTRAINT users_username_length_chk CHECK (char_length(username) BETWEEN 3 AND 30),
    CONSTRAINT users_username_format_chk CHECK (username ~ '^[A-Za-z0-9._-]+$'),
    CONSTRAINT users_email_length_chk CHECK (char_length(email) <= 254),
    CONSTRAINT users_avatar_length_chk CHECK (avatar IS NULL OR char_length(avatar) <= 500)
);

CREATE UNIQUE INDEX users_email_lower_uidx ON users (lower(email));
CREATE INDEX users_created_at_idx ON users (created_at DESC);
CREATE INDEX users_updated_at_idx ON users (updated_at DESC);

CREATE TRIGGER users_set_updated_at_trg
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    CONSTRAINT roles_name_not_blank_chk CHECK (btrim(name) <> ''),
    CONSTRAINT roles_name_length_chk CHECK (char_length(name) BETWEEN 2 AND 50),
    CONSTRAINT roles_description_length_chk CHECK (description IS NULL OR char_length(description) <= 255)
);

CREATE UNIQUE INDEX roles_name_lower_uidx ON roles (lower(name));

CREATE TRIGGER roles_set_updated_at_trg
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    assigned_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX user_roles_role_id_idx ON user_roles (role_id);

INSERT INTO roles (name, description)
VALUES
    ('admin', 'Full administrative access'),
    ('manager', 'Manage operational resources'),
    ('viewer', 'Read-only access')
ON CONFLICT DO NOTHING;
