ALTER TABLE subscriptions
    DROP CONSTRAINT IF EXISTS subscriptions_email_key;

ALTER TABLE subscriptions
    ADD CONSTRAINT subscriptions_email_city_frequency_unique
    UNIQUE (email, city, frequency);
