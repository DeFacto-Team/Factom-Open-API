package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/errors"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/service"
	"github.com/DeFacto-Team/Factom-Open-API/webpack"
	"github.com/FactomProject/factom"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/swaggo/echo-swagger"
	_ "github.com/swaggo/echo-swagger/example/docs"
	"gopkg.in/go-playground/validator.v9"
)

type API struct {
	HTTP       *echo.Echo
	conf       *config.Config
	service    service.Service
	configFile string
	apiInfo    APIInfo
	validate   *validator.Validate
	user       *model.User
	jwtSecret  []byte
}

type APIInfo struct {
	Version string   `json:"version"`
	Port    int      `json:"-"`
	MW      []string `json:"-"`
}

type ErrorResponse struct {
	Result bool   `json:"result" default:"false"`
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

type ViewData struct {
	assetsMapper webpack.AssetsMapper
}

const (
	Version                = "1.1.0"
	DefaultPaginationStart = 0
	DefaultPaginationLimit = 30
	DefaultSort            = "desc"
	AlternativeSort        = "asc"
	AccessTokenLength      = 32
)

// NewViewData creates new data for the view
func NewViewData(buildPath string) (ViewData, error) {
	assetsMapper, err := webpack.NewAssetsMapper(buildPath)
	if err != nil {
		return ViewData{}, err
	}

	return ViewData{
		assetsMapper: assetsMapper,
	}, nil
}

// Webpack maps file name to path
func (d ViewData) Webpack(file string) string {
	return d.assetsMapper(file)
}

func NewAPI(conf *config.Config, s service.Service, configFile string) *API {

	api := &API{}

	api.validate = validator.New()

	api.conf = conf
	api.service = s
	api.configFile = configFile

	api.HTTP = echo.New()
	api.HTTP.HideBanner = true
	api.HTTP.HidePort = true
	api.apiInfo.Version = Version
	api.apiInfo.Port = api.conf.API.HTTPPort
	api.HTTP.Pre(middleware.RemoveTrailingSlash())

	if conf.API.Logging {
		api.HTTP.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "  API[${status}] ${method} ${uri} (ip=${remote_ip}, latency=${latency_human})\n",
		}))
		api.apiInfo.MW = append(api.apiInfo.MW, "Logger")
	}

	api.HTTP.Use(middleware.Recover())
	api.apiInfo.MW = append(api.apiInfo.MW, "Recover")

	authGroup := api.HTTP.Group("/v1")
	authGroup.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
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

	adminGroup := api.HTTP.Group("/admin")
	api.jwtSecret = generateJWTSecret()
	adminGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  api.jwtSecret,
		TokenLookup: "cookie:token",
	}))
	api.apiInfo.MW = append(api.apiInfo.MW, "JWT")

	// Login endpoint
	api.HTTP.POST("/login", api.login)

	// Admin UI
	api.HTTP.File("/", "ui/build/index.html")
	api.HTTP.File("/manifest.json", "ui/build/manifest.json")
	api.HTTP.File("/favicon.ico", "ui/build/favicon.ico")
	api.HTTP.Static("/static", "ui/build/static")

	// Admin endpoints
	adminGroup.GET("", api.adminIndex)
	adminGroup.GET("/queue", api.adminGetQueue)
	adminGroup.DELETE("/queue", api.adminDeleteQueue)
	adminGroup.GET("/users", api.adminGetUsers)
	adminGroup.POST("/users", api.adminCreateUser)
	adminGroup.DELETE("/users", api.adminDeleteUser)
	adminGroup.PUT("/users/:id", api.adminUpdateUser)
	adminGroup.POST("/users/rotate", api.adminRotateUserToken)
	adminGroup.GET("/logout", api.adminLogout)
	adminGroup.GET("/settings", api.adminGetSettings)
	adminGroup.POST("/settings", api.adminUpdateSettings)
	adminGroup.GET("/restart", api.adminRestartAPI)

	// Status
	api.HTTP.GET("/v1", api.index)

	// Documentation
	url := echoSwagger.URL("swagger.json")
	api.HTTP.Static("/docs/swagger.json", "./docs/swagger.json")
	api.HTTP.GET("/docs/*", echoSwagger.EchoWrapHandler(url))

	// Chains
	authGroup.POST("/chains", api.createChain)
	authGroup.GET("/chains", api.getChains)
	authGroup.GET("/chains/:chainid", api.getChain)
	authGroup.POST("/chains/search", api.searchChains)

	// Chains entries
	authGroup.GET("/chains/:chainid/entries", api.getChainEntries)
	authGroup.POST("/chains/:chainid/entries/search", api.searchChainEntries)
	authGroup.GET("/chains/:chainid/entries/:item", api.getChainFirstOrLastEntry)

	// Entries
	authGroup.POST("/entries", api.createEntry)
	authGroup.GET("/entries/:entryhash", api.getEntry)

	// User
	authGroup.GET("/user", api.getUser)

	// Direct factomd call
	authGroup.POST("/factomd/:method", api.factomd)

	return api
}

// Start API server
func (api *API) Start() error {
	return api.HTTP.Start(":" + strconv.Itoa(api.conf.API.HTTPPort))
}

