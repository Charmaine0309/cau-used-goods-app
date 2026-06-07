package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"cau-used-goods-app/backend/internal/admin"
	"cau-used-goods-app/backend/internal/ai"
	"cau-used-goods-app/backend/internal/appeal"
	"cau-used-goods-app/backend/internal/auth"
	"cau-used-goods-app/backend/internal/chat"
	"cau-used-goods-app/backend/internal/config"
	"cau-used-goods-app/backend/internal/db"
	"cau-used-goods-app/backend/internal/favorite"
	"cau-used-goods-app/backend/internal/message"
	"cau-used-goods-app/backend/internal/middleware"
	"cau-used-goods-app/backend/internal/order"
	"cau-used-goods-app/backend/internal/product"
	"cau-used-goods-app/backend/internal/report"
	"cau-used-goods-app/backend/internal/review"
	"cau-used-goods-app/backend/internal/sensitive"
	"cau-used-goods-app/backend/internal/stats"
	"cau-used-goods-app/backend/internal/upload"
	"cau-used-goods-app/backend/internal/user"
)

func main() {
	configPath := os.Getenv("APP_CONFIG")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	if err := db.Init(cfg.Database); err != nil {
		log.Fatalf("init database failed: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database failed: %v", err)
		}
	}()

	log.Printf("database connected: %s:%d/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	authRepo := auth.NewRepository(db.DB())
	authService := auth.NewService(authRepo, cfg.JWT, cfg.Wechat)
	authHandler := auth.NewHandler(authService)

	userRepo := user.NewRepository(db.DB())
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	sensitiveRepo := sensitive.NewRepository(db.DB())
	sensitiveService := sensitive.NewService(sensitiveRepo)

	messageRepo := message.NewRepository(db.DB())
	messageService := message.NewService(messageRepo)
	messageHandler := message.NewHandler(messageService)

	productRepo := product.NewRepository(db.DB())
	productService := product.NewService(productRepo, sensitiveService)
	productHandler := product.NewHandler(productService)

	orderRepo := order.NewRepository(db.DB())
	orderService := order.NewService(orderRepo, productService, messageService)
	orderHandler := order.NewHandler(orderService)

	favoriteRepo := favorite.NewRepository(db.DB())
	favoriteService := favorite.NewService(favoriteRepo)
	favoriteHandler := favorite.NewHandler(favoriteService)

	reviewRepo := review.NewRepository(db.DB())
	reviewService := review.NewService(reviewRepo, messageService)
	reviewHandler := review.NewHandler(reviewService)

	reportRepo := report.NewRepository(db.DB())
	reportService := report.NewService(reportRepo, db.DB(), sensitiveService, messageService)
	reportHandler := report.NewHandler(reportService)

	chatRepo := chat.NewRepository(db.DB())
	chatService := chat.NewService(chatRepo)
	chatHandler := chat.NewHandler(chatService)

	adminRepo := admin.NewRepository(db.DB())
	adminService := admin.NewService(adminRepo)
	adminHandler := admin.NewHandler(adminService)
	sensitiveService.SetAdminLogger(adminService)
	sensitiveHandler := sensitive.NewHandler(sensitiveService)

	appealRepo := appeal.NewRepository(db.DB())
	appealService := appeal.NewService(appealRepo, adminService, messageService)
	appealHandler := appeal.NewHandler(appealService)

	uploadService := upload.NewService()
	uploadHandler := upload.NewHandler(uploadService)

	statsRepo := stats.NewRepository(db.DB())
	statsService := stats.NewService(statsRepo)
	statsHandler := stats.NewHandler(statsService)

	aiService := ai.NewService(cfg.AI.APIKey)
	aiHandler := ai.NewHandler(aiService)

	r := gin.Default()
	r.Static("/uploads", "./uploads")

	authMiddleware := middleware.Auth(cfg.JWT.Secret)
	verifiedMiddleware := middleware.Verified(db.DB())
	adminMiddleware := middleware.Admin(db.DB())

	auth.RegisterRoutes(r, authHandler, authMiddleware, cfg.Server.Env == "dev")
	user.RegisterRoutes(r, userHandler, authMiddleware)
	user.RegisterAdminRoutes(r, userHandler, authMiddleware, adminMiddleware)
	order.RegisterRoutes(r, orderHandler, authMiddleware, verifiedMiddleware, adminMiddleware)
	favorite.RegisterRoutes(r, favoriteHandler, authMiddleware, verifiedMiddleware)
	review.RegisterRoutes(r, reviewHandler, authMiddleware, verifiedMiddleware)
	report.RegisterRoutes(r, reportHandler, authMiddleware, verifiedMiddleware, adminMiddleware)
	message.RegisterRoutes(r, messageHandler, authMiddleware)
	chat.RegisterRoutes(r, chatHandler, authMiddleware, verifiedMiddleware)
	appeal.RegisterRoutes(r, appealHandler, authMiddleware, verifiedMiddleware, adminMiddleware)
	admin.RegisterRoutes(r, adminHandler, authMiddleware, adminMiddleware)
	sensitive.RegisterAdminRoutes(r, sensitiveHandler, authMiddleware, adminMiddleware)

	product.RegisterRoutes(r, productHandler, authMiddleware, adminMiddleware)
	upload.RegisterRoutes(r, uploadHandler, authMiddleware)
	ai.RegisterRoutes(r, aiHandler, authMiddleware)
	stats.RegisterRoutes(r, statsHandler, authMiddleware, adminMiddleware)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}
