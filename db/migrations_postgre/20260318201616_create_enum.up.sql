-- Create Enum Types First
CREATE TYPE role_enum AS ENUM ('super_admin', 'guru', 'member');
CREATE TYPE jenis_kelamin_enum AS ENUM ('L', 'P', 'Lainnya');
CREATE TYPE content_type_enum AS ENUM ('kuis', 'cerita', 'puzzle');
CREATE TYPE platform_enum AS ENUM ('Instagram', 'Twitter', 'TikTok', 'Facebook', 'YouTube', 'Lainnya');
CREATE TYPE activity_type_enum AS ENUM ('login', 'update_profile', 'complete_quiz', 'read_story', 'solve_puzzle');
CREATE TYPE tipe_pertanyaan_enum AS ENUM ('pilihan_ganda', 'benar_salah');
CREATE TYPE kategori_puzzle_enum AS ENUM ('Tempat_Wisata', 'Tokoh_Sejarah', 'Peta', 'Budaya', 'Lainnya');
CREATE TYPE tier_achievement_enum AS ENUM ('bronze', 'silver', 'gold', 'platinum');
CREATE TYPE status_enum AS ENUM ('draft', 'published');