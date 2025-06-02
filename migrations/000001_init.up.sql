CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    email VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    frequency VARCHAR(255) NOT NULL,
    confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    last_sent_at TIMESTAMPTZ DEFAULT NULL,

    UNIQUE(email),
    UNIQUE(token)
);
