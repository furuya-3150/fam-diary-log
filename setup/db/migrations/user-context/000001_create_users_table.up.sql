CREATE TABLE
  users (
    id UUID NOT NULL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT unique_provider_id UNIQUE (provider, provider_id)
  );

-- Index for email lookup (login, duplicate check)
CREATE INDEX idx_users_email ON users (email);

-- Index for provider ID lookup (OAuth authentication)
CREATE INDEX idx_users_provider_id ON users (provider, provider_id);
