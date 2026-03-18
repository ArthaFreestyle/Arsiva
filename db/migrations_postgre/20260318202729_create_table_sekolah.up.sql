CREATE TABLE sekolah (
    sekolah_id SERIAL NOT NULL,
    nama_sekolah TEXT NULL,
    alamat_sekolah TEXT NULL,

    CONSTRAINT sekolah_pkey PRIMARY KEY (sekolah_id)
);