CREATE TABLE kategori_kuis (
    kategori_id SERIAL NOT NULL,
    nama_kategori VARCHAR(50) NOT NULL,
    created_at TIMESTAMP(0) NULL,
    created_by INTEGER NULL,
    deskripsi TEXT NULL,

    CONSTRAINT kategori_kuis_pkey PRIMARY KEY (kategori_id)
);