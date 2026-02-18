CREATE TABLE auth_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    family_id UUID NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    CONSTRAINT auth_refresh_tokens_token_hash_not_blank_chk CHECK (btrim(token_hash) <> '')
);

CREATE UNIQUE INDEX auth_refresh_tokens_token_hash_uidx ON auth_refresh_tokens (token_hash);
CREATE INDEX auth_refresh_tokens_user_id_idx ON auth_refresh_tokens (user_id);
CREATE INDEX auth_refresh_tokens_family_id_idx ON auth_refresh_tokens (family_id);
CREATE INDEX auth_refresh_tokens_expires_at_idx ON auth_refresh_tokens (expires_at);
