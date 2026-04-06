-- Revert artikel
ALTER TABLE artikel ADD COLUMN thumbnail VARCHAR(255) NULL;
ALTER TABLE artikel DROP CONSTRAINT IF EXISTS fk_artikel_thumbnail;
ALTER TABLE artikel DROP COLUMN IF EXISTS thumbnail_asset_id;

-- Revert pertanyaan_kuis
ALTER TABLE pertanyaan_kuis ADD COLUMN image VARCHAR(255) NULL;
ALTER TABLE pertanyaan_kuis DROP CONSTRAINT IF EXISTS fk_pertanyaan_image;
ALTER TABLE pertanyaan_kuis DROP COLUMN IF EXISTS image_asset_id;

-- Revert kuis
ALTER TABLE kuis ADD COLUMN thumbnail VARCHAR(255) NULL;
ALTER TABLE kuis ADD COLUMN gambar VARCHAR(255) NULL;
ALTER TABLE kuis DROP CONSTRAINT IF EXISTS fk_kuis_gambar;
ALTER TABLE kuis DROP COLUMN IF EXISTS gambar_asset_id;
ALTER TABLE kuis DROP CONSTRAINT IF EXISTS fk_kuis_thumbnail;
ALTER TABLE kuis DROP COLUMN IF EXISTS thumbnail_asset_id;

-- Revert puzzles
ALTER TABLE puzzles ADD COLUMN thumbnail VARCHAR(255) NULL;
ALTER TABLE puzzles ADD COLUMN gambar VARCHAR(255) NULL;
ALTER TABLE puzzles DROP CONSTRAINT IF EXISTS fk_puzzle_thumbnail;
ALTER TABLE puzzles DROP COLUMN IF EXISTS thumbnail_asset_id;
ALTER TABLE puzzles DROP CONSTRAINT IF EXISTS fk_puzzle_gambar;
ALTER TABLE puzzles DROP COLUMN IF EXISTS gambar_asset_id;

-- Revert scene
ALTER TABLE scene ADD COLUMN scene_image VARCHAR(255) NULL;
ALTER TABLE scene DROP CONSTRAINT IF EXISTS fk_scene_image;
ALTER TABLE scene DROP COLUMN IF EXISTS scene_image_asset_id;

-- Revert cerita_interaktif
ALTER TABLE cerita_interaktif ADD COLUMN thumbnail VARCHAR(255) NULL;
ALTER TABLE cerita_interaktif DROP CONSTRAINT IF EXISTS fk_cerita_thumbnail;
ALTER TABLE cerita_interaktif DROP COLUMN IF EXISTS thumbnail_asset_id;
