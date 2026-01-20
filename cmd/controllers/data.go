package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"solback/internal/models"
	"solback/internal/services"

	"github.com/gin-gonic/gin"
)

type DataProvider interface {
	GetData(ctx context.Context, period string, technology string, groupPeriod string, sumTech bool, from string, to string, techIn string, sort string, limit string) ([]models.AuctionResult, error)
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
	groupPeriod := ctx.Query("group_period")
	from := ctx.Query("from")
	to := ctx.Query("to")
	techIn := ctx.Query("tech_in")
	sort := ctx.Query("sort")
	limit := ctx.Query("limit")
	technology := ctx.Query("tech")
	if technology == "" {
		technology = ctx.Query("technology")
	}
	sumTechParam := ctx.Query("sum_tech")
	sumTech := false
	if sumTechParam != "" {
		parsed, err := strconv.ParseBool(sumTechParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid sum_tech"})
			return
		}
		sumTech = parsed
	}

	results, err := c.service.GetData(ctx.Request.Context(), period, technology, groupPeriod, sumTech, from, to, techIn, sort, limit)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPeriod) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid period"})
			return
		}
		if errors.Is(err, services.ErrInvalidGroupPeriod) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid group_period"})
			return
		}
		if errors.Is(err, services.ErrInvalidMonthRange) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid month range"})
			return
		}
		if errors.Is(err, services.ErrInvalidSort) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid sort"})
			return
		}
		if errors.Is(err, services.ErrInvalidLimit) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid limit"})
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
