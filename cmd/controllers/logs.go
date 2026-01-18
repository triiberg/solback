package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"solback/internal/models"

	"github.com/gin-gonic/gin"
)

const defaultLogsLimit = 20

type LogProvider interface {
	GetLogs(ctx context.Context, limit int) ([]models.Log, error)
}

type LogsController struct {
	service LogProvider
}

func NewLogsController(service LogProvider) (*LogsController, error) {
	if service == nil {
		return nil, errors.New("log service is nil")
	}

	return &LogsController{service: service}, nil
}

func (c *LogsController) RegisterRoutes(router *gin.Engine) error {
	if c == nil {
		return errors.New("logs controller is nil")
	}
	if router == nil {
		return errors.New("router is nil")
	}

	router.GET("/logs", c.getLogs)
	return nil
}

func (c *LogsController) getLogs(ctx *gin.Context) {
	limit, err := parseLogsLimit(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid logs limit"})
		return
	}

	logs, err := c.service.GetLogs(ctx.Request.Context(), limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load logs"})
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func parseLogsLimit(ctx *gin.Context) (int, error) {
	value := ctx.Query("n")
	if value == "" {
		return defaultLogsLimit, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if limit <= 0 {
		return 0, errors.New("limit must be positive")
	}

	return limit, nil
}
