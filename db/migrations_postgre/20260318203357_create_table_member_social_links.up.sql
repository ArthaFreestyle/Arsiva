CREATE TABLE member_social_links (
    social_id SERIAL NOT NULL,
    member_id INTEGER NULL,
    platform platform_enum NOT NULL,
    url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT member_social_links_pkey PRIMARY KEY (social_id)
);