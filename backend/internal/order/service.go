package order

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cau-used-goods-app/backend/internal/db"
	"cau-used-goods-app/backend/internal/message"
	"cau-used-goods-app/backend/internal/product"
)

type Service struct {
	repo    *Repository
	product *product.Service
	message *message.Service
}

func NewService(repo *Repository, productService *product.Service, messageService *message.Service) *Service {
	return &Service{repo: repo, product: productService, message: messageService}
}

type CreateOrderInput struct {
	ProductID    uint64
	BuyerID      uint64
	Remark       *string
	MeetTime     *string
	MeetLocation *string
}

func (s *Service) Create(ctx context.Context, input CreateOrderInput) (*Order, error) {
	// 不能购买自己的商品
	sellerID, err := s.repo.GetProductSeller(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}
	if sellerID == input.BuyerID {
		return nil, fmt.Errorf("cannot buy your own product")
	}

	// 检查是否已有有效订单
	exists, err := s.repo.HasActiveOrderByBuyer(ctx, input.BuyerID, input.ProductID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("you already have an active order for this product")
	}

	title, price, err := s.repo.GetProductInfo(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}

	expireTime := time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	order := &Order{
		OrderNo:              generateOrderNo(),
		ProductID:            input.ProductID,
		BuyerID:              input.BuyerID,
		SellerID:             sellerID,
		ProductTitleSnapshot: title,
		ProductPriceSnapshot: price,
		Status:               "PENDING_CONFIRM",
		Remark:               input.Remark,
		MeetTime:             input.MeetTime,
		MeetLocation:         input.MeetLocation,
		ExpireTime:           expireTime,
	}

	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.product.LockProduct(ctx, input.ProductID); err != nil {
			return err
		}
		if err := s.repo.Create(ctx, tx, order); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 发送订单创建消息通知卖家
	if s.message != nil {
		relatedType := message.RelatedTypeOrder
		_, _ = s.message.Create(ctx, message.CreateMessageInput{
			ReceiverID:  sellerID,
			MessageType: message.MessageTypeOrderCreated,
			Title:       "新订单提醒",
			Content:     fmt.Sprintf("您的商品「%s」有新的订单，买家已预约，请尽快确认。", title),
			RelatedType: &relatedType,
			RelatedID:   &order.ID,
		})
	}

	return order, nil
}

type ConfirmOrderInput struct {
	OrderID  uint64
	SellerID uint64
}

