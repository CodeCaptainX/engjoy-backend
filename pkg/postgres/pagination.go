package postgres

import (
	"fmt"

	"sentenceminer/pkg/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type PaginationOptions struct {
	PerPage int `json:"perPage" query:"per_page" validate:"required,min=1"`
	Page    int `json:"page" query:"page" validate:"required,min=1"`
}

type Filter struct {
	Property string      `json:"property" validate:"required"`
	Value    interface{} `json:"value" validate:"required"`
}

type Sort struct {
	Property  string `json:"property" validate:"required"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"`
}

type QueryParamRequest struct {
	PagingOptions PaginationOptions `json:"pagingOptions" query:"paging_options" validate:"required"`
	Filters       []Filter          `json:"filters" query:"filters"`
	Sorts         []Sort            `json:"sorts" query:"sorts"`
	Offset        int               `json:"offset" query:"offset"`
}

type ResponseWithPagination struct {
	Success       bool              `json:"success"`
	Message       string            `json:"message"`
	StatusCode    int               `json:"statusCode"`
	Data          interface{}       `json:"data"`
	PagingOptions PaginationOptions `json:"pagingOptions"`
	Total         int               `json:"total"`
}

func parsePaginationOptions(c *fiber.Ctx) PaginationOptions {
	perPage, _ := strconv.Atoi(c.Query("paging_options[per_page]", "10"))
	page, _ := strconv.Atoi(c.Query("paging_options[page]", "1"))
	return PaginationOptions{
		PerPage: perPage,
		Page:    page,
	}
}

func parseFilters(c *fiber.Ctx) []Filter {
	var filters []Filter

	for i := 0; ; i++ {
		property := c.Query(fmt.Sprintf("filters[%d][property]", i))
		if property == "" {
			break
		}

		// Get value as string
		valueStr := c.Query(fmt.Sprintf("filters[%d][value]", i))
		var val interface{}

		if valueStr != "" {
			// If value contains comma, convert to slice
			parts := strings.Split(valueStr, ",")
			if len(parts) == 1 {
				val = parts[0] // single value
			} else {
				slice := make([]interface{}, len(parts))
				for j, v := range parts {
					slice[j] = v
				}
				val = slice // array for BETWEEN
			}
		}

		filters = append(filters, Filter{
			Property: property,
			Value:    val,
		})
	}

	return filters
}

func parseSorts(c *fiber.Ctx) []Sort {
	var sorts []Sort
	for i := 0; ; i++ {
		property := c.Query(fmt.Sprintf("sorts[%d][property]", i))
		direction := c.Query(fmt.Sprintf("sorts[%d][direction]", i))
		if property == "" || direction == "" {
			break
		}
		sorts = append(sorts, Sort{
			Property:  property,
			Direction: direction,
		})
	}
	return sorts
}

// GenerateSortQuery constructs a SQL ORDER BY clause from the provided sorts and table alias.
func GenerateSortQuery(sorts []Sort, tableAlias string) string {
	var sortQuery string
	if len(sorts) > 0 {
		sortQuery = " ORDER BY "
		for i, sort := range sorts {
			// Use the provided dynamic table alias in the sort clause
			sortQuery += fmt.Sprintf("%s.%s %s", tableAlias, sort.Property, sort.Direction)
			if i < len(sorts)-1 {
				sortQuery += ", "
			}
		}
	}
	return sortQuery
}

// GenerateFilterQuery constructs a SQL filter query string based on the provided filters and table alias.
func GenerateFilterQuery(filter []Filter, tableAlias string) string {
	var sortQuery string
	if len(filter) > 0 {
		sortQuery = " AND "
		for i, sort := range filter {
			// Use the provided dynamic table alias in the sort clause
			sortQuery += fmt.Sprintf("%s.%s %s %s", tableAlias, sort.Property, "=", sort.Value)
			if i < len(filter)-1 {
				sortQuery += " AND "
			}
		}
	}
	return sortQuery
}

func parseQueryParams(c *fiber.Ctx) QueryParamRequest {
	return QueryParamRequest{
		PagingOptions: parsePaginationOptions(c),
		Filters:       parseFilters(c),
		Sorts:         parseSorts(c),
	}
}

func ExtractQueryParamsRequest(c *fiber.Ctx) (*QueryParamRequest, error) {
	var queryParamRequest QueryParamRequest

	v := utils.NewCustomValidator()
	// //TODO: Validate Struct
	err := v.Bind(c, &queryParamRequest)
	if err != nil {
		return nil, err
	}

	queryParams := parseQueryParams(c)
	page := queryParams.PagingOptions.Page

	if page <= 0 {
		page = 1
	}
	perPage := queryParams.PagingOptions.PerPage
	if perPage <= 0 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	queryParamRequest.PagingOptions.Page = page
	queryParamRequest.PagingOptions.PerPage = perPage
	queryParamRequest.Offset = offset

	queryParamRequest.Filters = queryParams.Filters
	queryParamRequest.Sorts = queryParams.Sorts

	return &queryParamRequest, nil
}

