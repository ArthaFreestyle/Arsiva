CREATE TABLE group_members (
    group_id VARCHAR(191) NOT NULL,
    member_id INTEGER NOT NULL,
    tanggal_bergabung TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT group_members_pkey PRIMARY KEY (group_id, member_id)
);