// Stop API server
func (api *API) Stop() error {
	return api.HTTP.Close()
}

// Returns API information
func (api *API) GetAPIInfo() APIInfo {
	return api.apiInfo
}

func (api *API) login(c echo.Context) error {
	user := c.FormValue("user")
	password := c.FormValue("password")

	// Check admin auth
	if user == api.conf.Admin.User && password == api.conf.Admin.Password {

		// Create token
		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims
		claims := token.Claims.(jwt.MapClaims)
		claims["admin"] = true
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

		// Generate encoded token and send it as response.
		t, err := token.SignedString(api.jwtSecret)
		if err != nil {
			return err
		}

		writeCookie(c, t)

		return c.JSON(http.StatusOK, map[string]string{
			"token": t,
		})
	}

	return echo.ErrUnauthorized

}

func (api *API) adminLogout(c echo.Context) error {

	deleteCookie(c)

	return c.JSON(http.StatusOK, map[string]bool{
		"ok": true,
	})

}

func (api *API) adminGetSettings(c echo.Context) error {

	return c.JSON(http.StatusOK, api.conf)

}

func (api *API) adminUpdateSettings(c echo.Context) error {

	newConf := &config.Config{}
	copier.Copy(newConf, api.conf)

	if err := c.Bind(newConf); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	if err := config.UpdateConfig(api.configFile, newConf); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	return c.JSON(http.StatusOK, map[string]bool{
		"ok": true,
	})

}

func (api *API) adminRestartAPI(c echo.Context) error {

	api.Stop()

	return c.JSON(http.StatusOK, map[string]bool{
		"ok": true,
	})

}

func (api *API) adminIndex(c echo.Context) error {

	return api.SuccessResponse(api.GetAPIInfo(), c)

}

func (api *API) adminGetQueue(c echo.Context) error {

	resp := api.service.GetQueue(&model.Queue{})

	return api.SuccessResponse(resp, c)

}

func (api *API) adminDeleteQueue(c echo.Context) error {

	req := &model.Queue{}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	if err := api.service.DeleteQueue(req); err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	return api.SuccessResponse(req, c)

}

func (api *API) adminGetUsers(c echo.Context) error {

	user := &model.User{}

	resp := api.service.GetUsers(user)

	return api.SuccessResponse(resp, c)

}

func (api *API) adminCreateUser(c echo.Context) error {

	req := &model.User{}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	// generate access token
	req.AccessToken = req.GenerateAccessToken(AccessTokenLength)

	// validate Name, AccessToken
	if err := api.validate.StructExcept(req, "ID"); err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	resp, err := api.service.CreateUser(req)

	if err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	return api.SuccessResponse(resp, c)

}

func (api *API) adminUpdateUser(c echo.Context) error {

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return api.ErrorResponse(errors.New(errors.ValidationError, err), c)
	}

	user := api.service.GetUser(&model.User{ID: userID})

	// bind input data
	if err := c.Bind(user); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	if err := api.service.UpdateUser(user); err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	return api.SuccessResponse(user, c)

}

func (api *API) adminRotateUserToken(c echo.Context) error {

	req := &model.User{}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	req.AccessToken = req.GenerateAccessToken(AccessTokenLength)

	if err := api.service.UpdateUser(req); err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	return api.SuccessResponse(req, c)

}

func (api *API) adminDeleteUser(c echo.Context) error {

	req := &model.User{}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(errors.New(errors.BindDataError, err), c)
	}

	if err := api.service.DeleteUser(req); err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	return api.SuccessResponse(req, c)

}

// getUser godoc
// @Summary User info
// @Description Get API user info
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Success 200 {object} api.SuccessResponse
// @Router /user [get]
// @Security ApiKeyAuth
func (api *API) getUser(c echo.Context) error {
	resp, _ := api.user.FilterStruct([]string{"api"})
	return c.JSON(http.StatusOK, &resp)
}

// index godoc
// @Summary API info
// @Description Get API version
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Success 200 {object} api.SuccessResponse
// @Router / [get]
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

// createChain godoc
// @Summary Create a chain
// @Description Creates chain on the Factom blockchain
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param extIds formData array true "One or many external ids identifying new chain.<br />**Should be provided as array of base64 strings.**"
// @Param content formData string false "The content of the first entry of the chain.<br />**Should be provided as base64 string.**"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /chains [post]
// @Security ApiKeyAuth
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

// getChains godoc
// @Summary Get chains
// @Description Returns all user's chains
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param start query integer false "Select item you would like to start.<br />E.g. if you've already seen 30 items and want to see next 30, then you will provide **start=30**.<br />*Default: 0*"
// @Param limit query integer false "The number of items you would like back in each page.<br />*Default: 30*"
// @Param status query string false "Filter results by chain's status.<br />One of: **queue**, **processing**, **completed**<br />*By default filtering disabled.*"
// @Param sort query string false "Sorting order.<br />One of: **asc** or **desc**<br />*Default: desc*"
// @Success 200 {object} api.SuccessResponsePagination
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /chains [get]
// @Security ApiKeyAuth
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

