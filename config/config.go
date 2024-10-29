package config

import (
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/stackvity/aidoc-server/internal/validation" // Import your validation package
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config struct for application configuration
type Config struct {
	Postgres struct {
		User     string `mapstructure:"POSTGRES_USER"`
		Password string `mapstructure:"POSTGRES_PASSWORD"`
		DB       string `mapstructure:"POSTGRES_DB"`
		Port     string `mapstructure:"POSTGRES_PORT"`
		Host     string `mapstructure:"POSTGRES_HOST"`
	} `mapstructure:"Postgres"`
	Clerk struct {
		PublishableKey string `mapstructure:"CLERK_PUBLISHABLE_KEY"`
		SecretKey      string `mapstructure:"CLERK_SECRET_KEY"`
	} `mapstructure:"Clerk"`

	Gemini struct {
		APIKey string `mapstructure:"GEMINI_API_KEY"`
	} `mapstructure:"Gemini"`

	Render struct {
		ExternalURL string `mapstructure:"RENDER_EXTERNAL_URL"`
	} `mapstructure:"Render"`
	App struct {
		GinMode string `mapstructure:"GIN_MODE"`
	} `mapstructure:"App"`

	Sentry struct {
		DSN string `mapstructure:"SENTRY_DSN"`
	} `mapstructure:"Sentry"`
	// Add other config fields as needed
}

var (
	Log      *zap.Logger
	Validate *validator.Validate
)

// LoadConfig loads configuration from .env and environment variables
func LoadConfig() (config Config, err error) {

	if err = InitLogger(); err != nil { // Initialize logger first
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err) // Log to stderr if logger setup fails
		return config, err                                               // Return the error
	}

	viper.AddConfigPath("./")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			Log.Error(".env file not found or error reading it", zap.Error(err)) // Log the error if it's not a file not found error
			return config, fmt.Errorf("failed to read config file: %w", err)
		}
		Log.Warn(".env file not found, using environment variables only.")
	}

	if err = viper.Unmarshal(&config); err != nil {
		Log.Error("Failed to unmarshal config", zap.Error(err)) // Log the error.
		return                                                  // config is the zero value if there's an error
	}

	if err := InitValidator(); err != nil { // Initialize validator after config is loaded. Updated
		Log.Error("validator library initialize error", zap.Error(err)) // Log error
		return config, err                                              // Return error and empty config
	}

	setLogLevel(config) // Call setLogLevel *after* Log is initialized

	return // Return config (and nil error if everything is successful)
}

// InitLogger initializes the Zap logger. Updated.
func InitLogger() error {

	cfg := zap.NewProductionConfig()                          // Use NewProductionConfig for better performance. Updated
	cfg.EncoderConfig.TimeKey = "timestamp"                   // Consistent timestamp key
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // ISO8601 time format
	cfg.Level.SetLevel(zapcore.DebugLevel)                    // Default log level

	Log, err := cfg.Build()
	if err != nil {

		return fmt.Errorf("failed to create logger: %w", err) // Return wrapped error for more context. Updated
	}

	Log.Info("Logger initialized") // Log initialization

	return nil
}

func setLogLevel(cfg Config) {
	var logLevel zapcore.Level

	switch cfg.App.GinMode { // Set log level based on configuration, not gin.Mode() directly. Updated.
	case "release":
		logLevel = zapcore.InfoLevel // Or zap.WarnLevel
	case "test":
		logLevel = zapcore.FatalLevel
	default:
		logLevel = zapcore.DebugLevel // For development and other non-release/test modes.
	}

	if Log != nil { // Check if Log is initialized. Updated.

		Log.Info("Setting log level", zap.String("level", logLevel.String()))
		Log = Log.WithOptions(zap.IncreaseLevel(logLevel))
	} else {
		fmt.Fprintf(os.Stderr, "Log is not initialized. Cannot set log level.\n") // Log to stderr since logger isn't available. Updated
	}
}

func InitValidator() error {
	Validate = validator.New()

	// Register custom validators
	if err := Validate.RegisterValidation("pastdate", validation.PastDateValidator); err != nil { // Register custom validator for past dates
		return fmt.Errorf("register pastdate validator err: %w", err)
	}
	if err := Validate.RegisterValidation("dateformat", validation.DateFormatValidator); err != nil { // register custom validator for date format
		return fmt.Errorf("register dateformat validator err: %w", err)
	}
	if err := Validate.RegisterValidation("minage", validation.MinimumAgeValidator); err != nil { // Register custom validator for minimum age
		return fmt.Errorf("register minage validator err: %w", err)
	}

	if err := Validate.RegisterValidation("phoneNumber", validation.PhoneNumberValidator); err != nil { // Corrected name
		return fmt.Errorf("register phoneNumber validator err: %w", err)
	}

	if err := Validate.RegisterValidation("oneof", validation.OneOfValidator); err != nil { // Corrected the validator function name
		return fmt.Errorf("register oneof validator err: %w", err)
	}
	return nil

}

// DBURL builds the database connection URL.
func DBURL(cfg Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DB)
}

// InitSentry initializes Sentry for error tracking.
func InitSentry(cfg Config) {
	if cfg.Sentry.DSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.Sentry.DSN,
			TracesSampleRate: 1.0, // Adjust as needed
			Environment:      gin.Mode(),
		})
		if err != nil {
			Log.Fatal("Sentry initialization failed:", zap.Error(err))
		}

		defer sentry.Flush(2 * time.Second) // Flush buffered events before the program terminates.
	}
}

// SetupCORSMiddleware sets up CORS middleware with Gin.
func SetupCORSMiddleware(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Update in production with your allowed origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}
