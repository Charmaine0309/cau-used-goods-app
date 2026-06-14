package report

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cau-used-goods-app/backend/internal/admin"
	"cau-used-goods-app/backend/internal/db"
	"cau-used-goods-app/backend/internal/message"
	"cau-used-goods-app/backend/internal/sensitive"
)

type Service struct {
	repo      *Repository
	sensitive *sensitive.Service
	message   *message.Service
	admin     *admin.Service
}

func NewService(repo *Repository, sensitiveService *sensitive.Service, messageService *message.Service, adminService *admin.Service) *Service {
	return &Service{repo: repo, sensitive: sensitiveService, message: messageService, admin: adminService}
}

type CreateReportInput struct {
	ReporterID  uint64
	TargetType  string
	TargetID    uint64
	ReasonType  string
	Description *string
	Images      []string
}

func (s *Service) Create(ctx context.Context, input CreateReportInput) (*ReportDetail, error) {
	input.TargetType = strings.ToUpper(strings.TrimSpace(input.TargetType))
	input.ReasonType = strings.TrimSpace(input.ReasonType)
	if input.ReporterID == 0 {
		return nil, fmt.Errorf("reporter id is required")
	}
	if input.TargetID == 0 {
		return nil, fmt.Errorf("targetId is required")
	}
	if input.TargetType != "PRODUCT" && input.TargetType != "USER" && input.TargetType != "ORDER" {
		return nil, fmt.Errorf("invalid targetType")
	}
	if input.ReasonType == "" {
		return nil, fmt.Errorf("reasonType is required")
	}
	if err := s.repo.ValidateTarget(ctx, input.ReporterID, input.TargetType, input.TargetID); err != nil {
		return nil, err
	}

	// 检查是否已举报
	reported, err := s.repo.HasReported(ctx, input.ReporterID, input.TargetType, input.TargetID)
	if err != nil {
		return nil, err
	}
	if reported {
		return nil, fmt.Errorf("you have already reported this target")
	}

	// 敏感词检测
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
		if description != "" && s.sensitive != nil {
			checkResult, err := s.sensitive.CheckText(ctx, description)
			if err != nil {
				return nil, fmt.Errorf("sensitive word check failed: %w", err)
			}
			if !checkResult.Passed {
				return nil, fmt.Errorf("report description contains sensitive words: %v", checkResult.HitWords)
			}
		}
	}

	report := &Report{
		ReporterID:  input.ReporterID,
		TargetType:  input.TargetType,
		TargetID:    input.TargetID,
		ReasonType:  input.ReasonType,
		Description: input.Description,
		Status:      "PENDING",
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	if len(input.Images) > 0 {
		if err := s.repo.AddImages(ctx, report.ID, input.Images); err != nil {
			return nil, err
		}
	}

	return s.GetByID(ctx, report.ID)
}

func (s *Service) GetByID(ctx context.Context, id uint64) (*ReportDetail, error) {
	rd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rd == nil {
		return nil, fmt.Errorf("report not found")
	}
	images, err := s.repo.GetImages(ctx, id)
	if err != nil {
		return nil, err
	}
	rd.Images = images
	return rd, nil
}

func (s *Service) ListByReporter(ctx context.Context, reporterID uint64, page, pageSize int) ([]ReportDetail, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	items, total, err := s.repo.ListByReporter(ctx, reporterID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		images, err := s.repo.GetImages(ctx, items[i].ID)
		if err != nil {
			return nil, 0, err
		}
		items[i].Images = images
	}
	return items, total, nil
}

func (s *Service) ListAll(ctx context.Context, status string, page, pageSize int) ([]ReportDetail, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	items, total, err := s.repo.ListAll(ctx, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		images, err := s.repo.GetImages(ctx, items[i].ID)
		if err != nil {
			return nil, 0, err
		}
		items[i].Images = images
	}
	return items, total, nil
}

type HandleReportInput struct {
	ReportID     uint64
	HandlerID    uint64
	Status       string
	HandleResult *string
	IPAddress    *string
}

type MarkReportProcessingInput struct {
	ReportID  uint64
	HandlerID uint64
	IPAddress *string
}

type CloseReportInput struct {
	ReportID    uint64
	ReporterID  uint64
	CloseReason *string
}

