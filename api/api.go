package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/errors"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/service"
	"github.com/FactomProject/factom"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	glog "github.com/labstack/gommon/log"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

type API struct {
	HTTP     *echo.Echo
	conf     *config.Config
	service  service.Service
	apiInfo  APIInfo
	validate *validator.Validate
	user     *model.User
}

type APIInfo struct {
	Version string   `json:"version"`
	MW      []string `json:"-"`
}

type ErrorResponse struct {
	Result bool   `json:"result"`
	Code   int    `json:"code"`
	Error  string `json:"error"`
}

type AcceptedResponse struct {
	Result  interface{} `json:"result"`
	Message string      `json:"message"`
}

type SuccessResponse struct {
	Result interface{} `json:"result"`
}

type SuccessResponsePagination struct {
	Result interface{} `json:"result"`
	Start  *int        `json:"start"`
	Limit  *int        `json:"limit"`
	Total  *int        `json:"total"`
}

const (
	Version                = "1.0.0"
	DefaultPaginationStart = 0
	DefaultPaginationLimit = 30
	DefaultSort            = "desc"
	AlternativeSort        = "asc"
)

func NewAPI(conf *config.Config, s service.Service) *API {

	api := &API{}

	api.validate = validator.New()

	api.conf = conf
	api.service = s

	api.HTTP = echo.New()
	api.HTTP.Logger.SetLevel(glog.Lvl(conf.API.LogLevel))
	api.apiInfo.Version = Version
	api.HTTP.HideBanner = true
	api.HTTP.Pre(middleware.RemoveTrailingSlash())

	if conf.API.Logging {
		api.HTTP.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "  API[${status}] ${method} ${uri} (ip=${remote_ip}, latency=${latency_human})\n",
		}))
		api.apiInfo.MW = append(api.apiInfo.MW, "Logger")
	}

	api.HTTP.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		user := api.service.CheckUser(key)
		if user != nil {
			api.user = user
			return true, nil
		}
		err := fmt.Errorf("Invalid auth key: %s", key)
		log.Error(err)
		return false, err
	}))

	api.apiInfo.MW = append(api.apiInfo.MW, "KeyAuth")

	api.HTTP.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: conf.API.GzipLevel,
	}))
	api.apiInfo.MW = append(api.apiInfo.MW, "Gzip")

	// Status
	api.HTTP.GET("/v1", api.index)

	// Chains
	api.HTTP.POST("/v1/chains", api.createChain)
	api.HTTP.GET("/v1/chains", api.getChains)
	api.HTTP.GET("/v1/chains/:chainid", api.getChain)
	api.HTTP.POST("/v1/chains/search", api.searchChains)

	// Chains entries
	api.HTTP.GET("/v1/chains/:chainid/entries", api.getChainEntries)
	api.HTTP.POST("/v1/chains/:chainid/entries/search", api.searchChainEntries)
	api.HTTP.GET("/v1/chains/:chainid/entries/:item", api.getChainFirstOrLastEntry)

	// Entries
	api.HTTP.POST("/v1/entries", api.createEntry)
	api.HTTP.GET("/v1/entries/:entryhash", api.getEntry)

	// User
	api.HTTP.GET("/v1/user", api.getUser)

	// Direct factomd call
	api.HTTP.POST("/v1/factomd/:method", api.factomd)

	return api
}

// Start API server
func (api *API) Start() error {
	return api.HTTP.Start(":" + strconv.Itoa(api.conf.API.HTTPPort))
}

// Returns API information
func (api *API) GetAPIInfo() APIInfo {
	return api.apiInfo
}

// Returns API user info
func (api *API) getUser(c echo.Context) error {
	return c.JSON(http.StatusOK, &api.user)
}

// Returns API info
func (api *API) index(c echo.Context) error {
	return api.SuccessResponse(api.GetAPIInfo(), c)
}

func (api *API) checkUserLimit(action string, c echo.Context) error {

	var usageCost int

	switch action {
	case model.QueueActionChain:
		usageCost = 2
	case model.QueueActionEntry:
		usageCost = 1
	}

	if api.user.UsageLimit != 0 && api.user.UsageLimit-api.user.Usage < usageCost {
		return fmt.Errorf("Writes limit (%d writes) is exceeded for API user '%s'", api.user.UsageLimit, api.user.Name)
	}

	return nil

}

// Success API response
func (api *API) SuccessResponse(res interface{}, c echo.Context) error {
	resp := &SuccessResponse{
		Result: res,
	}
	return c.JSON(http.StatusOK, resp)
}

// Accepted API response
func (api *API) AcceptedResponse(res interface{}, mes string, c echo.Context) error {
	resp := &AcceptedResponse{
		Result:  res,
		Message: mes,
	}
	return c.JSON(http.StatusAccepted, resp)
}