func (s *Service) Confirm(ctx context.Context, input ConfirmOrderInput) (*Order, error) {
	order, err := s.repo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.SellerID != input.SellerID {
		return nil, fmt.Errorf("permission denied")
	}
	if order.Status != "PENDING_CONFIRM" {
		return nil, fmt.Errorf("order cannot be confirmed")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	err = s.repo.UpdateStatus(ctx, nil, input.OrderID, "WAIT_MEET", map[string]interface{}{
		"confirm_time": now,
	})
	if err != nil {
		return nil, err
	}

	// 发送订单确认消息通知买家
	if s.message != nil {
		relatedType := message.RelatedTypeOrder
		_, _ = s.message.Create(ctx, message.CreateMessageInput{
			ReceiverID:  order.BuyerID,
			MessageType: message.MessageTypeOrderConfirmed,
			Title:       "订单已确认",
			Content:     fmt.Sprintf("卖家已确认您的订单「%s」，请按约定时间地点交易。", order.ProductTitleSnapshot),
			RelatedType: &relatedType,
			RelatedID:   &order.ID,
		})
	}

	return s.repo.GetByID(ctx, input.OrderID)
}

type CancelOrderInput struct {
	OrderID uint64
	UserID  uint64
	Reason  string
}

func (s *Service) Cancel(ctx context.Context, input CancelOrderInput) (*Order, error) {
	order, err := s.repo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.BuyerID != input.UserID && order.SellerID != input.UserID {
		return nil, fmt.Errorf("permission denied")
	}
	if order.Status != "PENDING_CONFIRM" && order.Status != "WAIT_MEET" {
		return nil, fmt.Errorf("order cannot be cancelled")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.product.UnlockProduct(ctx, order.ProductID); err != nil {
			return err
		}
		if err := s.repo.UpdateStatus(ctx, tx, input.OrderID, "CANCELED", map[string]interface{}{
			"cancel_reason": input.Reason,
			"cancel_by":     input.UserID,
			"close_time":    now,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 发送订单取消消息通知对方
	if s.message != nil {
		relatedType := message.RelatedTypeOrder
		receiverID := order.BuyerID
		if input.UserID == order.BuyerID {
			receiverID = order.SellerID
		}
		_, _ = s.message.Create(ctx, message.CreateMessageInput{
			ReceiverID:  receiverID,
			MessageType: message.MessageTypeOrderCanceled,
			Title:       "订单已取消",
			Content:     fmt.Sprintf("订单「%s」已被取消，原因：%s", order.ProductTitleSnapshot, input.Reason),
			RelatedType: &relatedType,
			RelatedID:   &order.ID,
		})
	}

	return s.repo.GetByID(ctx, input.OrderID)
}

type CompleteOrderInput struct {
	OrderID  uint64
	SellerID uint64
}

type ExceptionCloseOrderInput struct {
	OrderID uint64
	UserID  uint64
	Reason  string
}

type AdminUpdateOrderStatusInput struct {
	AdminID   uint64
	OrderID   uint64
	Status    string
	Reason    string
	IPAddress *string
}

func (s *Service) Complete(ctx context.Context, input CompleteOrderInput) (*Order, error) {
	order, err := s.repo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.SellerID != input.SellerID {
		return nil, fmt.Errorf("permission denied")
	}
	if order.Status != "WAIT_MEET" {
		return nil, fmt.Errorf("order cannot be completed")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.product.MarkProductSold(ctx, order.ProductID); err != nil {
			return err
		}
		if err := s.repo.UpdateStatus(ctx, tx, input.OrderID, "COMPLETED", map[string]interface{}{
			"finish_time": now,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 发送订单完成消息通知买家
	if s.message != nil {
		relatedType := message.RelatedTypeOrder
		_, _ = s.message.Create(ctx, message.CreateMessageInput{
			ReceiverID:  order.BuyerID,
			MessageType: message.MessageTypeOrderConfirmed,
			Title:       "交易完成",
			Content:     fmt.Sprintf("订单「%s」已完成交易，欢迎评价。", order.ProductTitleSnapshot),
			RelatedType: &relatedType,
			RelatedID:   &order.ID,
		})
	}

	return s.repo.GetByID(ctx, input.OrderID)
}

func (s *Service) ExceptionClose(ctx context.Context, input ExceptionCloseOrderInput) (*Order, error) {
	order, err := s.repo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.BuyerID != input.UserID && order.SellerID != input.UserID {
		return nil, fmt.Errorf("permission denied")
	}
	if order.Status != "PENDING_CONFIRM" && order.Status != "WAIT_MEET" {
		return nil, fmt.Errorf("order cannot be exception closed")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.product.UnlockProduct(ctx, order.ProductID); err != nil {
			return err
		}
		if err := s.repo.UpdateStatus(ctx, tx, input.OrderID, "EXCEPTION_CLOSED", map[string]interface{}{
			"cancel_reason": input.Reason,
			"cancel_by":     input.UserID,
			"close_time":    now,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, input.OrderID)
}

func (s *Service) AdminUpdateStatus(ctx context.Context, input AdminUpdateOrderStatusInput) (*Order, error) {
	if input.AdminID == 0 {
		return nil, fmt.Errorf("adminId is required")
	}
	if input.OrderID == 0 {
		return nil, fmt.Errorf("orderId is required")
	}
	input.Status = strings.ToUpper(strings.TrimSpace(input.Status))
	input.Reason = strings.TrimSpace(input.Reason)
	if input.IPAddress != nil {
		trimmed := strings.TrimSpace(*input.IPAddress)
		input.IPAddress = &trimmed
	}
	if !isValidAdminOrderStatus(input.Status) {
		return nil, fmt.Errorf("status must be PENDING_CONFIRM, WAIT_MEET, COMPLETED, CANCELED or EXCEPTION_CLOSED")
	}
	if len([]rune(input.Reason)) > 500 {
		return nil, fmt.Errorf("reason cannot exceed 500 characters")
	}

	order, err := s.repo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	updates := map[string]interface{}{}
	switch input.Status {
	case "WAIT_MEET":
		updates["confirm_time"] = now
	case "COMPLETED":
		updates["finish_time"] = now
	case "CANCELED", "EXCEPTION_CLOSED":
		updates["cancel_reason"] = input.Reason
		updates["cancel_by"] = input.AdminID
		updates["close_time"] = now
	}

	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		switch input.Status {
		case "PENDING_CONFIRM", "WAIT_MEET":
			if err := s.repo.UpdateProductStatusForAdmin(ctx, tx, order.ProductID, "LOCKED"); err != nil {
				return err
			}
		case "COMPLETED":
			if err := s.repo.UpdateProductStatusForAdmin(ctx, tx, order.ProductID, "SOLD"); err != nil {
				return err
			}
		case "CANCELED", "EXCEPTION_CLOSED":
			if err := s.repo.UpdateProductStatusForAdmin(ctx, tx, order.ProductID, "ON_SALE"); err != nil {
				return err
			}
		}
		if err := s.repo.UpdateStatus(ctx, tx, input.OrderID, input.Status, updates); err != nil {
			return err
		}
		operationType := "UPDATE_ORDER_STATUS"
		if input.Status == "EXCEPTION_CLOSED" {
			operationType = "ORDER_EXCEPTION_CLOSE"
		}
		return s.repo.CreateAdminLog(ctx, tx, input.AdminID, operationType, "ORDER", input.OrderID, buildStatusDescription(input.Status, input.Reason), input.IPAddress)
	})
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, input.OrderID)
}

func (s *Service) AdminList(ctx context.Context, status string, page, pageSize int) ([]OrderDetail, int, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	if status == "ALL" {
		status = ""
	}
	if status != "" && !isValidAdminOrderStatus(status) {
		return nil, 0, fmt.Errorf("invalid order status")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	return s.repo.ListAll(ctx, status, page, pageSize)
}

func isValidAdminOrderStatus(status string) bool {
	switch status {
	case "PENDING_CONFIRM", "WAIT_MEET", "COMPLETED", "CANCELED", "EXCEPTION_CLOSED":
		return true
	default:
		return false
	}
}

func buildStatusDescription(status string, reason string) string {
	statusText := map[string]string{
		"PENDING_CONFIRM":  "待确认",
		"WAIT_MEET":        "待面交",
		"COMPLETED":        "已完成",
		"CANCELED":         "已取消",
		"EXCEPTION_CLOSED": "异常关闭",
	}[status]
	if statusText == "" {
		statusText = status
	}
	description := fmt.Sprintf("订单状态更新为%s", statusText)
	if status == "EXCEPTION_CLOSED" {
		description = "异常关闭订单"
	}
	if reason != "" {
		description = fmt.Sprintf("%s：%s", description, reason)
	}
	return description
}

func (s *Service) GetByID(ctx context.Context, orderID uint64) (*Order, error) {
	return s.repo.GetByID(ctx, orderID)
}

func (s *Service) ListByBuyer(ctx context.Context, buyerID uint64, status string, page, pageSize int) ([]OrderDetail, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListByBuyer(ctx, buyerID, status, page, pageSize)
}

func (s *Service) ListBySeller(ctx context.Context, sellerID uint64, status string, page, pageSize int) ([]OrderDetail, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListBySeller(ctx, sellerID, status, page, pageSize)
}

func (s *Service) CancelExpiredOrders(ctx context.Context) (int, error) {
	orders, err := s.repo.ListExpiredOrders(ctx)
	if err != nil {
		return 0, err
	}

	cancelled := 0
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, order := range orders {
		err = db.WithTx(ctx, func(tx *sql.Tx) error {
			if err := s.product.UnlockProduct(ctx, order.ProductID); err != nil {
				return err
			}
			if err := s.repo.UpdateStatus(ctx, tx, order.ID, "CANCELED", map[string]interface{}{
				"cancel_reason": "订单超时未确认，系统自动取消",
				"cancel_by":     order.BuyerID,
				"close_time":    now,
			}); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			continue
		}

		// 发送超时取消消息通知买卖双方
		if s.message != nil {
			relatedType := message.RelatedTypeOrder
			_, _ = s.message.Create(ctx, message.CreateMessageInput{
				ReceiverID:  order.BuyerID,
				MessageType: message.MessageTypeOrderTimeout,
				Title:       "订单超时取消",
				Content:     fmt.Sprintf("订单「%s」因超时未确认，已自动取消。", order.ProductTitleSnapshot),
				RelatedType: &relatedType,
				RelatedID:   &order.ID,
			})
			_, _ = s.message.Create(ctx, message.CreateMessageInput{
				ReceiverID:  order.SellerID,
				MessageType: message.MessageTypeOrderTimeout,
				Title:       "订单超时取消",
				Content:     fmt.Sprintf("订单「%s」因超时未确认，已自动取消。", order.ProductTitleSnapshot),
				RelatedType: &relatedType,
				RelatedID:   &order.ID,
			})
		}

		cancelled++
	}
	return cancelled, nil
}
