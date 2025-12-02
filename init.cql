DROP KEYSPACE IF EXISTS go_fundraising;

CREATE KEYSPACE go_fundraising
WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor': 1
};

CREATE TABLE IF NOT EXISTS go_fundraising.users (
    id uuid PRIMARY KEY,
    email text,
    username text,
    password_hash text,
    created_at timestamp
);
CREATE INDEX IF NOT EXISTS idx_users_username ON go_fundraising.users(username);

CREATE TABLE IF NOT EXISTS go_fundraising.refresh_tokens (
    refresh_token text PRIMARY KEY,
    user_id UUID,
    expires_at timestamp
);

CREATE TABLE IF NOT EXISTS go_fundraising.comments (
    campaign_id UUID,
    id UUID,
    user_id UUID,
    username text,
    content text,
    created_at timestamp,
    PRIMARY KEY ((campaign_id), created_at, id)
) WITH CLUSTERING ORDER BY (created_at DESC);


CREATE TABLE IF NOT EXISTS go_fundraising.campaigns (
    user_id UUID,
    created_at timestamp,
    id UUID,
    username text,
    title text,
    description text,
    target int,
    amount_collected int,
    image text,
    deadline timestamp,
    PRIMARY KEY (id)
);


CREATE TABLE IF NOT EXISTS go_fundraising.payment_history (
    user_id UUID,
    campaign_id UUID,
    username text,
    id UUID,
    amount int,
    created_at timestamp,
    checkout_id text,
    PRIMARY KEY ((campaign_id), created_at, id)
) WITH CLUSTERING ORDER BY (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_checkout_id
ON go_fundraising.payment_history (checkout_id);