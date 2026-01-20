package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

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

	processedFileService, err := services.NewProcessedFileService(db)
	if err != nil {
		log.Fatalf("create processed file service: %v", err)
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
		processedFileService,
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

	dataController, err := controllers.NewDataController(dataService)
	if err != nil {
		log.Fatalf("create data controller: %v", err)
	}

	refreshController, err := controllers.NewRefreshController(pipelineService)
	if err != nil {
		log.Fatalf("create refresh controller: %v", err)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(allowedOrigins()))

	if err := controllers.RegisterHealthRoutes(router); err != nil {
		log.Fatalf("register health routes: %v", err)
	}
	if err := sourcesController.RegisterRoutes(router); err != nil {
		log.Fatalf("register sources routes: %v", err)
	}
	if err := logsController.RegisterRoutes(router); err != nil {
		log.Fatalf("register logs routes: %v", err)
	}
	if err := dataController.RegisterRoutes(router); err != nil {
		log.Fatalf("register data routes: %v", err)
	}
	if err := refreshController.RegisterRoutes(router); err != nil {
		log.Fatalf("register refresh routes: %v", err)
	}
	router.StaticFile("/", "index.html")
	router.StaticFile("/index.html", "index.html")

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

func allowedOrigins() map[string]struct{} {
	originEnv := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if originEnv == "" {
		return map[string]struct{}{
			"http://localhost:3000":         {},
			"https://soldera.pukser.ee":     {},
			"https://soldera-api.pukser.ee": {},
		}
	}

	allowed := map[string]struct{}{}
	for _, part := range strings.Split(originEnv, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		allowed[trimmed] = struct{}{}
	}

	return allowed
}

func corsMiddleware(allowed map[string]struct{}) gin.HandlerFunc {
	if allowed == nil {
		allowed = map[string]struct{}{}
	}

	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		_, ok := allowed[origin]
		if origin != "" && ok {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			ctx.Writer.Header().Set("Vary", "Origin")
			ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if ctx.Request.Method == http.MethodOptions {
			if origin != "" && !ok {
				ctx.AbortWithStatus(http.StatusForbidden)
				return
			}
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
