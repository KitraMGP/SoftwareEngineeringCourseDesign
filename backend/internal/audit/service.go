package audit

import "context"

type Event struct {
	Action       string
	Resource     string
	ResourceID   string
	Result       string
	TargetUserID string
}

type Service interface {
	Log(ctx context.Context, event Event) error
}
