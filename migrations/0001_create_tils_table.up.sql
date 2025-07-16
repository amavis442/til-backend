CREATE TABLE tils (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT,
    content TEXT NOT NULL,
    html TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);