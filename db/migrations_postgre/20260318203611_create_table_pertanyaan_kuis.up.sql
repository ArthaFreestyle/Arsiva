CREATE TABLE pertanyaan_kuis (
    pertanyaan_id SERIAL NOT NULL,
    kuis_id INTEGER NULL,
    teks_pertanyaan TEXT NOT NULL,
    image VARCHAR(200) NULL,
    tipe tipe_pertanyaan_enum NOT NULL,
    poin INTEGER NOT NULL DEFAULT 10,
    urutan INTEGER NOT NULL,

    CONSTRAINT pertanyaan_kuis_pkey PRIMARY KEY (pertanyaan_id)
);