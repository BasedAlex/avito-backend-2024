-- +goose Up
-- +goose StatementBegin

CREATE TABLE tags (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE features (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE banners (
    id BIGSERIAL PRIMARY KEY,
    feature_id BIGSERIAL REFERENCES features(id),
    content jsonb,
    is_active BOOLEAN DEFAULT FALSE NOT NULL,
    UNIQUE (id, feature_id),
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE banners_tags (
    tag_id BIGSERIAL REFERENCES tags(id),
    banner_id BIGSERIAL REFERENCES banners(id),
    PRIMARY KEY(tag_id, banner_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE banners_tags;
DROP TABLE banners;
DROP TABLE features;
DROP TABLE tags;

-- +goose StatementEnd
