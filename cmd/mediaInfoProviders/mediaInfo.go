package mediaInfoProviders

import "context"

type MediaInfo struct {
	Authors  []string // alternatively Director
	Link     string
	Summary  string
	Category []string
	Rating   float64 // min 1.0, max 5.0
	Image    string  // link to thumbnail image

	PageCount int64
}

type MediaInfoProvider interface {
	GetMediaInfo(ctx context.Context, title string) (*MediaInfo, error)
}