// Success API response with pagination params
func (api *API) SuccessResponsePagination(res interface{}, total int, c echo.Context) error {

	// err should be already checked into API function, so not checking it in response
	start, limit, _, _ := api.GetPaginationParams(c)

	resp := &SuccessResponsePagination{
		Result: res,
		Start:  &start,
		Limit:  &limit,
		Total:  &total,
	}

	return c.JSON(http.StatusOK, resp)
}

// Custom API response in case of error
func (api *API) ErrorResponse(err *errors.Error, c echo.Context) error {
	resp := &ErrorResponse{
		Result: false,
		Code:   err.Code,
		Error:  err.Error(),
	}

	var HTTPResponseCode int

	// factomd error codes will be lt 0
	// error codes from 1400 to 1499 will be lt 0
	// error codes from 1500 will be gte 0
	if err.Code-1500 < 0 {
		HTTPResponseCode = http.StatusBadRequest
	} else {
		HTTPResponseCode = http.StatusInternalServerError
	}

	log.Error(err.Error())
	return c.JSON(HTTPResponseCode, resp)
}

// Helper function: check if pagination params are int;
// returns start, limit, sort;
// number from const is used if param was not provided
func (api *API) GetPaginationParams(c echo.Context) (int, int, string, error) {

	start := DefaultPaginationStart
	limit := DefaultPaginationLimit
	sort := DefaultSort
	var err error

	if c.QueryParam("start") != "" {
		start, err = strconv.Atoi(c.QueryParam("start"))
		if err != nil {
			err = fmt.Errorf("'start' expected to be an integer, '%s' received", c.QueryParam("start"))
			log.Error(err)
			return 0, 0, sort, err
		}
	}

	if c.QueryParam("limit") != "" {
		limit, err = strconv.Atoi(c.QueryParam("limit"))
		if err != nil {
			err = fmt.Errorf("'limit' expected to be an integer, '%s' received", c.QueryParam("limit"))
			log.Error(err)
			return 0, 0, sort, err
		}
	}

	if c.QueryParam("sort") == AlternativeSort {
		sort = AlternativeSort
	}

	return start, limit, sort, nil

}

// API functions

func (api *API) createChain(c echo.Context) error {

	// check user limits
	if err := api.checkUserLimit(model.QueueActionChain, c); err != nil {
		return api.ErrorResponse(errors.New(errors.LimitationError, err), c)
	}

	// Open API Chain struct
	req := &model.Chain{}

	// if JSON request, parse Content from it
	body, err := bodyToJSON(c)
	if err == nil {
		if content, ok := body["content"].(string); ok {
			req.Content = content
		}
	}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	log.Debug("Validating input data")

	// validate ExtIDs, Content
	if err := api.validate.StructExcept(req, "ChainID"); err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	chain, err := api.service.CreateChain(req, api.user)

	if err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	resp := &model.ChainWithLinks{Chain: chain}
	resp.Links = append(resp.Links, model.Link{Rel: "firstEntry", Href: "/entries/" + chain.Base64Decode().FirstEntryHash()})

	return api.SuccessResponse(resp, c)
}

func (api *API) getChains(c echo.Context) error {

	chain := &model.Chain{}

	if c.QueryParam("status") != "" {
		log.Debug("Validating input data")
		chain.Status = c.QueryParam("status")
		// validate Status
		if err := api.validate.StructPartial(chain, "Status"); err != nil {
			return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
		}
	}

	start, limit, sort, err := api.GetPaginationParams(c)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.PaginationError, err), c)
	}

	resp, total := api.service.GetUserChains(chain, api.user, start, limit, sort)

	chains := &model.Chains{Items: resp}

	return api.SuccessResponsePagination(chains.ConvertToChainsWithLinks(), total, c)

}

func (api *API) searchChains(c echo.Context) error {

	// Open API Chain struct
	req := &model.Chain{}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	log.Debug("Validating input data")
	req.Status = c.QueryParam("status")

	// validate ExtIDs
	if err := api.validate.StructPartial(req, "ExtIDs", "Status"); err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	start, limit, sort, err := api.GetPaginationParams(c)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.PaginationError, err), c)
	}

	resp, total := api.service.SearchUserChains(req, api.user, start, limit, sort)

	chains := &model.Chains{Items: resp}

	return api.SuccessResponsePagination(chains.ConvertToChainsWithLinks(), total, c)

}

func (api *API) getChain(c echo.Context) error {

	req := &model.Chain{ChainID: c.Param("chainid")}

	log.Debug("Validating input data")

	// validate ExtIDs, Content
	if err := api.validate.StructPartial(req, "ChainID"); err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	resp, err := api.service.GetChain(req, api.user)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	return api.SuccessResponse(resp.ConvertToChainWithLinks(), c)

}

