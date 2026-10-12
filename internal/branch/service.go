package branch

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// أقصى طول للاسم (نفس VARCHAR(100) في الجدول).
const maxNameLength = 100

// أقصى خصم عن الساعة بالجنيه. قاعدة عمل مؤقتة لحد ما التيم يحدد الرقم النهائي.
// (حد الداتا بيز أكبر من كده بكتير: NUMERIC(10,2))
var maxLateRate = decimal.RequireFromString("10000")

// validateName بتشيل المسافات الزيادة وبتتأكد إن الاسم مش فاضي ومش أطول من الحد.
func validateName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", ErrNameRequired
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		return "", ErrNameTooLong
	}
	return name, nil
}

// validateLateRate بتتأكد إن القيمة مش سالبة، ومش أكبر من الحد، وبحد أقصى خانتين عشريتين.
func validateLateRate(r decimal.Decimal) error {
	if r.IsNegative() {
		return ErrInvalidLateRate
	}
	// حماية: أس ضخم (زي 1e999999999) ممكن يستهلك ذاكرة في المقارنة والتقريب،
	// فنرفضه قبل أي حساب.
	if r.Exponent() > 20 {
		return ErrLateRateTooLarge
	}
	if r.Exponent() < -20 {
		return ErrLateRatePrecision
	}
	if r.GreaterThan(maxLateRate) {
		return ErrLateRateTooLarge
	}
	if !r.Equal(r.Round(2)) {
		return ErrLateRatePrecision
	}
	return nil
}

// SystemClock: الوقت الحقيقي. في الاختبارات بنستبدله بوقت ثابت.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type Service struct {
	repo  Repository
	clock Clock
}

func NewService(repo Repository, clock Clock) *Service {
	return &Service{repo: repo, clock: clock}
}

func (s *Service) Create(ctx context.Context, req CreateBranchRequest) (Branch, error) {
	name, err := validateName(req.Name)
	if err != nil {
		return Branch{}, err
	}

	bt := BranchType(req.Type)
	if bt != TypeBranch && bt != TypeCentralKitchen {
		return Branch{}, ErrInvalidType
	}

	startTime, err := normalizeTime(req.OfficialStartTime)
	if err != nil {
		return Branch{}, err
	}

	if req.LateRate == nil {
		return Branch{}, ErrLateRateRequired
	}
	if err := validateLateRate(*req.LateRate); err != nil {
		return Branch{}, err
	}

	now := s.clock.Now()
	b := Branch{
		ID:                uuid.New(),
		Name:              name,
		Type:              bt,
		Address:           strings.TrimSpace(req.Address),
		OfficialStartTime: startTime,
		LateRate:          *req.LateRate,
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return Branch{}, err
	}
	return b, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Branch, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Branch, error) {
	return s.repo.List(ctx)
}

// Update بيغيّر الحقول المبعوتة بس. لو مفيش حاجة اتغيّرت فعلًا، بيرجّع الفرع
// زي ما هو من غير ما يكتب في الداتا بيز ومن غير ما يلمس updated_at.
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateBranchRequest) (Branch, error) {
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Branch{}, err
	}

	changed := false

	if req.Name != nil {
		name, err := validateName(*req.Name)
		if err != nil {
			return Branch{}, err
		}
		if name != b.Name {
			b.Name = name
			changed = true
		}
	}
	if req.Address != nil {
		address := strings.TrimSpace(*req.Address)
		if address != b.Address {
			b.Address = address
			changed = true
		}
	}
	if req.OfficialStartTime != nil {
		t, err := normalizeTime(*req.OfficialStartTime)
		if err != nil {
			return Branch{}, err
		}
		if t != b.OfficialStartTime {
			b.OfficialStartTime = t
			changed = true
		}
	}
	if req.LateRate != nil {
		if err := validateLateRate(*req.LateRate); err != nil {
			return Branch{}, err
		}
		if !req.LateRate.Equal(b.LateRate) {
			b.LateRate = *req.LateRate
			changed = true
		}
	}
	if req.IsActive != nil {
		if *req.IsActive != b.IsActive {
			b.IsActive = *req.IsActive
			changed = true
		}
	}

	if !changed {
		return b, nil
	}

	b.UpdatedAt = s.clock.Now()
	if err := s.repo.Update(ctx, b); err != nil {
		return Branch{}, err
	}
	return b, nil
}

// normalizeTime بتتأكد إن الوقت بصيغة HH:MM وترجّعه موحّد (9:00 تبقى 09:00).
func normalizeTime(s string) (string, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return "", ErrInvalidTime
	}
	return t.Format("15:04"), nil
}
