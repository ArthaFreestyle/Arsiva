CREATE TABLE artikel (
    artikel_id SERIAL NOT NULL,
    judul VARCHAR(100) NOT NULL,
    konten TEXT NOT NULL,
    kategori_id INTEGER NULL,
    created_by INTEGER NULL,
    created_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    thumbnail VARCHAR(255) NULL,

    CONSTRAINT artikel_pkey PRIMARY KEY (artikel_id)
);