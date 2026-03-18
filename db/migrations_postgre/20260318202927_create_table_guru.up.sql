CREATE TABLE guru (
    guru_id SERIAL NOT NULL,
    user_id INTEGER NULL,
    sekolah_id INTEGER NULL,
    nip VARCHAR(20) NULL,
    bidang_ajar VARCHAR(100) NULL,

    CONSTRAINT guru_pkey PRIMARY KEY (guru_id),
    CONSTRAINT guru_user_id_key UNIQUE (user_id),
    CONSTRAINT guru_nip_key UNIQUE (nip)
);