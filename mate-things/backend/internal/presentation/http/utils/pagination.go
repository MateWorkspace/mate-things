package presentationhttputils

import (
	"strconv"
	"strings"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	presentationhttpresponse "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/response"
	"github.com/labstack/echo/v5"
)

const (
	defaultPage  = 1
	defaultLimit = 10
)

type PaginationArgs struct {
	Page   int
	Limit  int
	Search *string
}

func PageArgs(c *echo.Context) (PaginationArgs, error) {
	page, err := queryPositiveInt(c, "page", defaultPage)
	if err != nil {
		return PaginationArgs{}, err
	}

	limit, err := queryPositiveInt(c, "limit", defaultLimit)
	if err != nil {
		return PaginationArgs{}, err
	}

	return PaginationArgs{
		Page:   page,
		Limit:  limit,
		Search: QueryString(c, "search"),
	}, nil
}

func PageResponse(args PaginationArgs, total int) presentationhttpresponse.PageResponse {
	return presentationhttpresponse.PageResponse{
		Page:       args.Page,
		Limit:      args.Limit,
		TotalItems: total,
	}
}

func queryPositiveInt(c *echo.Context, field string, defaultValue int) (int, error) {
	value := strings.TrimSpace(c.QueryParam(field))
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, domainmodels.NewError(field+" must be a valid integer", domainmodels.ErrTypeValidation, err)
	}
	if parsed <= 0 {
		return defaultValue, nil
	}

	return parsed, nil
}
