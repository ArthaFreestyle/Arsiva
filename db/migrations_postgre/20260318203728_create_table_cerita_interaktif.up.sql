CREATE TABLE cerita_interaktif (
    cerita_id SERIAL NOT NULL,
    judul VARCHAR(100) NOT NULL,
    thumbnail_asset_id INTEGER NULL,
    deskripsi TEXT NULL,
    kategori_id INTEGER NOT NULL DEFAULT 1,
    xp_reward INTEGER NOT NULL DEFAULT 150,
    created_by INTEGER NULL,
    created_at TIMESTAMP(0) NULL,
    is_published BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT cerita_interaktif_pkey PRIMARY KEY (cerita_id)
);