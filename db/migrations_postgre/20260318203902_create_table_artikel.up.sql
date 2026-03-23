CREATE TABLE artikel (
    artikel_id SERIAL NOT NULL,
    slug VARCHAR(255) NOT NULL,
    judul VARCHAR(100) NOT NULL,
    konten JSONB NULL,
    kategori_id INTEGER NULL,
    status status_enum NOT NULL,
    excerpt VARCHAR(255) NULL,
    created_by INTEGER NULL,
    created_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    thumbnail VARCHAR(255) NULL,

    CONSTRAINT artikel_pkey PRIMARY KEY (artikel_id),
    CONSTRAINT articles_slug_key UNIQUE (slug)
);