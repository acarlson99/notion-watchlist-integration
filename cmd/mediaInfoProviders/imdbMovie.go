package mediaInfoProviders

import "context"

type IMDBMediaInfo struct{}

func (*IMDBMediaInfo) GetMediaInfo(ctx context.Context, title string) {}
