package branch

import (
	"time"

	"github.com/shopspring/decimal"
)

// CreateBranchRequest: اللي الـ CEO بيبعته لما يعمل فرع جديد.
// late_rate إجباري، فهو مؤشر عشان نعرف لو ما اتبعتش.
type CreateBranchRequest struct {
	Name              string           `json:"name"`
	Type              string           `json:"type"`
	Address           string           `json:"address"`
	OfficialStartTime string           `json:"official_start_time"` // "HH:MM"
	LateRate          *decimal.Decimal `json:"late_rate"`
}

// UpdateBranchRequest: كل الحقول اختيارية. اللي ما اتبعتش ما بيتغيّرش.
type UpdateBranchRequest struct {
	Name              *string          `json:"name"`
	Address           *string          `json:"address"`
	OfficialStartTime *string          `json:"official_start_time"`
	LateRate          *decimal.Decimal `json:"late_rate"`
	IsActive          *bool            `json:"is_active"`
}

// BranchResponse: اللي الـ API بيرجّعه. late_rate نص بخانتين عشريتين.
type BranchResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	Address           string `json:"address"`
	OfficialStartTime string `json:"official_start_time"`
	LateRate          string `json:"late_rate"`
	IsActive          bool   `json:"is_active"`
	GeneratesRevenue  bool   `json:"generates_revenue"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// ToResponse بتحوّل Branch لشكل الـ JSON، والوقت بصيغة ISO 8601 UTC.
func ToResponse(b Branch) BranchResponse {
	return BranchResponse{
		ID:                b.ID.String(),
		Name:              b.Name,
		Type:              string(b.Type),
		Address:           b.Address,
		OfficialStartTime: b.OfficialStartTime,
		LateRate:          b.LateRate.StringFixed(2),
		IsActive:          b.IsActive,
		GeneratesRevenue:  b.Type.GeneratesRevenue(),
		CreatedAt:         b.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         b.UpdatedAt.UTC().Format(time.RFC3339),
	}
}