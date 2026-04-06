CREATE TABLE groups (
    group_id VARCHAR(191) NOT NULL,
    group_name VARCHAR(191) NULL,
    group_thumbnail_asset_id INTEGER NULL,
    created_by INTEGER NULL,
    created_at TIMESTAMP(0) NULL,
    updated_at TIMESTAMP(0) NULL,

    CONSTRAINT groups_pkey PRIMARY KEY (group_id)
);