package main

import (
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

	sourcesController, err := controllers.NewSourcesController(sourceService)
	if err != nil {
		log.Fatalf("create sources controller: %v", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	if err := controllers.RegisterHealthRoutes(router); err != nil {
		log.Fatalf("register health routes: %v", err)
	}
	if err := sourcesController.RegisterRoutes(router); err != nil {
		log.Fatalf("register sources routes: %v", err)
	}

	if err := startCron(); err != nil {
		log.Fatalf("start cron: %v", err)
	}

	addr := ":8080"
	if err := router.Run(addr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

func startCron() error {
	scheduler := cron.New()

	if _, err := scheduler.AddFunc("@every 1h", func() {
		log.Println("hello")
	}); err != nil {
		return err
	}

	scheduler.Start()
	return nil
}
