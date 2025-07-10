-- Enum type
CREATE TYPE public.platform AS ENUM (
    'google',
    'tiktok',
    'instagram',
    'youtube'
);

-- Sequences
CREATE SEQUENCE public.api_keys_id_seq START 1;
CREATE SEQUENCE public.media_assets_id_seq START 1;
CREATE SEQUENCE public.posting_history_id_seq START 1;
CREATE SEQUENCE public.posts_id_seq START 1;
CREATE SEQUENCE public.social_accounts_id_seq START 1;
CREATE SEQUENCE public.subscriptions_id_seq START 1;
CREATE SEQUENCE public.users_id_seq START 1;

-- Tables
CREATE TABLE public.api_keys (
    id integer NOT NULL DEFAULT nextval('public.api_keys_id_seq'::regclass),
    user_id integer NOT NULL,
    api_key varchar(100) NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT api_keys_pkey PRIMARY KEY (id),
    CONSTRAINT api_keys_api_key_key UNIQUE (api_key)
);

CREATE TABLE public.media_assets (
    id integer NOT NULL DEFAULT nextval('public.media_assets_id_seq'::regclass),
    user_id integer,
    file_name varchar(255) NOT NULL,
    file_type varchar(50) NOT NULL,
    file_size integer NOT NULL,
    file_url text NOT NULL,
    thumbnail_url text,
    duration integer,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT media_assets_pkey PRIMARY KEY (id)
);

CREATE TABLE public.post_media (
    post_id integer NOT NULL,
    asset_id integer NOT NULL,
    display_order integer NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT post_media_pkey PRIMARY KEY (post_id, asset_id)
);

CREATE TABLE public.posting_history (
    id integer NOT NULL DEFAULT nextval('public.posting_history_id_seq'::regclass),
    user_id integer NOT NULL,
    post_id integer NOT NULL,
    account_id integer NOT NULL,
    error_message text,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT posting_history_pkey PRIMARY KEY (id)
);

CREATE TABLE public.posts (
    id integer NOT NULL DEFAULT nextval('public.posts_id_seq'::regclass),
    user_id integer NOT NULL,
    post_type varchar(50) NOT NULL,
    caption text,
    scheduled_time timestamp NOT NULL,
    status varchar(20) DEFAULT 'scheduled',
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    title text,
    CONSTRAINT posts_pkey PRIMARY KEY (id)
);

CREATE TABLE public.selected_accounts (
    post_id integer NOT NULL,
    account_id integer NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT selected_accounts_pkey PRIMARY KEY (post_id, account_id)
);

CREATE TABLE public.social_accounts (
    id integer NOT NULL DEFAULT nextval('public.social_accounts_id_seq'::regclass),
    user_id integer,
    platform public.platform NOT NULL,
    account_id varchar(100) NOT NULL,
    account_name varchar(100) NOT NULL,
    account_username varchar(50) NOT NULL,
    profile_picture_url text,
    access_token text,
    refresh_token text,
    token_expires_at timestamp,
    account_status varchar(20) DEFAULT 'active',
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT social_accounts_pkey PRIMARY KEY (id),
    CONSTRAINT social_accounts_account_id_key UNIQUE (account_id)
);

CREATE TABLE public.subscriptions (
    id integer NOT NULL DEFAULT nextval('public.subscriptions_id_seq'::regclass),
    user_id integer NOT NULL,
    subscription_id varchar(100) NOT NULL,
    subscription_end_date timestamp,
    status varchar(50),
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT subscriptions_pkey PRIMARY KEY (id),
    CONSTRAINT subscriptions_user_id_subscription_id_key UNIQUE (user_id, subscription_id)
);

CREATE TABLE public.users (
    id integer NOT NULL DEFAULT nextval('public.users_id_seq'::regclass),
    google_id varchar(50) NOT NULL,
    email varchar(255) NOT NULL,
    name varchar(100) NOT NULL,
    profile_picture varchar(255) NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_email_key UNIQUE (email)
);

-- Foreign Keys
ALTER TABLE public.api_keys
    ADD CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE public.media_assets
    ADD CONSTRAINT media_assets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE public.post_media
    ADD CONSTRAINT post_media_asset_id_fkey FOREIGN KEY (asset_id) REFERENCES public.media_assets(id) ON DELETE CASCADE,
    ADD CONSTRAINT post_media_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE CASCADE;

ALTER TABLE public.posting_history
    ADD CONSTRAINT posting_history_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.social_accounts(id) ON DELETE CASCADE,
    ADD CONSTRAINT posting_history_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE CASCADE,
    ADD CONSTRAINT posting_history_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE public.posts
    ADD CONSTRAINT posts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


ALTER TABLE public.selected_accounts
    ADD CONSTRAINT selected_accounts_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE CASCADE,
    ADD CONSTRAINT selected_accounts_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.social_accounts(id) ON DELETE CASCADE;

ALTER TABLE public.social_accounts
    ADD CONSTRAINT social_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE public.subscriptions
    ADD CONSTRAINT subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
