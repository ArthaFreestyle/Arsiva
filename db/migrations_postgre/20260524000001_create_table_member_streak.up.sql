CREATE TABLE member_streaks (
    member_id         INTEGER  NOT NULL,
    current_streak    INTEGER  NOT NULL DEFAULT 0,
    longest_streak    INTEGER  NOT NULL DEFAULT 0,
    last_active_date  DATE     NULL,
    freezes_available INTEGER  NOT NULL DEFAULT 0,
    updated_at        TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT member_streaks_pkey PRIMARY KEY (member_id),
    CONSTRAINT member_streaks_member_fk FOREIGN KEY (member_id)
        REFERENCES members (member_id) ON DELETE CASCADE
);
