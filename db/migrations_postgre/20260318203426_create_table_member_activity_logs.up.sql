CREATE TABLE member_activity_logs (
    log_id SERIAL NOT NULL,
    member_id INTEGER NULL,
    activity_type activity_type_enum NOT NULL,
    description TEXT NULL,
    timestamp TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT member_activity_logs_pkey PRIMARY KEY (log_id)
);