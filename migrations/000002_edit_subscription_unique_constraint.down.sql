ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_email_city_frequency_unique;

ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_email_key UNIQUE (email);
