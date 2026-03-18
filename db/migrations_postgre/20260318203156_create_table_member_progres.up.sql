CREATE TABLE member_progress (
    progres_id SERIAL NOT NULL,
    member_id INTEGER NULL,
    group_id VARCHAR(191) NULL,
    content_type content_type_enum NOT NULL,
    content_id INTEGER NOT NULL,
    skor INTEGER NULL,
    completed_at TIMESTAMP(0) NULL,
    duration INTEGER NOT NULL,

    CONSTRAINT member_progress_pkey PRIMARY KEY (progres_id)
);