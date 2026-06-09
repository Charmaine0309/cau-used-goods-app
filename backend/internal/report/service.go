package report

import (
	"context"
	"database/sql"
	"fmt"

	"cau-used-goods-app/backend/internal/message"
	"cau-used-goods-app/backend/internal/sensitive"
)

type Service struct {
	repo      *Repository
	db        *sql.DB
	sensitive *sensitive.Service
	message   *message.Service
}

func NewService(repo *Repository, db *sql.DB, sensitiveService *sensitive.Service, messageService *message.Service) *Service {
	return &Service{repo: repo, db: db, sensitive: sensitiveService, message: messageService}
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
	// 检查是否已举报
	reported, err := s.repo.HasReported(ctx, input.ReporterID, input.TargetType, input.TargetID)
	if err != nil {
		return nil, err
	}
	if reported {
		return nil, fmt.Errorf("you have already reported this target")
	}

	// 敏感词检测
	if input.Description != nil && *input.Description != "" {
		checkResult, err := s.sensitive.CheckText(ctx, *input.Description)
		if err != nil {
			return nil, fmt.Errorf("sensitive word check failed: %w", err)
		}
		if !checkResult.Passed {
			return nil, fmt.Errorf("report description contains sensitive words: %v", checkResult.HitWords)
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
}

type MarkReportProcessingInput struct {
	ReportID  uint64
	HandlerID uint64
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

	if err := s.repo.MarkProcessing(ctx, input.ReportID, input.HandlerID); err != nil {
		return nil, err
	}

	result := "举报标记为处理中"
	_ = s.logAdminAction(ctx, input.HandlerID, input.ReportID, "PROCESSING", &result)

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

	if err := s.repo.UpdateStatus(ctx, input.ReportID, input.Status, input.HandleResult, input.HandlerID); err != nil {
		return nil, err
	}

	// 写入管理员操作日志
	_ = s.logAdminAction(ctx, input.HandlerID, input.ReportID, input.Status, input.HandleResult)

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

func (s *Service) logAdminAction(ctx context.Context, adminID, reportID uint64, status string, handleResult *string) error {
	var actionType string
	switch status {
	case "APPROVED":
		actionType = "REPORT_APPROVE"
	case "REJECTED":
		actionType = "REPORT_REJECT"
	default:
		actionType = "REPORT_HANDLE"
	}

	result := ""
	if handleResult != nil {
		result = *handleResult
	}

	query := `
		INSERT INTO admin_logs (admin_id, operation_type, target_type, target_id, description)
		VALUES (?, ?, 'REPORT', ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query, adminID, actionType, reportID, result)
	return err
}
