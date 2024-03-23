package mediaInfoProviders

import (
	"context"
	"fmt"

	"google.golang.org/api/books/v1"
)

type GoogleBooksMediaInfo struct{}

func (*GoogleBooksMediaInfo) findVolumes(title string) (*books.Volumes, error) {
	ctx := context.Background()

	bservice, err := books.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("books.NewService: %v", err)
	}
	service := books.NewVolumesService(bservice)

	volume, err := service.List(fmt.Sprintf("intitle:%s", title)).Do()
	if err != nil {
		return nil, fmt.Errorf("Volumes.List: %v", err)
	}

	return volume, nil
}

func (gminf *GoogleBooksMediaInfo) GetMediaInfo(ctx context.Context, title string) (*MediaInfo, error) {
	volumes, err := gminf.findVolumes(title)
	if err != nil {
		return nil, err
	}
	if len(volumes.Items) < 1 {
		return nil, fmt.Errorf("no volumes found for %s", title)
	}
	volume := volumes.Items[0].VolumeInfo
	mediaInfo := &MediaInfo{
		Authors:   volume.Authors,
		Category:  volume.Categories,
		Rating:    volume.AverageRating,
		Summary:   volume.Description,
		Link:      volume.CanonicalVolumeLink,
		Image:     volume.ImageLinks.Thumbnail,
		PageCount: volume.PageCount,
	}
	return mediaInfo, nil
}
