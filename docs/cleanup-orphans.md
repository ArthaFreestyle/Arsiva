# Perbaikan relasi dan cleanup file yatim (issue #56)

Migrasi `20261110000001_group_content_integrity` memasang trigger untuk Kuis,
Cerita, dan Puzzle. Penghapusan konten otomatis menghapus pasangan
`content_type` + `content_id` dari semua grup dalam transaksi yang sama.
Trigger kedua menolak penambahan/pemindahan relasi ke konten yang tidak ada
dan mengunci baris target untuk menangani insert/delete bersamaan.
Migrasi juga membersihkan relasi yatim yang sudah ada. Jalankan melalui
alur migrasi deployment biasa; file fisik tidak dihapus oleh migrasi.

## Cleanup sekali jalan di server Docker Compose

Image backend sekarang menyertakan `/app/cleanup-orphans`. Jalankan dari
direktori repository server setelah membangun image terbaru:

```sh
docker compose build api-server

# Lihat kandidat terlebih dahulu, tanpa perubahan database/file.
docker compose run --rm --no-deps --entrypoint /app/cleanup-orphans api-server \
  --config /config.json --uploads /app/uploads
```

Pastikan hasil menunjuk ke direktori uploads yang benar. Buat backup database
dan uploads sebelum penghapusan. Hentikan API, worker, dan penulis uploads lain
selama apply. Flag `--offline` menyatakan layanan tersebut sudah dihentikan;
alat tidak menghentikannya secara otomatis.

```sh
docker compose stop api-server
# Hentikan worker/upload writer eksternal juga jika ada.

docker compose run --rm --no-deps --entrypoint /app/cleanup-orphans api-server \
  --config /config.json --uploads /app/uploads --apply --offline

# Jalankan kembali setelah cleanup selesai, termasuk bila cleanup gagal.
docker compose up -d api-server
```

Perintah memakai mount config dan uploads yang sudah didefinisikan di Compose.
PostgreSQL harus tetap berjalan. Secara default hanya file DAN baris aset yang
berumur lebih dari 168 jam (7 hari) yang menjadi kandidat. Untuk mengubahnya,
tambahkan misalnya `--min-age 720h`. `--min-age 0s` dapat dipakai saat offline
jika memang ingin memeriksa seluruh umur file, termasuk unggahan terbaru.

## Tanpa Docker

```sh
go build -o cleanup-orphans ./cmd/cleanup-orphans
./cleanup-orphans --config /path/config.json --uploads /path/uploads
./cleanup-orphans --config /path/config.json --uploads /path/uploads --apply --offline
```

Alternatif koneksi: isi environment `DATABASE_URL` untuk menggantikan konfigurasi
database dari JSON (termasuk pengaturan TLS). Jangan masukkan kredensial ke kode.

## Perilaku dan batasan

- Default adalah dry-run. Output `ORPHAN` menunjukkan kandidat, bukan bukti file
  sudah dihapus. Ringkasan menunjukkan jumlah kandidat file/aset dan relasi yatim.
- Kandidat meliputi aset terdaftar (termasuk `is_used=true` yang sudah usang),
  aset soft-deleted yang file-nya tertinggal, dan file tanpa baris `assets`.
- Referensi foreign key ke `assets` ditemukan melalui katalog PostgreSQL;
  URL yang masih muncul di kolom teks/JSON seluruh tabel `public` juga dilindungi,
  termasuk artikel dan scene. Referensi berdasarkan URL melindungi semua aset
  yang memakai URL identik. `is_used` bukan penentu tunggal.
- Hanya URL lokal kanonis `/uploads/...` yang dihapus. URL eksternal, path tidak
  aman, dan file khusus tidak dihapus. Symlink tidak ditelusuri; path kandidat
  yang mengandung symlink menyebabkan error tanpa menghapus target symlink.
- Semua tabel `public` dikunci terhadap perubahan selama apply. Tetap wajib
  offline karena upload file mendahului pencatatan database. Role database perlu
  izin membaca tabel, mengambil lock, update `assets`, dan delete `group_contents`.
- File yang masih dipakai di luar database (misalnya HTML statis atau sistem
  lain) tidak bisa dideteksi. Simpan file tersebut di luar direktori uploads
  yang dikelola aplikasi sebelum menjalankan cleanup.
- Jika penghapusan file gagal, proses berhenti dan transaksi database rollback.
  File yang sudah terhapus tidak bisa di-rollback oleh PostgreSQL. Jalankan ulang
  setelah penyebab gagal diperbaiki; file yang sudah tidak ada dianggap selesai.
  Baris aset dipertahankan dengan `deleted_at` untuk jejak cleanup.
- Cleanup aman dijalankan ulang. Aset yatim yang sudah tidak memiliki file dapat
  tetap muncul pada laporan berikutnya karena metadata aset tidak dihapus permanen.
- Down migration menghapus trigger/index, tetapi tidak mengembalikan relasi yatim
  yang sudah dibersihkan. Pemulihan data memerlukan backup.

## Pengujian

```sh
go test ./internal/maintenance ./cmd/cleanup-orphans
# Role test perlu CREATEDB. Tes membuat/menghapus database terpisah secara otomatis.
TEST_DATABASE_URL='postgres://USER:PASSWORD@localhost:5432/postgres?sslmode=disable' \
  go test ./internal/maintenance -v -count=1
```

Pengujian integrasi dilewati jika `TEST_DATABASE_URL` tidak disediakan.
