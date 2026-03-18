CREATE TABLE puzzles (
    puzzle_id SERIAL NOT NULL,
    judul VARCHAR(100) NOT NULL,
    gambar VARCHAR(255) NOT NULL,
    kategori kategori_puzzle_enum NOT NULL,
    xp_reward INTEGER NOT NULL DEFAULT 80,
    created_by INTEGER NULL,
    created_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_published BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT puzzles_pkey PRIMARY KEY (puzzle_id)
);