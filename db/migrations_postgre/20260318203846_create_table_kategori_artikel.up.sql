CREATE TABLE kategori_artikel (
    kategori_artikel_id SERIAL NOT NULL,
    nama_kategori VARCHAR(50) NOT NULL,

    CONSTRAINT kategori_artikel_pkey PRIMARY KEY (kategori_artikel_id)
);