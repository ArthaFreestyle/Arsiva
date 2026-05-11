ALTER TABLE member_social_links
    ADD CONSTRAINT member_social_links_member_id_fkey
    FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_member_social_links_member_id
    ON member_social_links(member_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_member_social_links_member_platform
    ON member_social_links(member_id, platform);
