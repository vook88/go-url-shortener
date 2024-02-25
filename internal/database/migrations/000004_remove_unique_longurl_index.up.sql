ALTER TABLE url_mappings DROP CONSTRAINT url_mappings_long_url_key;

CREATE UNIQUE INDEX long_url_user_id_unique ON url_mappings (long_url, user_id);

