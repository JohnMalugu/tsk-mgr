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

// CreateProfile inserts a new member skill profile
func (r *sqlWorkforceRepository) CreateProfile(ctx context.Context, p *model.MemberSkillProfile) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if p.LastAssessedAt.IsZero() {
		p.LastAssessedAt = now
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	if p.ProficiencyLevel == "" {
		p.ProficiencyLevel = model.WFProficiencyNovice
	}

	query := `
		INSERT INTO workforce_skill_profiles (
			id, workspace_id, user_id, skill_id, proficiency_level,
			years_experience, verified_by, verified_at, last_assessed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.WorkspaceID, p.UserID, p.SkillID, p.ProficiencyLevel,
		p.YearsExperience, p.VerifiedBy, p.VerifiedAt, p.LastAssessedAt, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func scanWFProfile(scanner interface{ Scan(dest ...interface{}) error }) (*model.MemberSkillProfile, error) {
	var p model.MemberSkillProfile
	var verifiedBy sql.NullString
	var verifiedAt sql.NullTime
	err := scanner.Scan(
		&p.ID, &p.WorkspaceID, &p.UserID, &p.SkillID, &p.ProficiencyLevel,
		&p.YearsExperience, &verifiedBy, &verifiedAt, &p.LastAssessedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if verifiedBy.Valid {
		p.VerifiedBy = verifiedBy.String
	}
	if verifiedAt.Valid {
		p.VerifiedAt = &verifiedAt.Time
	}
	return &p, nil
}

// GetProfileByID retrieves a member skill profile by ID
func (r *sqlWorkforceRepository) GetProfileByID(ctx context.Context, id string) (*model.MemberSkillProfile, error) {
	query := `
		SELECT id, workspace_id, user_id, skill_id, proficiency_level,
		       years_experience, verified_by, verified_at, last_assessed_at, created_at, updated_at
		FROM workforce_skill_profiles
		WHERE id = ?`

	p, err := scanWFProfile(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWFProfileNotFound
		}
		return nil, err
	}
	return p, nil
}

// GetProfileByUserAndSkill retrieves a member's profile for a specific skill
func (r *sqlWorkforceRepository) GetProfileByUserAndSkill(ctx context.Context, workspaceID, userID, skillID string) (*model.MemberSkillProfile, error) {
	query := `
		SELECT id, workspace_id, user_id, skill_id, proficiency_level,
		       years_experience, verified_by, verified_at, last_assessed_at, created_at, updated_at
		FROM workforce_skill_profiles
		WHERE workspace_id = ? AND user_id = ? AND skill_id = ?`

	p, err := scanWFProfile(r.db.QueryRowContext(ctx, query, workspaceID, userID, skillID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWFProfileNotFound
		}
		return nil, err
	}
	return p, nil
}

// UpdateProfile updates proficiency, verification, and assessment metadata on a profile
func (r *sqlWorkforceRepository) UpdateProfile(ctx context.Context, p *model.MemberSkillProfile) error {
	p.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE workforce_skill_profiles
		SET proficiency_level = ?, years_experience = ?, verified_by = ?,
		    verified_at = ?, last_assessed_at = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		p.ProficiencyLevel, p.YearsExperience, p.VerifiedBy,
		p.VerifiedAt, p.LastAssessedAt, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWFProfileNotFound
	}
	return nil
}

// DeleteProfile removes a member skill profile
func (r *sqlWorkforceRepository) DeleteProfile(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM workforce_skill_profiles WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWFProfileNotFound
	}
	return nil
}

// ListProfiles queries member skill profiles based on filter criteria
func (r *sqlWorkforceRepository) ListProfiles(ctx context.Context, filter model.WFProfileFilter) ([]model.MemberSkillProfile, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.SkillID != "" {
		conditions = append(conditions, "skill_id = ?")
		args = append(args, filter.SkillID)
	}
	if filter.ProficiencyLevel != "" {
		conditions = append(conditions, "proficiency_level = ?")
		args = append(args, filter.ProficiencyLevel)
	}
	if filter.VerifiedOnly {
		conditions = append(conditions, "verified_at IS NOT NULL")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM workforce_skill_profiles %s", whereClause), args...).Scan(&total)
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
		SELECT id, workspace_id, user_id, skill_id, proficiency_level,
		       years_experience, verified_by, verified_at, last_assessed_at, created_at, updated_at
		FROM workforce_skill_profiles
		%s
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var profiles []model.MemberSkillProfile
	for rows.Next() {
		p, err := scanWFProfile(rows)
		if err != nil {
			return nil, 0, err
		}
		profiles = append(profiles, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

// CreateCertification inserts a new workforce professional certification
func (r *sqlWorkforceRepository) CreateCertification(ctx context.Context, c *model.WorkforceCertification) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if c.IssuedAt.IsZero() {
		c.IssuedAt = now
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	if c.Status == "" {
		c.Status = model.WFCertStatusActive
	}

	query := `
		INSERT INTO workforce_certifications (
			id, workspace_id, user_id, cert_code, name, issuer,
			credential_id, issued_at, expires_at, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.WorkspaceID, c.UserID, c.CertCode, c.Name, c.Issuer,
		c.CredentialID, c.IssuedAt, c.ExpiresAt, c.Status, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func scanWFCert(scanner interface{ Scan(dest ...interface{}) error }) (*model.WorkforceCertification, error) {
	var c model.WorkforceCertification
	var credID sql.NullString
	var expiresAt sql.NullTime
	err := scanner.Scan(
		&c.ID, &c.WorkspaceID, &c.UserID, &c.CertCode, &c.Name, &c.Issuer,
		&credID, &c.IssuedAt, &expiresAt, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if credID.Valid {
		c.CredentialID = credID.String
	}
	if expiresAt.Valid {
		c.ExpiresAt = &expiresAt.Time
	}
	return &c, nil
}

// GetCertificationByID retrieves a certification by ID
func (r *sqlWorkforceRepository) GetCertificationByID(ctx context.Context, id string) (*model.WorkforceCertification, error) {
	query := `
		SELECT id, workspace_id, user_id, cert_code, name, issuer,
		       credential_id, issued_at, expires_at, status, created_at, updated_at
		FROM workforce_certifications
		WHERE id = ?`

	c, err := scanWFCert(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWFCertificationNotFound
		}
		return nil, err
	}
	return c, nil
}

// UpdateCertification updates credential details, expiry, or status on a certification
func (r *sqlWorkforceRepository) UpdateCertification(ctx context.Context, c *model.WorkforceCertification) error {
	c.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE workforce_certifications
		SET credential_id = ?, expires_at = ?, status = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query, c.CredentialID, c.ExpiresAt, c.Status, c.UpdatedAt, c.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWFCertificationNotFound
	}
	return nil
}

// DeleteCertification removes a certification record
func (r *sqlWorkforceRepository) DeleteCertification(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM workforce_certifications WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWFCertificationNotFound
	}
	return nil
}

