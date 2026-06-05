package product

import (
	"context"
	"fmt"
	"strings"

	"cau-used-goods-app/backend/internal/admin"
	"cau-used-goods-app/backend/internal/sensitive"
)

type Service struct {
	repo             *Repository
	sensitiveService *sensitive.Service
	adminLogger      *admin.Service
}

func NewService(repo *Repository, sensitiveService *sensitive.Service) *Service {
	return &Service{
		repo:             repo,
		sensitiveService: sensitiveService,
	}
}

func (s *Service) SetAdminLogger(adminLogger *admin.Service) {
	s.adminLogger = adminLogger
}

func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *Service) ListAllCategories(ctx context.Context, status string) ([]Category, error) {
	status = strings.TrimSpace(status)
	if status != "" && !isValidCategoryStatus(status) {
		return nil, fmt.Errorf("invalid category status")
	}
	return s.repo.ListAllCategories(ctx, status)
}

type CategoryCreateInput struct {
	AdminID   uint64
	Name      string
	ParentID  uint64
	SortOrder int
	Status    string
}

func (s *Service) CreateCategory(ctx context.Context, input CategoryCreateInput) (uint64, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = "ENABLED"
	}
	if err := validateCategoryInput(input.Name, input.Status); err != nil {
		return 0, err
	}
	id, err := s.repo.CreateCategory(ctx, CreateCategoryInput{
		Name:      input.Name,
		ParentID:  input.ParentID,
		SortOrder: input.SortOrder,
		Status:    input.Status,
	})
	if err != nil {
		return 0, err
	}
	description := fmt.Sprintf("新增标签：%s", input.Name)
	if err := s.logAdminAction(ctx, input.AdminID, admin.OperationCategoryCreate, id, description); err != nil {
		return 0, err
	}
	return id, nil
}

type CategoryUpdateInput struct {
	AdminID   uint64
	ID        uint64
	Name      string
	ParentID  uint64
	SortOrder int
	Status    string
}

func (s *Service) UpdateCategory(ctx context.Context, input CategoryUpdateInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = "ENABLED"
	}
	if input.ID == 0 {
		return fmt.Errorf("category id is required")
	}
	if err := validateCategoryInput(input.Name, input.Status); err != nil {
		return err
	}
	if input.ParentID == input.ID {
		return fmt.Errorf("category parent cannot be itself")
	}
	if err := s.repo.UpdateCategory(ctx, UpdateCategoryInput{
		ID:        input.ID,
		Name:      input.Name,
		ParentID:  input.ParentID,
		SortOrder: input.SortOrder,
		Status:    input.Status,
	}); err != nil {
		return err
	}
	description := fmt.Sprintf("编辑标签：%s", input.Name)
	return s.logAdminAction(ctx, input.AdminID, admin.OperationCategoryUpdate, input.ID, description)
}

func (s *Service) UpdateCategoryStatus(ctx context.Context, adminID uint64, id uint64, status string) error {
	status = strings.TrimSpace(status)
	if id == 0 {
		return fmt.Errorf("category id is required")
	}
	if !isValidCategoryStatus(status) {
		return fmt.Errorf("invalid category status")
	}
	if err := s.repo.UpdateCategoryStatus(ctx, id, status); err != nil {
		return err
	}
	operationType := admin.OperationCategoryEnable
	description := "启用标签"
	if status == "DISABLED" {
		operationType = admin.OperationCategoryDisable
		description = "停用标签"
	}
	return s.logAdminAction(ctx, adminID, operationType, id, description)
}

func (s *Service) logAdminAction(ctx context.Context, adminID uint64, operationType string, targetID uint64, description string) error {
	if s.adminLogger == nil || adminID == 0 {
		return nil
	}
	_, err := s.adminLogger.LogAction(ctx, admin.LogActionInput{
		AdminID:       adminID,
		OperationType: operationType,
		TargetType:    admin.TargetTypeCategory,
		TargetID:      targetID,
		Description:   &description,
	})
	return err
}

func validateCategoryInput(name string, status string) error {
	if name == "" {
		return fmt.Errorf("category name is required")
	}
	if len([]rune(name)) > 50 {
		return fmt.Errorf("category name cannot exceed 50 characters")
	}
	if !isValidCategoryStatus(status) {
		return fmt.Errorf("invalid category status")
	}
	return nil
}

func isValidCategoryStatus(status string) bool {
	return status == "ENABLED" || status == "DISABLED"
}

