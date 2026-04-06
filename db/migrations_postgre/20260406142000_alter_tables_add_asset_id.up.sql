-- Add thumbnail_asset_id to cerita_interaktif
ALTER TABLE cerita_interaktif ADD COLUMN IF NOT EXISTS thumbnail_asset_id INT NULL;
ALTER TABLE cerita_interaktif DROP CONSTRAINT IF EXISTS fk_cerita_thumbnail;
ALTER TABLE cerita_interaktif ADD CONSTRAINT fk_cerita_thumbnail FOREIGN KEY (thumbnail_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE cerita_interaktif DROP COLUMN IF EXISTS thumbnail;

-- Add scene_image_asset_id to scene
ALTER TABLE scene ADD COLUMN IF NOT EXISTS scene_image_asset_id INT NULL;
ALTER TABLE scene DROP CONSTRAINT IF EXISTS fk_scene_image;
ALTER TABLE scene ADD CONSTRAINT fk_scene_image FOREIGN KEY (scene_image_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE scene DROP COLUMN IF EXISTS scene_image;

-- Add gambar_asset_id, thumbnail_asset_id to puzzles
ALTER TABLE puzzles ADD COLUMN IF NOT EXISTS gambar_asset_id INT NULL;
ALTER TABLE puzzles DROP CONSTRAINT IF EXISTS fk_puzzle_gambar;
ALTER TABLE puzzles ADD CONSTRAINT fk_puzzle_gambar FOREIGN KEY (gambar_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE puzzles ADD COLUMN IF NOT EXISTS thumbnail_asset_id INT NULL;
ALTER TABLE puzzles DROP CONSTRAINT IF EXISTS fk_puzzle_thumbnail;
ALTER TABLE puzzles ADD CONSTRAINT fk_puzzle_thumbnail FOREIGN KEY (thumbnail_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE puzzles DROP COLUMN IF EXISTS gambar;
ALTER TABLE puzzles DROP COLUMN IF EXISTS thumbnail;

-- Add thumbnail_asset_id, gambar_asset_id to kuis
ALTER TABLE kuis ADD COLUMN IF NOT EXISTS thumbnail_asset_id INT NULL;
ALTER TABLE kuis DROP CONSTRAINT IF EXISTS fk_kuis_thumbnail;
ALTER TABLE kuis ADD CONSTRAINT fk_kuis_thumbnail FOREIGN KEY (thumbnail_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE kuis ADD COLUMN IF NOT EXISTS gambar_asset_id INT NULL;
ALTER TABLE kuis DROP CONSTRAINT IF EXISTS fk_kuis_gambar;
ALTER TABLE kuis ADD CONSTRAINT fk_kuis_gambar FOREIGN KEY (gambar_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE kuis DROP COLUMN IF EXISTS thumbnail;
ALTER TABLE kuis DROP COLUMN IF EXISTS gambar;

-- Add image_asset_id to pertanyaan_kuis
ALTER TABLE pertanyaan_kuis ADD COLUMN IF NOT EXISTS image_asset_id INT NULL;
ALTER TABLE pertanyaan_kuis DROP CONSTRAINT IF EXISTS fk_pertanyaan_image;
ALTER TABLE pertanyaan_kuis ADD CONSTRAINT fk_pertanyaan_image FOREIGN KEY (image_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE pertanyaan_kuis DROP COLUMN IF EXISTS image;

-- Add thumbnail_asset_id to artikel
ALTER TABLE artikel ADD COLUMN IF NOT EXISTS thumbnail_asset_id INT NULL;
ALTER TABLE artikel DROP CONSTRAINT IF EXISTS fk_artikel_thumbnail;
ALTER TABLE artikel ADD CONSTRAINT fk_artikel_thumbnail FOREIGN KEY (thumbnail_asset_id) REFERENCES assets(asset_id) ON DELETE SET NULL;
ALTER TABLE artikel DROP COLUMN IF EXISTS thumbnail;
