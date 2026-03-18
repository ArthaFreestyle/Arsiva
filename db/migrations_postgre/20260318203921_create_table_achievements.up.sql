CREATE TABLE achievements (
    achievement_id SERIAL NOT NULL,
    nama VARCHAR(50) NOT NULL,
    deskripsi TEXT NULL,
    badge_icon VARCHAR(255) NOT NULL,
    xp_required INTEGER NOT NULL,
    tier tier_achievement_enum NOT NULL,

    CONSTRAINT achievements_pkey PRIMARY KEY (achievement_id)
);