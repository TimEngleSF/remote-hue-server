package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/TimEngleSF/remote-hue-server/internal/service"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	goopenai "github.com/sashabaranov/go-openai"
	"github.com/twilio/twilio-go"
)

const version = "1.0.0"

type config struct {
	port              int
	env               string
	userPhoneNumber   string
	auth_token        string
	twilioPhoneNumber string
}

type application struct {
	config       config
	logger       *slog.Logger
	twilio       *twilio.RestClient
	openai       *service.OpenaiService
	wsConnection *websocket.Conn
	groupsState  *service.Groups
	groupNames   service.GroupNames
	responseMap  map[string]chan JSONMessage
	responseMu   sync.Mutex
}

func main() {
	var cfg config
	var useEnvFile bool

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.userPhoneNumber, "userPhoneNumber", "", "User phone number: '+19875551234'")
	flag.StringVar(&cfg.twilioPhoneNumber, "twilioPhoneNumber", "", "Twilio phone number: '+19875551234'")
	flag.StringVar(&cfg.auth_token, "auth_token", "password", "Authentication token for home client")
	flag.BoolVar(&useEnvFile, "envFile", false, "Use .env file for environment variables")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if useEnvFile {
		err := godotenv.Load()
		if err != nil {
			logger.Error("Error loading .env file")
		}
	}

	// If the userPhoneNumber flag is not set, check the environment
	if cfg.userPhoneNumber == "" {
		cfg.userPhoneNumber = os.Getenv("USER_PHONE_NUMBER")
		// check if phone number is the correct format '+19875551234' using regex

		re := regexp.MustCompile(`^\+[1-9]\d{10,14}$`)
		if !re.MatchString(cfg.userPhoneNumber) {
			logger.Error("User phone number is not in the correct format", "phone_number", cfg.userPhoneNumber)
			os.Exit(1)
		}

		if cfg.userPhoneNumber == "" {
			logger.Error("User phone number is not set")
			os.Exit(1)
		}
	}

	// If the twilioPhoneNumber flag is not set, check the environment
	if cfg.twilioPhoneNumber == "" {
		cfg.twilioPhoneNumber = os.Getenv("TWILIO_PHONE_NUMBER")
		// check if phone number is the correct format '+19875551234' using regex
		re := regexp.MustCompile(`^\+[1-9]\d{10,14}$`)
		if !re.MatchString(cfg.twilioPhoneNumber) {
			logger.Error("Twilio phone number is not in the correct format", "phone_number", cfg.twilioPhoneNumber)
			os.Exit(1)
		}
		if cfg.twilioPhoneNumber == "" {
			logger.Error("Twilio phone number is not set")
			os.Exit(1)
		}
	}

	for _, envVar := range []string{"TWILIO_ACCOUNT_SID", "TWILIO_AUTH_TOKEN", "OPENAI_API_KEY"} {
		if os.Getenv(envVar) == "" {
			logger.Error(fmt.Sprintf("Environment variable %s is not set", envVar))
			os.Exit(1)
		}
	}
	// Initialize the Twilio client
	var twilioClient *twilio.RestClient
	twilioUsername := os.Getenv("TWILIO_ACCOUNT_SID")
	twilioPassword := os.Getenv("TWILIO_AUTH_TOKEN")
	if useEnvFile {
		twilioParams := twilio.ClientParams{Username: twilioUsername, Password: twilioPassword}
		twilioClient = twilio.NewRestClientWithParams(twilioParams)
	} else {
		twilioClient = twilio.NewRestClient()
	}

	// Initialize openai client
	openaiKey := os.Getenv("OPENAI_API_KEY")
	openaiClient := goopenai.NewClient(openaiKey)
	openaiService := service.OpenaiService{Client: openaiClient}

	// Application struct
	app := &application{
		config: cfg,
		logger: logger,
		twilio: twilioClient,
		openai: &openaiService,
	}

	svr := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("Starting server", "port", cfg.port)
	err := svr.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
