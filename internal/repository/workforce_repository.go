package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	_ = uuid.New
	_ = fmt.Sprintf
	_ = strings.Join
	_ = time.Now
	_ = errors.New
)

var (
	ErrWFSkillNotFound          = errors.New("workforce skill not found")
	ErrWFSkillCodeConflict      = errors.New("workforce skill with this code already exists in workspace")
	ErrWFProfileNotFound        = errors.New("member skill profile not found")
	ErrWFCertificationNotFound  = errors.New("workforce certification not found")
	ErrWFRoleCompetencyNotFound = errors.New("role competency not found")
	ErrWFTrainingNotFound       = errors.New("training enrollment not found")
)

// WorkforceRepository defines data access methods for ITIL Workforce & Talent Management
type WorkforceRepository interface {
	// Skills
	CreateSkill(ctx context.Context, s *model.WorkforceSkill) error
	GetSkillByID(ctx context.Context, id string) (*model.WorkforceSkill, error)
	GetSkillByCode(ctx context.Context, workspaceID, code string) (*model.WorkforceSkill, error)
	UpdateSkill(ctx context.Context, s *model.WorkforceSkill) error
	DeleteSkill(ctx context.Context, id string) error
	ListSkills(ctx context.Context, filter model.WFSkillFilter) ([]model.WorkforceSkill, int, error)

	// Member Skill Profiles
	CreateProfile(ctx context.Context, p *model.MemberSkillProfile) error
	GetProfileByID(ctx context.Context, id string) (*model.MemberSkillProfile, error)
	GetProfileByUserAndSkill(ctx context.Context, workspaceID, userID, skillID string) (*model.MemberSkillProfile, error)
	UpdateProfile(ctx context.Context, p *model.MemberSkillProfile) error
	DeleteProfile(ctx context.Context, id string) error
	ListProfiles(ctx context.Context, filter model.WFProfileFilter) ([]model.MemberSkillProfile, int, error)

	// Certifications
	CreateCertification(ctx context.Context, c *model.WorkforceCertification) error
	GetCertificationByID(ctx context.Context, id string) (*model.WorkforceCertification, error)
	UpdateCertification(ctx context.Context, c *model.WorkforceCertification) error
	DeleteCertification(ctx context.Context, id string) error
	ListCertifications(ctx context.Context, filter model.WFCertFilter) ([]model.WorkforceCertification, int, error)

	// Role Competencies
	CreateRoleCompetency(ctx context.Context, rc *model.RoleCompetency) error
	GetRoleCompetencyByID(ctx context.Context, id string) (*model.RoleCompetency, error)
	DeleteRoleCompetency(ctx context.Context, id string) error
	ListRoleCompetencies(ctx context.Context, filter model.WFRoleCompetencyFilter) ([]model.RoleCompetency, int, error)

	// Trainings
	CreateTraining(ctx context.Context, t *model.TrainingEnrollment) error
	GetTrainingByID(ctx context.Context, id string) (*model.TrainingEnrollment, error)
	UpdateTraining(ctx context.Context, t *model.TrainingEnrollment) error
	DeleteTraining(ctx context.Context, id string) error
	ListTrainings(ctx context.Context, filter model.WFTrainingFilter) ([]model.TrainingEnrollment, int, error)

	// Readiness Analytics
	GetWorkforceAnalytics(ctx context.Context, workspaceID string) (*model.WorkforceAnalytics, error)
}

type sqlWorkforceRepository struct {
	db *sql.DB
}

// NewWorkforceRepository initializes a SQLite workforce repository
func NewWorkforceRepository(db *sql.DB) WorkforceRepository {
	return &sqlWorkforceRepository{db: db}
}

