package controllers

import (
	"context"
	"errors"
	"net/http"

	"solback/internal/models"
	"solback/internal/services"

	"github.com/gin-gonic/gin"
)

type DataProvider interface {
	GetData(ctx context.Context, period string, technology string) ([]models.AuctionResult, error)
	DeleteData(ctx context.Context) (int, error)
}

type DataController struct {
	service DataProvider
}

type DeleteDataResponse struct {
	Deleted int `json:"deleted"`
}

func NewDataController(service DataProvider) (*DataController, error) {
	if service == nil {
		return nil, errors.New("data service is nil")
	}

	return &DataController{service: service}, nil
}

func (c *DataController) RegisterRoutes(router *gin.Engine) error {
	if c == nil {
		return errors.New("data controller is nil")
	}
	if router == nil {
		return errors.New("router is nil")
	}

	router.GET("/data", c.getData)
	router.DELETE("/data", c.deleteData)
	return nil
}

func (c *DataController) getData(ctx *gin.Context) {
	period := ctx.Query("period")
	technology := ctx.Query("tech")
	if technology == "" {
		technology = ctx.Query("technology")
	}

	results, err := c.service.GetData(ctx.Request.Context(), period, technology)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPeriod) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid period"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load data"})
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (c *DataController) deleteData(ctx *gin.Context) {
	deleted, err := c.service.DeleteData(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete data"})
		return
	}

	ctx.JSON(http.StatusOK, DeleteDataResponse{Deleted: deleted})
}
