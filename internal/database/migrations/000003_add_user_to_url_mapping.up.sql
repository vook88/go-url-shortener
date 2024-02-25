ALTER TABLE url_mappings
    ADD COLUMN user_id INTEGER,
    ADD CONSTRAINT fk_user_id
        FOREIGN KEY (user_id) REFERENCES users(id);
