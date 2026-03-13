package quota

import "context"

type Service interface {
	CheckChatAllowed(ctx context.Context, userID string) error
	CheckUploadAllowed(ctx context.Context, userID string, fileSizeBytes int64) error
}