func (api *API) createEntry(c echo.Context) error {

	// check user limits
	if err := api.checkUserLimit(model.QueueActionEntry, c); err != nil {
		return api.ErrorResponse(errors.New(errors.LimitationError, err), c)
	}

	// Open API Entry struct
	req := &model.Entry{}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	log.Debug("Validating input data")

	// validate ChainID, ExtID (if exists), Content (if exists)
	if err := api.validate.StructExcept(req, "EntryHash"); err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	// Create entry
	resp, err := api.service.CreateEntry(req, api.user)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	return api.SuccessResponse(resp, c)
}

func (api *API) getEntry(c echo.Context) error {

	req := &model.Entry{EntryHash: c.Param("entryhash")}

	log.Debug("Validating input data")

	// validate ExtIDs, Content
	if err := api.validate.StructPartial(req, "EntryHash"); err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	resp, err := api.service.GetEntry(req, api.user)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	return api.SuccessResponse(resp, c)

}

func (api *API) getChainEntries(c echo.Context) error {

	var force bool

	req := &model.Entry{ChainID: c.Param("chainid")}
	req.Status = c.QueryParam("status")

	log.Debug("Validating input data")

	// validate ChainID
	if err := api.validate.StructPartial(req, "ChainID", "Status"); err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	start, limit, sort, err := api.GetPaginationParams(c)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.PaginationError, err), c)
	}

	if c.QueryParam("force") == "true" {
		force = true
	}

	resp, total, err := api.service.GetChainEntries(req, api.user, start, limit, sort, force)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}
	if err == nil && resp == nil {
		return api.AcceptedResponse(resp, "Chain is syncing. Please wait for a while and try again. Or add 'force=true' to request to get partial data.", c)
	}

	return api.SuccessResponsePagination(resp, total, c)

}

func (api *API) searchChainEntries(c echo.Context) error {

	var force bool

	// Open API Entry struct
	req := &model.Entry{}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	if len(req.ExtIDs) == 0 {
		err := fmt.Errorf("Single or multiple 'extIds' are required")
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	req.ChainID = c.Param("chainid")
	req.Status = c.QueryParam("status")

	log.Debug("Validating input data")

	// validate ChainID, ExtID
	if err := api.validate.StructPartial(req, "ChainID", "ExtIDs", "Status"); err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	start, limit, sort, err := api.GetPaginationParams(c)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.PaginationError, err), c)
	}

	if c.QueryParam("force") == "true" {
		force = true
	}

	resp, total, err := api.service.SearchChainEntries(req, api.user, start, limit, sort, force)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}
	if err == nil && resp == nil {
		return api.AcceptedResponse(resp, "Chain is syncing. Please wait for a while and try again. Or add 'force=true' to request to get partial data.", c)
	}

	return api.SuccessResponsePagination(resp, total, c)

}

func (api *API) getChainFirstOrLastEntry(c echo.Context) error {

	log.Debug("Validating first/last item")

	var sort string

	switch c.Param("item") {
	case "first":
		sort = "asc"
	case "last":
		sort = "desc"
	default:
		return api.ErrorResponse(errors.New(errors.ValidationError, fmt.Errorf("Invalid request")), c)
	}

	req := &model.Entry{ChainID: c.Param("chainid")}

	log.Debug("Validating input data")

	// validate ChainID
	if err := api.validate.StructPartial(req, "ChainID"); err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	resp, err := api.service.GetChainFirstOrLastEntry(req, sort, api.user)
	if err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}
	if err == nil && resp == nil {
		return api.AcceptedResponse(resp, "Chain is syncing. Please wait for a while and try again.", c)
	}

	return api.SuccessResponse(resp, c)

}

func (api *API) factomd(c echo.Context) error {

	var params interface{}

	if c.FormValue("params") != "" {
		err := json.Unmarshal([]byte(c.FormValue("params")), &params)
		if err != nil {
			return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
		}
	}

	// if JSON request, parse Content from it
	body, err := bodyToJSON(c)
	if err == nil {
		params = body
	}

	log.Info(params)

	request := factom.NewJSON2Request(c.Param("method"), 0, params)

	resp, err := factom.SendFactomdRequest(request)
	if err != nil {
		return api.ErrorResponse(errors.New(resp.Error.Code, err), c)
	}

	if resp.Error != nil {
		return api.ErrorResponse(errors.New(resp.Error.Code, err), c)
	}

	return api.SuccessResponse(resp.Result, c)

}

func bodyToJSON(c echo.Context) (map[string]interface{}, error) {

	s, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}

	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(s))

	var body map[string]interface{}
	if err := json.Unmarshal(s, &body); err != nil {
		return nil, err
	}

	return body, nil
}
