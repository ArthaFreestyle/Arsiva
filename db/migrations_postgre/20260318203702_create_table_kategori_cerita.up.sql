CREATE TABLE kategori_cerita (
    kategori_id SERIAL NOT NULL,
    nama_kategori VARCHAR(100) NOT NULL,

    CONSTRAINT kategori_cerita_pkey PRIMARY KEY (kategori_id)
);