// searchChains godoc
// @Summary Search chains
// @Description Search user's chains by external id(s)
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param extIds formData array true "One or many external IDs, that used for search.<br />**Should be provided as array of base64 strings.**"
// @Param start query integer false "Select item you would like to start.<br />E.g. if you've already seen 30 items and want to see next 30, then you will provide **start=30**.<br />*Default: 0*"
// @Param limit query integer false "The number of items you would like back in each page.<br />*Default: 30*"
// @Param status query string false "Filter results by chain's status.<br />One of: **queue**, **processing**, **completed**<br />*By default filtering disabled.*"
// @Param sort query string false "Sorting order.<br />One of: **asc** or **desc**<br />*Default: desc*"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /chains/search [post]
// @Security ApiKeyAuth
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

// getChain godoc
// @Summary Get chain
// @Description Returns Factom chain by Chain ID
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID of the Factom chain."
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /chains/{chainId} [get]
// @Security ApiKeyAuth
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

// createEntry godoc
// @Summary Create an entry
// @Description Creates entry on the Factom blockchain
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param chainId formData string true "Chain ID of the Factom chain, where to add new entry."
// @Param extIds formData array false "One or many external ids identifying new chain.<br />**Should be provided as array of base64 strings.**"
// @Param content formData string false "The content of the new entry of the chain.<br />**Should be provided as base64 string.**"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /entries [post]
// @Security ApiKeyAuth
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

// getEntry godoc
// @Summary Get entry
// @Description Returns Factom entry by EntryHash
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param entryHash path string true "EntryHash of the Factom entry."
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /entries/{entryHash} [get]
// @Security ApiKeyAuth
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

// getChainEntries godoc
// @Summary Get chain entries
// @Description Returns entries of Factom chain
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID of the Factom chain."
// @Param start query integer false "Select item you would like to start.<br />E.g. if you've already seen 30 items and want to see next 30, then you will provide **start=30**.<br />*Default: 0*"
// @Param limit query integer false "The number of items you would like back in each page.<br />*Default: 30*"
// @Param status query string false "Filter results by chain's status.<br />One of: **queue**, **processing**, **completed**<br />*By default filtering disabled.*"
// @Param sort query string false "Sorting order.<br />One of: **asc** or **desc**<br />*Default: desc*"
// @Success 200 {object} api.SuccessResponsePagination
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /chains/{chainId}/entries [get]
// @Security ApiKeyAuth
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

// searchChainEntries godoc
// @Summary Search entries of chain
// @Description Search entries into Factom chain by external id(s)
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID of the Factom chain."
// @Param extIds formData array true "One or many external IDs, that used for search.<br />**Should be provided as array of base64 strings.**"
// @Param start query integer false "Select item you would like to start.<br />E.g. if you've already seen 30 items and want to see next 30, then you will provide **start=30**.<br />*Default: 0*"
// @Param limit query integer false "The number of items you would like back in each page.<br />*Default: 30*"
// @Param status query string false "Filter results by chain's status.<br />One of: **queue**, **processing**, **completed**<br />*By default filtering disabled.*"
// @Param sort query string false "Sorting order.<br />One of: **asc** or **desc**<br />*Default: desc*"
// @Success 200 {object} api.SuccessResponsePagination
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /chains/{chainId}/entries/search [post]
// @Security ApiKeyAuth
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

// getChainFirstEntry godoc
// @Summary Get first entry of the chain
// @Description Returns first entry of Factom chain
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID of the Factom chain."
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /chains/{chainId}/entries/first [get]
// @Security ApiKeyAuth
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

// getChainLastEntry godoc
// @Summary Get last entry of the chain
// @Description Returns last entry of Factom chain
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param chainId path string true "Chain ID of the Factom chain."
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /chains/{chainId}/entries/last [get]
// @Security ApiKeyAuth
func _() {}

// factomd godoc
// @Summary Generic factomd
// @Description Sends direct request to factomd API
// @Accept x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param method path string true "factomd API method"
// @Param params formData string false "factomd request's params.<br />**Should be provided as JSON string,** e.g. *{'chainid':'XXXX'}*"
// @Router /factomd/{method} [post]
// @Security ApiKeyAuth
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

	request := factom.NewJSON2Request(c.Param("method"), 0, params)

	resp, err := factom.SendFactomdRequest(request)

	if err != nil {
		return api.ErrorResponse(errors.New(errors.ServiceError, err), c)
	}

	if resp.Error != nil {
		return api.ErrorResponse(errors.New(resp.Error.Code, resp.Error), c)
	}

	return api.SuccessResponse(resp.Result, c)

}

// helpers

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

func generateJWTSecret() []byte {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, 128)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return []byte(string((b)))
}

func writeCookie(c echo.Context, token string) error {
	cookie := new(http.Cookie)
	cookie.Path = "/"
	cookie.Name = "token"
	cookie.Value = token
	cookie.Expires = time.Now().Add(time.Hour)
	c.SetCookie(cookie)
	return nil
}

func deleteCookie(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Path = "/"
	cookie.Name = "token"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	c.SetCookie(cookie)
	return nil
}