func (s *Service) MarkProcessing(ctx context.Context, input MarkReportProcessingInput) (*ReportDetail, error) {
	if input.ReportID == 0 {
		return nil, fmt.Errorf("report id is required")
	}
	if input.HandlerID == 0 {
		return nil, fmt.Errorf("handler id is required")
	}

	description := "mark report as PROCESSING"
	if err := db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.MarkProcessingTx(ctx, tx, input.ReportID, input.HandlerID); err != nil {
			return err
		}
		return s.logAdminActionTx(ctx, tx, input.HandlerID, input.ReportID, admin.OperationMarkReportProcessing, description, input.IPAddress)
	}); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, input.ReportID)
}

func (s *Service) Handle(ctx context.Context, input HandleReportInput) (*ReportDetail, error) {
	if input.Status != "APPROVED" && input.Status != "REJECTED" {
		return nil, fmt.Errorf("status must be APPROVED or REJECTED")
	}

	report, err := s.repo.GetByID(ctx, input.ReportID)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, fmt.Errorf("report not found")
	}
	if report.Status != "PENDING" && report.Status != "PROCESSING" {
		return nil, fmt.Errorf("report cannot be handled")
	}

	if err := db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.UpdateStatusTx(ctx, tx, input.ReportID, input.Status, input.HandleResult, input.HandlerID); err != nil {
			return err
		}
		description := fmt.Sprintf("handle report #%d: %s", input.ReportID, input.Status)
		if input.HandleResult != nil && strings.TrimSpace(*input.HandleResult) != "" {
			description = fmt.Sprintf("%s; result=%s", description, strings.TrimSpace(*input.HandleResult))
		}
		return s.logAdminActionTx(ctx, tx, input.HandlerID, input.ReportID, reportOperationType(input.Status), description, input.IPAddress)
	}); err != nil {
		return nil, err
	}

	// 通知举报人处理结果
	if s.message != nil {
		title := "举报处理结果"
		content := fmt.Sprintf("你的举报已处理，结果：%s", input.Status)
		if input.HandleResult != nil {
			content = fmt.Sprintf("你的举报已处理，结果：%s。处理说明：%s", input.Status, *input.HandleResult)
		}
		relatedType := message.RelatedTypeReport
		_, _ = s.message.Create(ctx, message.CreateMessageInput{
			ReceiverID:  report.ReporterID,
			SenderID:    &input.HandlerID,
			MessageType: message.MessageTypeReportHandled,
			Title:       title,
			Content:     content,
			RelatedType: &relatedType,
			RelatedID:   &input.ReportID,
		})
	}

	return s.GetByID(ctx, input.ReportID)
}

func (s *Service) Close(ctx context.Context, input CloseReportInput) (*ReportDetail, error) {
	if input.ReportID == 0 {
		return nil, fmt.Errorf("report id is required")
	}
	if input.ReporterID == 0 {
		return nil, fmt.Errorf("reporter id is required")
	}

	item, err := s.repo.GetByID(ctx, input.ReportID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("report not found")
	}
	if item.ReporterID != input.ReporterID {
		return nil, fmt.Errorf("permission denied")
	}
	if item.Status != "PENDING" {
		return nil, fmt.Errorf("report cannot be closed")
	}

	if err := s.repo.Close(ctx, input.ReportID, input.ReporterID, input.CloseReason); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, input.ReportID)
}

func reportOperationType(status string) string {
	switch status {
	case "APPROVED":
		return admin.OperationApproveReport
	case "REJECTED":
		return admin.OperationRejectReport
	default:
		return admin.OperationMarkReportProcessing
	}
}

func (s *Service) logAdminActionTx(ctx context.Context, tx *sql.Tx, adminID, reportID uint64, operationType string, description string, ipAddress *string) error {
	if s.admin == nil {
		return fmt.Errorf("admin logger is not configured")
	}
	_, err := s.admin.LogActionTx(ctx, tx, admin.LogActionInput{
		AdminID:       adminID,
		OperationType: operationType,
		TargetType:    admin.TargetTypeReport,
		TargetID:      reportID,
		Description:   &description,
		IPAddress:     ipAddress,
	})
	return err
}
