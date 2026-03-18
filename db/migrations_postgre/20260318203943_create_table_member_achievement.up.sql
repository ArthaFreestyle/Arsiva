CREATE TABLE member_achievements (
    member_id INTEGER NOT NULL,
    achievement_id INTEGER NOT NULL,
    unlocked_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT member_achievements_pkey PRIMARY KEY (member_id, achievement_id)
);