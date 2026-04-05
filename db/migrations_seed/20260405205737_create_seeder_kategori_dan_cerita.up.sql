INSERT INTO kategori_cerita (nama_kategori)
VALUES 
  ('Petualangan'),
  ('Misteri'),
  ('Edukasi'),
  ('Fantasi'),
  ('Sci-Fi');

INSERT INTO cerita_interaktif (judul, thumbnail, deskripsi, kategori_id, xp_reward, created_by, created_at, is_published)
VALUES
  ('Misteri Rumah Tua', 'https://dummyimage.com/800x600/000/fff&text=Misteri', 'Temukan rahasia yang tersembunyi di dalam rumah tua peninggalan kakek.', 2, 200, 1, NOW(), true),
  ('Petualangan di Hutan Ajaib', 'https://dummyimage.com/800x600/000/fff&text=Petualangan', 'Bantu sang pahlawan menyelesaikan misinya di hutan yang penuh sihir.', 1, 150, 2, NOW(), true),
  ('Belajar Tata Surya', NULL, 'Mengenal planet-planet di tata surya kita melalui petualangan luar angkasa yang seru.', 3, 100, 3, NOW(), false);
