CREATE TYPE daily_task_type_enum AS ENUM (
    'complete_quiz',
    'complete_story',
    'solve_puzzle',
    'earn_xp',
    'daily_complete_bonus'
);

CREATE TABLE daily_tasks (
    daily_task_id SERIAL      NOT NULL,
    member_id     INTEGER     NOT NULL,
    task_date     DATE        NOT NULL,
    task_type     daily_task_type_enum NOT NULL,
    target_count  INTEGER     NOT NULL DEFAULT 1,
    current_count INTEGER     NOT NULL DEFAULT 0,
    xp_reward     INTEGER     NOT NULL DEFAULT 0,
    completed_at  TIMESTAMP(0) NULL,
    created_at    TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT daily_tasks_pkey   PRIMARY KEY (daily_task_id),
    CONSTRAINT daily_tasks_member_fk FOREIGN KEY (member_id)
        REFERENCES members (member_id) ON DELETE CASCADE,
    CONSTRAINT daily_tasks_unique UNIQUE (member_id, task_date, task_type)
);

CREATE INDEX idx_daily_tasks_member_date ON daily_tasks (member_id, task_date);