// CreateSkill inserts a new cataloged skill
func (r *sqlWorkforceRepository) CreateSkill(ctx context.Context, s *model.WorkforceSkill) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	if s.Category == "" {
		s.Category = model.WFSkillCategorySoftwareEng
	}

	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workforce_skills WHERE workspace_id = ? AND skill_code = ?", s.WorkspaceID, s.SkillCode).Scan(&count)
	if err == nil && count > 0 {
		return ErrWFSkillCodeConflict
	}

	query := `
		INSERT INTO workforce_skills (
			id, workspace_id, skill_code, name, category, description, is_critical, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		s.ID, s.WorkspaceID, s.SkillCode, s.Name, s.Category, s.Description, s.IsCritical, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func scanWFSkill(scanner interface{ Scan(dest ...interface{}) error }) (*model.WorkforceSkill, error) {
	var s model.WorkforceSkill
	var desc sql.NullString
	err := scanner.Scan(
		&s.ID, &s.WorkspaceID, &s.SkillCode, &s.Name, &s.Category, &desc, &s.IsCritical, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		s.Description = desc.String
	}
	return &s, nil
}

// GetSkillByID retrieves a skill by ID
func (r *sqlWorkforceRepository) GetSkillByID(ctx context.Context, id string) (*model.WorkforceSkill, error) {
	query := `
		SELECT id, workspace_id, skill_code, name, category, description, is_critical, created_at, updated_at
		FROM workforce_skills
		WHERE id = ?`

	s, err := scanWFSkill(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWFSkillNotFound
		}
		return nil, err
	}
	return s, nil
}

// GetSkillByCode retrieves a skill by skill_code within a workspace
func (r *sqlWorkforceRepository) GetSkillByCode(ctx context.Context, workspaceID, code string) (*model.WorkforceSkill, error) {
	query := `
		SELECT id, workspace_id, skill_code, name, category, description, is_critical, created_at, updated_at
		FROM workforce_skills
		WHERE workspace_id = ? AND skill_code = ?`

	s, err := scanWFSkill(r.db.QueryRowContext(ctx, query, workspaceID, code))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWFSkillNotFound
		}
		return nil, err
	}
	return s, nil
}

// UpdateSkill updates an existing cataloged skill
func (r *sqlWorkforceRepository) UpdateSkill(ctx context.Context, s *model.WorkforceSkill) error {
	s.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE workforce_skills
		SET name = ?, category = ?, description = ?, is_critical = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query, s.Name, s.Category, s.Description, s.IsCritical, s.UpdatedAt, s.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWFSkillNotFound
	}
	return nil
}

// DeleteSkill removes a skill from the catalog
func (r *sqlWorkforceRepository) DeleteSkill(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM workforce_skills WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWFSkillNotFound
	}
	return nil
}

// ListSkills queries skills based on filter criteria
func (r *sqlWorkforceRepository) ListSkills(ctx context.Context, filter model.WFSkillFilter) ([]model.WorkforceSkill, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, filter.Category)
	}
	if filter.IsCritical != nil {
		conditions = append(conditions, "is_critical = ?")
		args = append(args, *filter.IsCritical)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(name LIKE ? OR skill_code LIKE ? OR description LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s, s)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM workforce_skills %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, workspace_id, skill_code, name, category, description, is_critical, created_at, updated_at
		FROM workforce_skills
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var skills []model.WorkforceSkill
	for rows.Next() {
		s, err := scanWFSkill(rows)
		if err != nil {
			return nil, 0, err
		}
		skills = append(skills, *s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return skills, total, nil
}

// Profiles stubs
func (r *sqlWorkforceRepository) CreateProfile(ctx context.Context, p *model.MemberSkillProfile) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) GetProfileByID(ctx context.Context, id string) (*model.MemberSkillProfile, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlWorkforceRepository) GetProfileByUserAndSkill(ctx context.Context, workspaceID, userID, skillID string) (*model.MemberSkillProfile, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlWorkforceRepository) UpdateProfile(ctx context.Context, p *model.MemberSkillProfile) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) DeleteProfile(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) ListProfiles(ctx context.Context, filter model.WFProfileFilter) ([]model.MemberSkillProfile, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Certifications stubs
func (r *sqlWorkforceRepository) CreateCertification(ctx context.Context, c *model.WorkforceCertification) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) GetCertificationByID(ctx context.Context, id string) (*model.WorkforceCertification, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlWorkforceRepository) UpdateCertification(ctx context.Context, c *model.WorkforceCertification) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) DeleteCertification(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) ListCertifications(ctx context.Context, filter model.WFCertFilter) ([]model.WorkforceCertification, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Role Competencies stubs
func (r *sqlWorkforceRepository) CreateRoleCompetency(ctx context.Context, rc *model.RoleCompetency) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) GetRoleCompetencyByID(ctx context.Context, id string) (*model.RoleCompetency, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlWorkforceRepository) DeleteRoleCompetency(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) ListRoleCompetencies(ctx context.Context, filter model.WFRoleCompetencyFilter) ([]model.RoleCompetency, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Trainings stubs
func (r *sqlWorkforceRepository) CreateTraining(ctx context.Context, t *model.TrainingEnrollment) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) GetTrainingByID(ctx context.Context, id string) (*model.TrainingEnrollment, error) {
	return nil, errors.New("not implemented")
}
func (r *sqlWorkforceRepository) UpdateTraining(ctx context.Context, t *model.TrainingEnrollment) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) DeleteTraining(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (r *sqlWorkforceRepository) ListTrainings(ctx context.Context, filter model.WFTrainingFilter) ([]model.TrainingEnrollment, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Analytics stub
func (r *sqlWorkforceRepository) GetWorkforceAnalytics(ctx context.Context, workspaceID string) (*model.WorkforceAnalytics, error) {
	return nil, errors.New("not implemented")
}
