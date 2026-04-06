CREATE TABLE puzzles (
    puzzle_id SERIAL NOT NULL,
    judul VARCHAR(100) NOT NULL,
    gambar_asset_id INTEGER NOT NULL,
    thumbnail_asset_id INTEGER NOT NULL,
    kategori kategori_puzzle_enum NOT NULL,
    xp_reward INTEGER NOT NULL DEFAULT 80,
    created_by INTEGER NULL,
    created_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_published BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT puzzles_pkey PRIMARY KEY (puzzle_id)
);