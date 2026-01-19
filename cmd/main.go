package main

import (
	"context"
	"errors"
	"log"
	"os"

	"solback/cmd/controllers"
	"solback/internal/config"
	"solback/internal/repo"
	"solback/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

const defaultConfigPath = "secrets.json"

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = defaultConfigPath
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := repo.Connect(cfg.DBDSN)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	if err := repo.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	sourceService, err := services.NewSourceService(db)
	if err != nil {
		log.Fatalf("create source service: %v", err)
	}

	logService, err := services.NewLogService(db)
	if err != nil {
		log.Fatalf("create log service: %v", err)
	}

	htmlService, err := services.NewHtmlService(nil)
	if err != nil {
		log.Fatalf("create html service: %v", err)
	}

	openAiService, err := services.NewOpenAiService(cfg.OpenAIAPIKey, logService, nil, "")
	if err != nil {
		log.Fatalf("create openai service: %v", err)
	}

	zipService, err := services.NewZipService(logService, nil)
	if err != nil {
		log.Fatalf("create zip service: %v", err)
	}

	xlsxService, err := services.NewXlsxService()
	if err != nil {
		log.Fatalf("create xlsx service: %v", err)
	}

	csvService, err := services.NewOpenAiCsvService(cfg.OpenAIAPIKey, logService, nil, "")
	if err != nil {
		log.Fatalf("create openai csv service: %v", err)
	}

	dataService, err := services.NewDataService(db, logService)
	if err != nil {
		log.Fatalf("create data service: %v", err)
	}

	pipelineService, err := services.NewPipelineService(
		sourceService,
		htmlService,
		openAiService,
		zipService,
		xlsxService,
		csvService,
		dataService,
		logService,
	)
	if err != nil {
		log.Fatalf("create pipeline service: %v", err)
	}

	sourcesController, err := controllers.NewSourcesController(sourceService)
	if err != nil {
		log.Fatalf("create sources controller: %v", err)
	}

	logsController, err := controllers.NewLogsController(logService)
	if err != nil {
		log.Fatalf("create logs controller: %v", err)
	}

	refreshController, err := controllers.NewRefreshController(pipelineService)
	if err != nil {
		log.Fatalf("create refresh controller: %v", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	if err := controllers.RegisterHealthRoutes(router); err != nil {
		log.Fatalf("register health routes: %v", err)
	}
	if err := sourcesController.RegisterRoutes(router); err != nil {
		log.Fatalf("register sources routes: %v", err)
	}
	if err := logsController.RegisterRoutes(router); err != nil {
		log.Fatalf("register logs routes: %v", err)
	}
	if err := refreshController.RegisterRoutes(router); err != nil {
		log.Fatalf("register refresh routes: %v", err)
	}

	if err := startCron(pipelineService); err != nil {
		log.Fatalf("start cron: %v", err)
	}

	addr := ":8080"
	if err := router.Run(addr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

type pipelineRefresher interface {
	Refresh(ctx context.Context) error
}

func startCron(service pipelineRefresher) error {
	if service == nil {
		return errors.New("html service is nil")
	}

	scheduler := cron.New()

	if _, err := scheduler.AddFunc("@every 1h", func() {
		if err := service.Refresh(context.Background()); err != nil {
			log.Printf("refresh sources: %v", err)
		}
	}); err != nil {
		return err
	}

	scheduler.Start()
	return nil
}
