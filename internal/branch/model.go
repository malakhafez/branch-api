package branch

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BranchType string

const (
	TypeBranch         BranchType = "branch"
	TypeCentralKitchen BranchType = "central_kitchen"
)

// GeneratesRevenue: الفرع بس بيجيب إيرادات. المطبخ المركزي إيراده صفر (FR-BRANCH-003).
func (t BranchType) GeneratesRevenue() bool {
	return t == TypeBranch
}

type Branch struct {
	ID                uuid.UUID
	Name              string
	Type              BranchType
	Address           string
	OfficialStartTime string // "HH:MM"
	LateRate          decimal.Decimal
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}