
CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS features (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS banners (
    id BIGSERIAL PRIMARY KEY,
    feature_id BIGSERIAL REFERENCES features(id),
    content jsonb,
    is_active BOOLEAN DEFAULT FALSE NOT NULL,
    UNIQUE (id, feature_id),
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS banners_tags (
    tag_id BIGSERIAL REFERENCES tags(id),
    banner_id BIGSERIAL REFERENCES banners(id),
    PRIMARY KEY(tag_id, banner_id)
);