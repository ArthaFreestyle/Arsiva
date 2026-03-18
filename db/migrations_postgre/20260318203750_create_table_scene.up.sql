CREATE TABLE scene (
    scene_id SERIAL NOT NULL,
    cerita_id INTEGER NULL,
    scene_key VARCHAR(50) NOT NULL,
    scene_image VARCHAR(200) NULL,
    scene_text TEXT NOT NULL,
    scene_choices JSONB NULL,
    is_ending BOOLEAN NOT NULL DEFAULT false,
    ending_point INTEGER NOT NULL DEFAULT 0,
    ending_type VARCHAR(50) NULL,
    urutan INTEGER NULL,

    CONSTRAINT scene_pkey PRIMARY KEY (scene_id)
);