// ValidateFilterColumns checks if the provided filter properties are valid based on the allowed columns and returns an error if any invalid property is found.
func ValidateFilterColumns(filters []Filter, validColumns map[string]bool) error {
	// Filters
	for _, filter := range filters {
		if !validColumns[filter.Property] {
			return fmt.Errorf("invalid filter property: %s", filter.Property)
		}
	}
	return nil
}

func ValidateSortColumns(sorts []Sort, validColumns map[string]bool) error {
	fmt.Println("🚀 ~ file: pagination.go ~ line 169 ~ funcValidateSortColumns ~ sorts : ", sorts)
	// Filters
	fmt.Println("🚀 ~ file: pagination.go ~ line 176 ~ funcValidateSortColumns ~ sort.Property : ", sorts[0].Property)
	for _, sort := range sorts {
		if !validColumns[sort.Property] {
			return fmt.Errorf("invalid sort property: %s", sort.Property)
		}
	}
	return nil
}

func getTableAliasForProperty(property string) string {
	// Define a map that associates property names or prefixes with table aliases
	propertyToAliasMap := map[string]string{
		"language": "l", // All properties starting with "language" belong to the "l" alias (tbl_languages)
		// Add other property prefixes or specific properties as needed
		"currency": "c", // Example for currency-related properties if you join with a currency table
		"user":     "u", // All properties starting with "user" belong to the "u" alias (tbl_users)
	}

	// Default to the "u" alias (tbl_users) if no match is found
	alias := "u"

	// Check if the property starts with any key in the map and return the corresponding alias
	for prefix, tableAlias := range propertyToAliasMap {
		if strings.HasPrefix(property, prefix) {
			alias = tableAlias
			break
		}
	}

	return fmt.Sprintf("%s.%s", alias, property)
}

func GenerateFilteredQuery(baseQuery string, extraQuery *QueryParamRequest) (string, string, []interface{}, []interface{}) {
	trimmedBaseQuery := strings.TrimSpace(baseQuery)

	fromIndex := strings.Index(trimmedBaseQuery, "FROM")
	if fromIndex == -1 {
		log.Error().Msg("Invalid query: no FROM clause found")
		return baseQuery, baseQuery, nil, nil
	}

	countQuery := "SELECT COUNT(*) " + trimmedBaseQuery[fromIndex:]

	var filterConditions []string
	var args []interface{}
	var countArgs []interface{}

	// Apply filters with dynamic alias determination
	for _, filter := range extraQuery.Filters {
		filteredProperty := getTableAliasForProperty(filter.Property)
		filterConditions = append(filterConditions, fmt.Sprintf("%s = $%d", filteredProperty, len(args)+1))
		args = append(args, filter.Value)
		countArgs = append(countArgs, filter.Value)
	}

	if len(filterConditions) > 0 {
		condition := " AND " + strings.Join(filterConditions, " AND ")
		baseQuery += condition
		countQuery += condition
	}

	// Apply sorting with dynamic alias determination
	if len(extraQuery.Sorts) > 0 {
		var sortClauses []string
		for _, sort := range extraQuery.Sorts {
			sortedProperty := getTableAliasForProperty(sort.Property)
			sortClauses = append(sortClauses, fmt.Sprintf("%s %s", sortedProperty, sort.Direction))
		}
		baseQuery += " ORDER BY " + strings.Join(sortClauses, ", ")
	}

	// Apply pagination
	perPage := extraQuery.PagingOptions.PerPage
	if perPage <= 0 {
		perPage = 10
	}
	offset := extraQuery.Offset
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, perPage, offset)

	return baseQuery, countQuery, args, countArgs
}

// APIResponseWithPagination constructs a paginated API response with the provided success status, message, status code, data, pagination options, and total count.
func APIResponseWithPagination(
	success bool,
	message string,
	statusCode int,
	data interface{},
	pagingOptions PaginationOptions,
	total int,
) *ResponseWithPagination {
	return &ResponseWithPagination{
		Success:       success,
		Message:       message,
		StatusCode:    statusCode,
		Data:          data,
		PagingOptions: pagingOptions,
		Total:         total,
	}
}

type ResponseError struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}

func APIResponseError(
	success bool,
	message string,
	statusCode int,
	data interface{},
) *ResponseError {
	return &ResponseError{
		Success:    success,
		Message:    message,
		StatusCode: statusCode,
		Data:       data,
	}
}

// func NewErrorBodyParserResponse(c *fiber.Ctx) error {
// 	return c.Status(fiber.StatusBadRequest).JSON(APIResponseError(
// 		constant.IsFail,
// 		constant.SmsErrJsonParse,
// 		465,
// 		nil,
// 	))
// }
