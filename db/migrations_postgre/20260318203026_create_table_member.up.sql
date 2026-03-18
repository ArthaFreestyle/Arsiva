CREATE TABLE members (
    member_id SERIAL NOT NULL,
    user_id INTEGER NULL,
    sekolah_id INTEGER NULL,
    nis VARCHAR(20) NULL,
    total_xp INTEGER NOT NULL DEFAULT 0,
    level INTEGER NOT NULL DEFAULT 0,
    foto_profil VARCHAR(255) NULL,
    bio TEXT NULL,
    tanggal_lahir DATE NULL,
    jenis_kelamin jenis_kelamin_enum NULL,
    minat TEXT NULL,
    last_active TIMESTAMP(0) NULL,

    CONSTRAINT members_pkey PRIMARY KEY (member_id),
    CONSTRAINT members_user_id_key UNIQUE (user_id)
);