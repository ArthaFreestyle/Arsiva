CREATE TABLE pilihan_kuis (
    jawaban_id SERIAL NOT NULL,
    pertanyaan_id INTEGER NULL,
    teks_jawaban TEXT NOT NULL,
    score INTEGER NOT NULL DEFAULT 0,

    CONSTRAINT pilihan_kuis_pkey PRIMARY KEY (jawaban_id)
);