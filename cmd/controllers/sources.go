package controllers

import (
	"context"
	"errors"
	"net/http"

	"solback/internal/models"

	"github.com/gin-gonic/gin"
)

type SourceProvider interface {
	GetSources(ctx context.Context) ([]models.Source, error)
}

type SourcesController struct {
	service SourceProvider
}

type SourcesResponse struct {
	Sources []models.Source `json:"sources"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewSourcesController(service SourceProvider) (*SourcesController, error) {
	if service == nil {
		return nil, errors.New("source service is nil")
	}

	return &SourcesController{service: service}, nil
}

func (c *SourcesController) RegisterRoutes(router *gin.Engine) error {
	if c == nil {
		return errors.New("sources controller is nil")
	}
	if router == nil {
		return errors.New("router is nil")
	}

	router.GET("/sources", c.getSources)
	return nil
}

func (c *SourcesController) getSources(ctx *gin.Context) {
	sources, err := c.service.GetSources(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load sources"})
		return
	}

	ctx.JSON(http.StatusOK, SourcesResponse{Sources: sources})
}
