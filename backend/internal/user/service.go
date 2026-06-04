package user

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

const (
	accountStatusNormal = "NORMAL"
	authStatusPending   = "PENDING"
	authStatusVerified  = "VERIFIED"
	authStatusRejected  = "REJECTED"
)

var (
	phonePattern     = regexp.MustCompile(`^1[3-9]\d{9}$`)
	studentIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{6,30}$`)
)

type Service struct {
	repo *Repository
}

type UpdateProfileInput struct {
	Nickname  *string
	AvatarURL *string
	Phone     *string
}

type SubmitStudentVerificationInput struct {
	StudentID string
	RealName  string
	College   string
}

type ReviewStudentVerificationInput struct {
	UserID      uint64
	AuthStatus  string
	Description string
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Me(ctx context.Context, userID uint64) (*User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *Service) EnsureAccountNormal(ctx context.Context, userID uint64) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}
	if user.AccountStatus != accountStatusNormal {
		return fmt.Errorf("当前账号状态不可操作")
	}
	return nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uint64, input UpdateProfileInput) (*User, error) {
	if input.Nickname == nil && input.AvatarURL == nil && input.Phone == nil {
		return nil, fmt.Errorf("nothing to update")
	}
	if err := s.EnsureAccountNormal(ctx, userID); err != nil {
		return nil, err
	}

	if input.Nickname != nil {
		trimmed := strings.TrimSpace(*input.Nickname)
		if err := validateStringLength("昵称", trimmed, 1, 50); err != nil {
			return nil, err
		}
		input.Nickname = &trimmed
	}
	if input.AvatarURL != nil {
		trimmed := strings.TrimSpace(*input.AvatarURL)
		if err := validateStringLength("头像地址", trimmed, 1, 255); err != nil {
			return nil, err
		}
		if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") && !strings.HasPrefix(trimmed, "/uploads/avatar/") {
			return nil, fmt.Errorf("头像地址必须是 http(s) URL 或 /uploads/avatar/ 路径")
		}
		input.AvatarURL = &trimmed
	}
	if input.Phone != nil {
		trimmed := strings.TrimSpace(*input.Phone)
		if !phonePattern.MatchString(trimmed) {
			return nil, fmt.Errorf("手机号格式不正确")
		}
		input.Phone = &trimmed
	}

	if err := s.repo.UpdateProfile(ctx, userID, input.Nickname, input.AvatarURL, input.Phone); err != nil {
		return nil, err
	}
	return s.Me(ctx, userID)
}

func (s *Service) SubmitStudentVerification(ctx context.Context, userID uint64, input SubmitStudentVerificationInput) (*StudentVerification, error) {
	input.StudentID = strings.TrimSpace(input.StudentID)
	input.RealName = strings.TrimSpace(input.RealName)
	input.College = strings.TrimSpace(input.College)

	if !studentIDPattern.MatchString(input.StudentID) {
		return nil, fmt.Errorf("学号格式不正确")
	}
	if err := validateStringLength("真实姓名", input.RealName, 2, 30); err != nil {
		return nil, err
	}
	if err := validateStringLength("学院名称", input.College, 1, 50); err != nil {
		return nil, err
	}

	current, err := s.repo.FindStudentVerification(ctx, userID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("用户不存在")
	}
	if current.AccountStatus != accountStatusNormal {
		return nil, fmt.Errorf("当前账号状态不可操作")
	}
	switch current.AuthStatus {
	case authStatusVerified:
		return nil, fmt.Errorf("学生认证已通过，不能重复提交")
	case authStatusPending:
		return nil, fmt.Errorf("学生认证正在审核中，请勿重复提交")
	case "UNVERIFIED", authStatusRejected:
	default:
		return nil, fmt.Errorf("当前认证状态不可提交")
	}

	used, err := s.repo.StudentIDUsedByOther(ctx, input.StudentID, userID)
	if err != nil {
		return nil, err
	}
	if used {
		return nil, fmt.Errorf("学号已被使用")
	}

	if err := s.repo.SubmitStudentVerification(ctx, userID, input.StudentID, input.RealName, input.College); err != nil {
		return nil, err
	}
	return s.StudentVerification(ctx, userID)
}

func (s *Service) StudentVerification(ctx context.Context, userID uint64) (*StudentVerification, error) {
	verification, err := s.repo.FindStudentVerification(ctx, userID)
	if err != nil {
		return nil, err
	}
	if verification == nil {
		return nil, fmt.Errorf("user not found")
	}
	return verification, nil
}

func (s *Service) ListStudentVerifications(ctx context.Context, status string) ([]StudentVerification, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		status = "PENDING"
	}
	switch status {
	case "PENDING", "VERIFIED", "REJECTED", "UNVERIFIED":
	default:
		return nil, fmt.Errorf("invalid authStatus")
	}
	return s.repo.ListStudentVerifications(ctx, status)
}

func (s *Service) ListUsers(ctx context.Context) ([]AdminUserItem, error) {
	return s.repo.ListUsers(ctx)
}

func (s *Service) ReviewStudentVerification(ctx context.Context, adminID uint64, input ReviewStudentVerificationInput) (*StudentVerification, error) {
	if input.UserID == 0 {
		return nil, fmt.Errorf("userId is required")
	}
	if input.UserID == adminID {
		return nil, fmt.Errorf("管理员不能审核自己的认证")
	}
	input.AuthStatus = strings.TrimSpace(input.AuthStatus)
	input.Description = strings.TrimSpace(input.Description)

	if input.AuthStatus != authStatusVerified && input.AuthStatus != authStatusRejected {
		return nil, fmt.Errorf("authStatus must be VERIFIED or REJECTED")
	}
	if len(input.Description) > 500 {
		return nil, fmt.Errorf("审核说明不能超过 500 个字符")
	}
	if input.AuthStatus == authStatusRejected && input.Description == "" {
		return nil, fmt.Errorf("驳回时必须填写审核说明")
	}
	if input.AuthStatus == authStatusVerified && input.Description == "" {
		input.Description = "学生认证审核通过"
	}

	target, err := s.repo.FindStudentVerification(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, fmt.Errorf("用户不存在")
	}
	if target.AccountStatus != accountStatusNormal {
		return nil, fmt.Errorf("目标用户账号状态不可审核")
	}
	if target.AuthStatus != authStatusPending {
		return nil, fmt.Errorf("认证状态已变化，请刷新后重试")
	}

	if err := s.repo.ReviewStudentVerification(ctx, adminID, input.UserID, input.AuthStatus, input.Description); err != nil {
		return nil, err
	}
	return s.StudentVerification(ctx, input.UserID)
}

func validateStringLength(field string, value string, min int, max int) error {
	length := len([]rune(value))
	if length < min || length > max {
		if min == max {
			return fmt.Errorf("%s长度应为 %d 个字符", field, min)
		}
		return fmt.Errorf("%s长度应为 %d-%d 个字符", field, min, max)
	}
	return nil
}