// ListCertifications queries certifications based on filter criteria
func (r *sqlWorkforceRepository) ListCertifications(ctx context.Context, filter model.WFCertFilter) ([]model.WorkforceCertification, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.Issuer != "" {
		conditions = append(conditions, "issuer LIKE ?")
		args = append(args, "%"+filter.Issuer+"%")
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM workforce_certifications %s", whereClause), args...).Scan(&total)
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
		SELECT id, workspace_id, user_id, cert_code, name, issuer,
		       credential_id, issued_at, expires_at, status, created_at, updated_at
		FROM workforce_certifications
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var certs []model.WorkforceCertification
	for rows.Next() {
		c, err := scanWFCert(rows)
		if err != nil {
			return nil, 0, err
		}
		certs = append(certs, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return certs, total, nil
}

// CreateRoleCompetency inserts a new role competency requirement
func (r *sqlWorkforceRepository) CreateRoleCompetency(ctx context.Context, rc *model.RoleCompetency) error {
	if rc.ID == "" {
		rc.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if rc.CreatedAt.IsZero() {
		rc.CreatedAt = now
	}
	rc.UpdatedAt = now
	if rc.RequiredLevel == "" {
		rc.RequiredLevel = model.WFProficiencyIntermediate
	}
	if rc.MinHeadcount <= 0 {
		rc.MinHeadcount = 1
	}
	if rc.Priority == "" {
		rc.Priority = model.WFPriorityMedium
	}

	query := `
		INSERT INTO workforce_role_competencies (
			id, workspace_id, role_name, skill_id, required_level, min_headcount, priority, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		rc.ID, rc.WorkspaceID, rc.RoleName, rc.SkillID, rc.RequiredLevel, rc.MinHeadcount, rc.Priority, rc.CreatedAt, rc.UpdatedAt,
	)
	return err
}

// GetRoleCompetencyByID retrieves a role competency by ID
func (r *sqlWorkforceRepository) GetRoleCompetencyByID(ctx context.Context, id string) (*model.RoleCompetency, error) {
	query := `
		SELECT id, workspace_id, role_name, skill_id, required_level, min_headcount, priority, created_at, updated_at
		FROM workforce_role_competencies
		WHERE id = ?`

	var rc model.RoleCompetency
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rc.ID, &rc.WorkspaceID, &rc.RoleName, &rc.SkillID, &rc.RequiredLevel, &rc.MinHeadcount, &rc.Priority, &rc.CreatedAt, &rc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWFRoleCompetencyNotFound
		}
		return nil, err
	}
	return &rc, nil
}

// DeleteRoleCompetency removes a role competency requirement
func (r *sqlWorkforceRepository) DeleteRoleCompetency(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM workforce_role_competencies WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWFRoleCompetencyNotFound
	}
	return nil
}

// ListRoleCompetencies queries role competencies based on filter criteria
func (r *sqlWorkforceRepository) ListRoleCompetencies(ctx context.Context, filter model.WFRoleCompetencyFilter) ([]model.RoleCompetency, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.RoleName != "" {
		conditions = append(conditions, "role_name LIKE ?")
		args = append(args, "%"+filter.RoleName+"%")
	}
	if filter.SkillID != "" {
		conditions = append(conditions, "skill_id = ?")
		args = append(args, filter.SkillID)
	}
	if filter.Priority != "" {
		conditions = append(conditions, "priority = ?")
		args = append(args, filter.Priority)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM workforce_role_competencies %s", whereClause), args...).Scan(&total)
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
		SELECT id, workspace_id, role_name, skill_id, required_level, min_headcount, priority, created_at, updated_at
		FROM workforce_role_competencies
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var roles []model.RoleCompetency
	for rows.Next() {
		var rc model.RoleCompetency
		if err := rows.Scan(
			&rc.ID, &rc.WorkspaceID, &rc.RoleName, &rc.SkillID, &rc.RequiredLevel, &rc.MinHeadcount, &rc.Priority, &rc.CreatedAt, &rc.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		roles = append(roles, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// CreateTraining inserts a new training enrollment
func (r *sqlWorkforceRepository) CreateTraining(ctx context.Context, t *model.TrainingEnrollment) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	if t.Status == "" {
		t.Status = model.WFTrainingAssigned
	}

	query := `
		INSERT INTO workforce_trainings (
			id, workspace_id, user_id, skill_id, course_title, provider,
			status, progress_pct, due_date, completed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.WorkspaceID, t.UserID, t.SkillID, t.CourseTitle, t.Provider,
		t.Status, t.ProgressPct, t.DueDate, t.CompletedAt, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func scanWFTraining(scanner interface{ Scan(dest ...interface{}) error }) (*model.TrainingEnrollment, error) {
	var t model.TrainingEnrollment
	var provider sql.NullString
	var dueDate, completedAt sql.NullTime
	err := scanner.Scan(
		&t.ID, &t.WorkspaceID, &t.UserID, &t.SkillID, &t.CourseTitle, &provider,
		&t.Status, &t.ProgressPct, &dueDate, &completedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if provider.Valid {
		t.Provider = provider.String
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.Time
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	return &t, nil
}

// GetTrainingByID retrieves a training enrollment by ID
func (r *sqlWorkforceRepository) GetTrainingByID(ctx context.Context, id string) (*model.TrainingEnrollment, error) {
	query := `
		SELECT id, workspace_id, user_id, skill_id, course_title, provider,
		       status, progress_pct, due_date, completed_at, created_at, updated_at
		FROM workforce_trainings
		WHERE id = ?`

	t, err := scanWFTraining(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWFTrainingNotFound
		}
		return nil, err
	}
	return t, nil
}

// UpdateTraining updates status, progress percentage, and completion timestamp on a training enrollment
func (r *sqlWorkforceRepository) UpdateTraining(ctx context.Context, t *model.TrainingEnrollment) error {
	t.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE workforce_trainings
		SET status = ?, progress_pct = ?, completed_at = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query, t.Status, t.ProgressPct, t.CompletedAt, t.UpdatedAt, t.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWFTrainingNotFound
	}
	return nil
}

// DeleteTraining removes a training enrollment
func (r *sqlWorkforceRepository) DeleteTraining(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM workforce_trainings WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWFTrainingNotFound
	}
	return nil
}

// ListTrainings queries training enrollments based on filter criteria
func (r *sqlWorkforceRepository) ListTrainings(ctx context.Context, filter model.WFTrainingFilter) ([]model.TrainingEnrollment, int, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != "" {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.SkillID != "" {
		conditions = append(conditions, "skill_id = ?")
		args = append(args, filter.SkillID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM workforce_trainings %s", whereClause), args...).Scan(&total)
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
		SELECT id, workspace_id, user_id, skill_id, course_title, provider,
		       status, progress_pct, due_date, completed_at, created_at, updated_at
		FROM workforce_trainings
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, selectQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var trainings []model.TrainingEnrollment
	for rows.Next() {
		t, err := scanWFTraining(rows)
		if err != nil {
			return nil, 0, err
		}
		trainings = append(trainings, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return trainings, total, nil
}

// Analytics stub
func (r *sqlWorkforceRepository) GetWorkforceAnalytics(ctx context.Context, workspaceID string) (*model.WorkforceAnalytics, error) {
	return nil, errors.New("not implemented")
}
