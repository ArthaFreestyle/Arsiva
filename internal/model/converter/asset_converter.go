package converter

import "ArthaFreestyle/Arsiva/internal/model"

// toAsset is a helper function reusable across all converters to build an AssetResponse
func toAsset(id *int, url string) *model.AssetResponse {
	if id == nil {
		return nil
	}
	return &model.AssetResponse{
		AssetId: *id,
		Url:     url,
	}
}