type ProductCreateInput struct {
	SellerID       uint64
	CategoryID     uint64
	Title          string
	Description    string
	OriginalPrice  *float64
	Price          float64
	ConditionLevel string
	MeetLocation   string
}

func (s *Service) CreateProduct(ctx context.Context, input ProductCreateInput) (uint64, error) {
	if err := s.checkSensitive(ctx, input.Title, input.Description); err != nil {
		return 0, err
	}

	return s.repo.CreateProduct(ctx, CreateProductInput{
		SellerID:       input.SellerID,
		CategoryID:     input.CategoryID,
		Title:          input.Title,
		Description:    input.Description,
		OriginalPrice:  input.OriginalPrice,
		Price:          input.Price,
		ConditionLevel: input.ConditionLevel,
		MeetLocation:   input.MeetLocation,
	})
}

type ProductListInput struct {
	Keyword        string
	CategoryID     uint64
	ConditionLevel string
	Status         string
	MinPrice       *float64
	MaxPrice       *float64
	Sort           string
	Page           int
	PageSize       int
}

func (s *Service) ListProducts(ctx context.Context, input ProductListInput) (*ProductListResult, error) {
	return s.repo.ListProducts(ctx, ListProductsInput{
		Keyword:        input.Keyword,
		CategoryID:     input.CategoryID,
		ConditionLevel: input.ConditionLevel,
		Status:         input.Status,
		MinPrice:       input.MinPrice,
		MaxPrice:       input.MaxPrice,
		Sort:           input.Sort,
		Page:           input.Page,
		PageSize:       input.PageSize,
	})
}

func (s *Service) GetProductByID(ctx context.Context, id uint64) (*Product, error) {
	return s.repo.GetProductByID(ctx, id)
}

func (s *Service) ListMyProducts(ctx context.Context, sellerID uint64) ([]Product, error) {
	return s.repo.ListMyProducts(ctx, sellerID)
}

func (s *Service) DeleteProduct(ctx context.Context, productID uint64, sellerID uint64) error {
	return s.repo.DeleteProduct(ctx, productID, sellerID)
}

type ProductUpdateInput struct {
	ProductID      uint64
	SellerID       uint64
	CategoryID     uint64
	Title          string
	Description    string
	OriginalPrice  *float64
	Price          float64
	ConditionLevel string
	MeetLocation   string
}

func (s *Service) UpdateProduct(ctx context.Context, input ProductUpdateInput) error {
	if err := s.checkSensitive(ctx, input.Title, input.Description); err != nil {
		return err
	}

	return s.repo.UpdateProduct(ctx, UpdateProductInput{
		ProductID:      input.ProductID,
		SellerID:       input.SellerID,
		CategoryID:     input.CategoryID,
		Title:          input.Title,
		Description:    input.Description,
		OriginalPrice:  input.OriginalPrice,
		Price:          input.Price,
		ConditionLevel: input.ConditionLevel,
		MeetLocation:   input.MeetLocation,
	})
}

func (s *Service) UpdateProductStatus(ctx context.Context, productID uint64, sellerID uint64, role string, status string, reason string) error {
	if role == "ADMIN" {
		return s.repo.UpdateProductStatusByAdmin(ctx, productID, sellerID, status, reason)
	}
	return s.repo.UpdateProductStatus(ctx, productID, sellerID, status, reason)
}

type ProductImagesInput struct {
	ProductID uint64
	SellerID  uint64
	Images    []string
}

func (s *Service) AddProductImages(ctx context.Context, input ProductImagesInput) error {
	return s.repo.AddProductImages(ctx, AddProductImagesInput{
		ProductID: input.ProductID,
		SellerID:  input.SellerID,
		Images:    input.Images,
	})
}

func (s *Service) checkSensitive(ctx context.Context, title string, description string) error {
	if s.sensitiveService == nil {
		return nil
	}

	checkText := strings.TrimSpace(title + " " + description)
	result, err := s.sensitiveService.CheckText(ctx, checkText)
	if err != nil {
		return err
	}

	if !result.Passed {
		return fmt.Errorf("%s：%s", result.Message, strings.Join(result.HitWords, "、"))
	}

	return nil
}

func (s *Service) LockProduct(ctx context.Context, productID uint64) error {
	return s.repo.LockProduct(ctx, productID)
}

func (s *Service) UnlockProduct(ctx context.Context, productID uint64) error {
	return s.repo.UnlockProduct(ctx, productID)
}

func (s *Service) MarkProductSold(ctx context.Context, productID uint64) error {
	return s.repo.MarkProductSold(ctx, productID)
}
