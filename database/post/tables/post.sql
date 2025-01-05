CREATE TABLE posts IF NOT EXISTS (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    post_type VARCHAR(50) NOT NULL, -- 'image', 'short', 'carousel', 'story'
    caption TEXT,
    scheduled_time TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'scheduled', -- 'scheduled', 'posted', 'failed', 'draft'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);