package sensitive

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"cau-used-goods-app/backend/internal/admin"
	"cau-used-goods-app/backend/internal/db"
)

const (
	WordTypeForbidden  = "FORBIDDEN"
	WordTypeRisk       = "RISK"
	WordStatusEnabled  = "ENABLED"
	WordStatusDisabled = "DISABLED"
)

var ErrInvalidSensitiveWordInput = errors.New("invalid sensitive word input")

type Service struct {
	repo        *Repository
	adminLogger *admin.Service
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetAdminLogger(adminLogger *admin.Service) {
	s.adminLogger = adminLogger
}

type CheckResult struct {
	Passed   bool     `json:"passed"`
	HitWords []string `json:"hitWords"`
	Message  string   `json:"message"`
}

func (s *Service) CheckText(ctx context.Context, text string) (*CheckResult, error) {
	words, err := s.repo.ListEnabledWords(ctx)
	if err != nil {
		return nil, err
	}

	text = strings.ToLower(text)
	var hits []string

	for _, w := range words {
		word := strings.ToLower(strings.TrimSpace(w.Word))
		if word == "" {
			continue
		}
		if strings.Contains(text, word) {
			hits = append(hits, w.Word)
		}
	}

	if len(hits) > 0 {
		return &CheckResult{
			Passed:   false,
			HitWords: hits,
			Message:  "内容包含敏感词，请修改后重试",
		}, nil
	}

	return &CheckResult{
		Passed:   true,
		HitWords: []string{},
		Message:  "内容检测通过",
	}, nil
}

func (s *Service) ListWords(ctx context.Context, query WordQuery) ([]SensitiveWord, int, error) {
	query.Status = strings.TrimSpace(query.Status)
	query.WordType = strings.TrimSpace(query.WordType)
	query.Keyword = strings.TrimSpace(query.Keyword)
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	if query.Status != "" && !isValidWordStatus(query.Status) {
		return nil, 0, ErrInvalidSensitiveWordInput
	}
	if query.WordType != "" && !isValidWordType(query.WordType) {
		return nil, 0, ErrInvalidSensitiveWordInput
	}
	return s.repo.ListWords(ctx, query)
}

func (s *Service) CreateWord(ctx context.Context, adminID uint64, input CreateWordInput, ipAddress *string) (uint64, error) {
	input.Word = strings.TrimSpace(input.Word)
	input.WordType = strings.TrimSpace(input.WordType)
	input.Status = strings.TrimSpace(input.Status)
	input.CreateBy = adminID
	trimOptionalString(&ipAddress)
	if input.Status == "" {
		input.Status = WordStatusEnabled
	}

	if err := validateCreateWordInput(input); err != nil {
		return 0, err
	}

	description := fmt.Sprintf("create sensitive word: %s", input.Word)
	var id uint64
	err := db.WithTx(ctx, func(tx *sql.Tx) error {
		var err error
		id, err = s.repo.CreateWordTx(ctx, tx, input)
		if err != nil {
			return err
		}
		return s.logAdminActionTx(ctx, tx, adminID, admin.OperationCreateWord, id, description, ipAddress)
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Service) UpdateWord(ctx context.Context, adminID uint64, input UpdateWordInput, ipAddress *string) error {
	input.Word = strings.TrimSpace(input.Word)
	input.WordType = strings.TrimSpace(input.WordType)
	input.Status = strings.TrimSpace(input.Status)
	trimOptionalString(&ipAddress)

	if err := validateUpdateWordInput(adminID, input); err != nil {
		return err
	}
	description := fmt.Sprintf("update sensitive word: %s", input.Word)
	return db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.UpdateWordTx(ctx, tx, input); err != nil {
			return err
		}
		return s.logAdminActionTx(ctx, tx, adminID, admin.OperationUpdateWord, input.ID, description, ipAddress)
	})
}

func (s *Service) DeleteWord(ctx context.Context, adminID, id uint64, ipAddress *string) error {
	trimOptionalString(&ipAddress)
	if adminID == 0 || id == 0 {
		return ErrInvalidSensitiveWordInput
	}
	description := "disable sensitive word by delete operation"
	return db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.DisableWordTx(ctx, tx, id); err != nil {
			return err
		}
		return s.logAdminActionTx(ctx, tx, adminID, admin.OperationDeleteWord, id, description, ipAddress)
	})
}

func (s *Service) logAdminActionTx(ctx context.Context, tx *sql.Tx, adminID uint64, operationType string, targetID uint64, description string, ipAddress *string) error {
	if s.adminLogger == nil {
		return fmt.Errorf("admin logger is not configured")
	}
	_, err := s.adminLogger.LogActionTx(ctx, tx, admin.LogActionInput{
		AdminID:       adminID,
		OperationType: operationType,
		TargetType:    admin.TargetTypeWord,
		TargetID:      targetID,
		Description:   &description,
		IPAddress:     ipAddress,
	})
	return err
}

func validateCreateWordInput(input CreateWordInput) error {
	if input.CreateBy == 0 {
		return ErrInvalidSensitiveWordInput
	}
	if input.Word == "" || len(input.Word) > 100 {
		return ErrInvalidSensitiveWordInput
	}
	if !isValidWordType(input.WordType) || !isValidWordStatus(input.Status) {
		return ErrInvalidSensitiveWordInput
	}
	return nil
}

func validateUpdateWordInput(adminID uint64, input UpdateWordInput) error {
	if adminID == 0 || input.ID == 0 {
		return ErrInvalidSensitiveWordInput
	}
	if input.Word == "" || len(input.Word) > 100 {
		return ErrInvalidSensitiveWordInput
	}
	if !isValidWordType(input.WordType) || !isValidWordStatus(input.Status) {
		return ErrInvalidSensitiveWordInput
	}
	return nil
}

func isValidWordType(wordType string) bool {
	return wordType == WordTypeForbidden || wordType == WordTypeRisk
}

func isValidWordStatus(status string) bool {
	return status == WordStatusEnabled || status == WordStatusDisabled
}

func trimOptionalString(value **string) {
	if value == nil || *value == nil {
		return
	}
	trimmed := strings.TrimSpace(**value)
	if trimmed == "" {
		*value = nil
		return
	}
	*value = &trimmed
}
