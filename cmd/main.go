package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stackvity/aidoc-server/api/handler"
	"github.com/stackvity/aidoc-server/api/middleware"
	"github.com/stackvity/aidoc-server/bootstrap"
	"github.com/stackvity/aidoc-server/config"
	"github.com/stackvity/aidoc-server/internal/auth"
	"github.com/stackvity/aidoc-server/internal/core/service"
	"github.com/stackvity/aidoc-server/internal/platform/repository/postgres"
	db "github.com/stackvity/aidoc-server/internal/platform/repository/sqlc"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	if err := config.InitLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		return
	}
	defer config.Log.Sync()

	if err := config.InitValidator(); err != nil {
		config.Log.Fatal("failed to initialize validator:", zap.Error(err))
	}

	dbPool, err := bootstrap.ConnectDB(cfg, config.Log)
	if err != nil {
		config.Log.Fatal("failed to connect to database:", zap.Error(err))
	}
	defer dbPool.Close()

	config.InitSentry(cfg)

	// Set the Clerk API key directly (for single-instance usage)
	clerk.SetKey(cfg.Clerk.SecretKey)

	// Create context for use in Clerk API calls
	ctx := context.Background()

	// Initialize the authentication client.
	authClient, err := auth.NewAuthClient(ctx, config.Log)
	if err != nil {
		config.Log.Fatal("failed initialize auth client", zap.Error(err))
	}

	// Initialize repositories.
	queries := db.New(dbPool)
	patientRepo := postgres.NewPatientRepository(queries, config.Log)
	lifestyleRepo := postgres.NewLifestyleRepository(queries, config.Log)
	medicalHistoryRepo := postgres.NewMedicalHistoryRepository(queries, config.Log)

	// Initialize services.
	patientService := service.NewPatientService(patientRepo, config.Log, config.Validate)
	lifestyleService := service.NewLifestyleService(lifestyleRepo, patientRepo, config.Log, config.Validate, authClient.Authorize)
	medicalHistoryService := service.NewMedicalHistoryService(medicalHistoryRepo, patientRepo, config.Log, config.Validate, authClient.Authorize)

	// Initialize handlers.
	patientHandler := handler.NewPatientHandler(patientService, config.Log)
	lifestyleHandler := handler.NewLifestyleHandler(lifestyleService, config.Log)
	medicalHistoryHandler := handler.NewMedicalHistoryHandler(medicalHistoryService, config.Log)

	router := gin.Default()

	// CORS configuration.  **Important:** In production, restrict AllowOrigins.
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Authorization", "Content-Type"}
	router.Use(cors.New(corsConfig))

	// Authentication middleware.
	authMiddleware := middleware.AuthMiddleware(config.Log)

	// Route definitions and middleware application.
	v1 := router.Group("/v1")
	{
		patients := v1.Group("/patients")
		patients.Use(authMiddleware)
		{
			patients.POST("/", middleware.RequirePermissions([]string{"patient:create"}, config.Log), patientHandler.CreatePatient)
			patients.GET("/:patient_id", middleware.RequirePermissions([]string{"patient:read"}, config.Log), patientHandler.GetPatient)
			patients.PUT("/:patient_id", middleware.RequirePermissions([]string{"patient:update"}, config.Log), patientHandler.UpdatePatient)

			medicalHistory := patients.Group("/:patient_id/medical_history")
			medicalHistory.Use(authMiddleware)
			{
				medicalHistory.POST("/", middleware.RequirePermissions([]string{"medical_history:create"}, config.Log), medicalHistoryHandler.CreateMedicalHistoryEntry)
				medicalHistory.GET("/", middleware.RequirePermissions([]string{"medical_history:read"}, config.Log), medicalHistoryHandler.GetMedicalHistoryEntries)
				medicalHistory.PUT("/:medical_history_id", middleware.RequirePermissions([]string{"medical_history:update"}, config.Log), medicalHistoryHandler.UpdateMedicalHistoryEntry)
				medicalHistory.DELETE("/:medical_history_id", middleware.RequirePermissions([]string{"medical_history:delete"}, config.Log), medicalHistoryHandler.DeleteMedicalHistoryEntry)
			}

			lifestyle := patients.Group("/:patient_id/lifestyle")
			lifestyle.Use(authMiddleware)
			{
				lifestyle.POST("/", middleware.RequirePermissions([]string{"lifestyle:create"}, config.Log), lifestyleHandler.CreateLifestyleEntry)
				lifestyle.GET("/", middleware.RequirePermissions([]string{"lifestyle:read"}, config.Log), lifestyleHandler.GetLifestyleEntries)
				lifestyle.PUT("/:lifestyle_id", middleware.RequirePermissions([]string{"lifestyle:update"}, config.Log), lifestyleHandler.UpdateLifestyleEntry)
				lifestyle.DELETE("/:lifestyle_id", middleware.RequirePermissions([]string{"lifestyle:delete"}, config.Log), lifestyleHandler.DeleteLifestyleEntry)
			}
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.App.GinMode,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) { // Corrected error handling
			config.Log.Error("listen: %s\n", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		config.Log.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	fmt.Println("Server exiting")

}
