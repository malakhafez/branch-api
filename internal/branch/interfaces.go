package branch

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository: اللي الـ Service محتاجه من قاعدة البيانات.
type Repository interface {
	Create(ctx context.Context, b Branch) error
	GetByID(ctx context.Context, id uuid.UUID) (Branch, error)
	List(ctx context.Context) ([]Branch, error)
	Update(ctx context.Context, b Branch) error
}

// Clock: مصدر الوقت. بنحقنه عشان الاختبارات تثبّت الوقت.
type Clock interface {
	Now() time.Time
}