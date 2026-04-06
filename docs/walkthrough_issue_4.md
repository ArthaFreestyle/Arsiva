# Walkthrough - Centralizing Asset Management System

I have successfully implemented a centralized asset management system. This ensures all images and media are tracked in a dedicated `assets` table, allowing for better lifecycle management and automated storage optimization.

## Changes Made

### 1. Database & Infrastructure
- [NEW] **`assets` table**: Stores `asset_id`, `url`, `is_used` (boolean), `created_at`, and `deleted_at`.
- [MODIFY] **Content Tables**: Removed direct URL columns (e.g., `thumbnail`, `gambar`) and added `_asset_id` foreign keys to:
  - `artikel`
  - `cerita_interaktif`
  - `scene`
  - `puzzles`
  - `kuis`
  - `pertanyaan_kuis`

### 2. Asset Management Layer
- [NEW] [AssetRepository](file:///home/artha/Documents/Arsiva/internal/repository/asset_repository.go): Handles CRUD operations for assets, including identifying orphaned records.
- [NEW] [AssetUsecase](file:///home/artha/Documents/Arsiva/internal/usecase/asset_usecase.go): Contains business logic for marking assets as used and the cleanup process.
- [MODIFY] [UploadController](file:///home/artha/Documents/Arsiva/internal/delivery/http/upload_controller.go): Now automatically registers every uploaded file into the `assets` table and returns the `asset_id`.

### 3. Module Refactoring
All major modules updated to use the new system:
- **Articles**: `thumbnail_asset_id`
- **Interactive Stories**: `thumbnail_asset_id` (Story) and `scene_image_asset_id` (Scene)
- **Quizzes**: `thumbnail_asset_id`, `gambar_asset_id` (Quiz) and `image_asset_id` (Question)
- **Puzzles**: `thumbnail_asset_id`, `gambar_asset_id`

### 4. Background Cleanup Task
- [MODIFY] [internal/config/app.go](file:///home/artha/Documents/Arsiva/internal/config/app.go): Added a daily background goroutine that:
  - Identifies assets where `is_used = false` and `created_at` is older than 7 days.
  - Deletes the physical files from the `./uploads` directory.
  - Soft-deletes the records from the `assets` table.

### 5. Documentation & Seeders
- [MODIFY] [docs/openapi.yaml](file:///home/artha/Documents/Arsiva/docs/openapi.yaml): Updated request schemas to use `asset_id` instead of string URLs.
- [MODIFY] [Seeder](file:///home/artha/Documents/Arsiva/db/migrations_seed/20260405205737_create_seeder_kategori_dan_cerita.up.sql): Updated to insert assets first before linking them to stories.

## Verification Results

### Build Status
- **Backend Build**: Successfully ran `go build ./...` without errors.

### Automated Cleanup
- Verified that `AssetUsecase` correctly maps URLs to file paths and handles missing files gracefully during cleanup.

> [!IMPORTANT]
> **Action Required**: The Frontend must be updated to use `asset_id` in POST/PUT payloads. The API still returns full URLs in GET responses for backward compatibility, but creation/updates now require IDs.
