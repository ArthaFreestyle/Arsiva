-- AddForeignKey for guru
ALTER TABLE guru ADD CONSTRAINT guru_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE guru ADD CONSTRAINT guru_sekolah_id_fkey FOREIGN KEY (sekolah_id) REFERENCES sekolah(sekolah_id) ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey for members
ALTER TABLE members ADD CONSTRAINT members_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE members ADD CONSTRAINT members_sekolah_id_fkey FOREIGN KEY (sekolah_id) REFERENCES sekolah(sekolah_id) ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey for groups
ALTER TABLE groups ADD CONSTRAINT groups_created_by_fkey FOREIGN KEY (created_by) REFERENCES guru(guru_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE groups ADD CONSTRAINT fk_groups_thumbnail FOREIGN KEY (group_thumbnail_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey for group_members
ALTER TABLE group_members ADD CONSTRAINT group_members_group_id_fkey FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE group_members ADD CONSTRAINT group_members_member_id_fkey FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey for group_contents
ALTER TABLE group_contents ADD CONSTRAINT group_contents_group_id_fkey FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey for member_progress
ALTER TABLE member_progress ADD CONSTRAINT member_progress_member_id_fkey FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE member_progress ADD CONSTRAINT member_progress_group_id_fkey FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey for member_social_links
ALTER TABLE member_social_links ADD CONSTRAINT member_social_links_member_id_fkey FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey for member_activity_logs
ALTER TABLE member_activity_logs ADD CONSTRAINT member_activity_logs_member_id_fkey FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey for kategori_kuis
ALTER TABLE kategori_kuis ADD CONSTRAINT kategori_kuis_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey for kuis
ALTER TABLE kuis ADD CONSTRAINT kuis_kategori_id_fkey FOREIGN KEY (kategori_id) REFERENCES kategori_kuis(kategori_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE kuis ADD CONSTRAINT kuis_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE kuis ADD CONSTRAINT fk_kuis_thumbnail FOREIGN KEY (thumbnail_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE kuis ADD CONSTRAINT fk_kuis_gambar FOREIGN KEY (gambar_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;

-- AddForeignKey for pertanyaan_kuis
ALTER TABLE pertanyaan_kuis ADD CONSTRAINT pertanyaan_kuis_kuis_id_fkey FOREIGN KEY (kuis_id) REFERENCES kuis(kuis_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE pertanyaan_kuis ADD CONSTRAINT fk_pertanyaan_image FOREIGN KEY (image_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;

-- AddForeignKey for pilihan_kuis
ALTER TABLE pilihan_kuis ADD CONSTRAINT pilihan_kuis_pertanyaan_id_fkey FOREIGN KEY (pertanyaan_id) REFERENCES pertanyaan_kuis(pertanyaan_id) ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey for cerita_interaktif
ALTER TABLE cerita_interaktif ADD CONSTRAINT cerita_interaktif_kategori_id_fkey FOREIGN KEY (kategori_id) REFERENCES kategori_cerita(kategori_id) ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE cerita_interaktif ADD CONSTRAINT cerita_interaktif_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE cerita_interaktif ADD CONSTRAINT fk_cerita_thumbnail FOREIGN KEY (thumbnail_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;

-- AddForeignKey for scene
ALTER TABLE scene ADD CONSTRAINT scene_cerita_id_fkey FOREIGN KEY (cerita_id) REFERENCES cerita_interaktif(cerita_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE scene ADD CONSTRAINT fk_scene_image FOREIGN KEY (scene_image_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;

-- AddForeignKey for puzzles
ALTER TABLE puzzles ADD CONSTRAINT puzzles_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE puzzles ADD CONSTRAINT fk_puzzle_gambar FOREIGN KEY (gambar_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE puzzles ADD CONSTRAINT fk_puzzle_thumbnail FOREIGN KEY (thumbnail_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;

-- AddForeignKey for artikel
ALTER TABLE artikel ADD CONSTRAINT artikel_kategori_id_fkey FOREIGN KEY (kategori_id) REFERENCES kategori_artikel(kategori_artikel_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE artikel ADD CONSTRAINT artikel_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE artikel ADD CONSTRAINT fk_artikel_thumbnail FOREIGN KEY (thumbnail_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;

-- AddForeignKey for member_achievements
ALTER TABLE member_achievements ADD CONSTRAINT member_achievements_member_id_fkey FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE member_achievements ADD CONSTRAINT member_achievements_achievement_id_fkey FOREIGN KEY (achievement_id) REFERENCES achievements(achievement_id) ON DELETE CASCADE ON UPDATE CASCADE;