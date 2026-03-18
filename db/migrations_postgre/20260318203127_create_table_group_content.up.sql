CREATE TABLE group_contents (
    group_content_id SERIAL NOT NULL,
    group_id VARCHAR(191) NOT NULL,
    content_type content_type_enum NOT NULL,
    content_id INTEGER NOT NULL,

    CONSTRAINT group_contents_pkey PRIMARY KEY (group_content_id)
);