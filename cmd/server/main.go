package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	goopenai "github.com/sashabaranov/go-openai"
	"github.com/twilio/twilio-go"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

type application struct {
	config config
	logger *slog.Logger
	twilio *twilio.RestClient
	openai *goopenai.Client
}

func main() {
	var cfg config
	var useEnvFile bool

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.BoolVar(&useEnvFile, "envFile", false, "Use .env file for environment variables")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if useEnvFile {
		err := godotenv.Load()
		if err != nil {
			logger.Error("Error loading .env file")
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

	app := &application{
		config: cfg,
		logger: logger,
		twilio: twilioClient,
		openai: openaiClient,
	}

	fmt.Println(app)
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
