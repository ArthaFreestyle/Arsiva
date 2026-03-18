CREATE TABLE kuis (
    kuis_id SERIAL NOT NULL,
    judul VARCHAR(100) NOT NULL,
    deskripsi TEXT NULL,
    kategori_id INTEGER NULL,
    xp_reward INTEGER NOT NULL DEFAULT 100,
    created_by INTEGER NULL,
    created_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_published BOOLEAN NOT NULL DEFAULT false,
    thumbnail VARCHAR(200) NULL,

    CONSTRAINT kuis_pkey PRIMARY KEY (kuis_id)
);