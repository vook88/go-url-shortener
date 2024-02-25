CREATE TABLE url_mappings (
    id SERIAL PRIMARY KEY,
    long_url TEXT NOT NULL UNIQUE,
    short_url VARCHAR(255) NOT NULL UNIQUE
);