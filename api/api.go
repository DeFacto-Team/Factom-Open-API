// Factom-Open-API
// Version: 1.0
// Schemes: http
// Host: localhost
// BasePath: /v1
// Consumes:
// - application/json
// - application/x-www-form-urlencoded
// - multipart/form-data
// Produces:
// - application/json
// Contact: team@de-facto.pro
// swagger:meta
package api

import (
	"net/http"
	"strconv"

	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/service"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"gopkg.in/go-playground/validator.v9"
)

type Api struct {
	Http     *echo.Echo
	conf     *config.Config
	es       service.EntryService
	cs       service.ChainService
	us       service.UserService
	apiInfo  ApiInfo
	validate *validator.Validate
	user     *model.User
}

type ApiInfo struct {
	Address string
	MW      []string
}

type ErrorResponse struct {
	Result bool   `json:"result"`
	Code   int    `json:"code"`
	Error  string `json:"error"`
}

type SuccessResponse struct {
	Result interface{} `json:"result"`
}

type EntryResponse struct {
	*model.Entry
	Status string   `json:"status" form:"status" query:"status" validate:"oneof=queue processing completed"`
	Links  []string `json:"links" form:"links" query:"links" validate:""`
}

type ChainResponse struct {
	*model.Chain
	Status string   `json:"status" form:"status" query:"status" validate:"oneof=queue processing completed"`
	Links  []string `json:"links" form:"links" query:"links" validate:""`
}

func NewApi(conf *config.Config, es service.EntryService, cs service.ChainService, us service.UserService) *Api {

	api := &Api{}

	api.validate = validator.New()

	api.conf = conf
	api.es = es
	api.cs = cs
	api.us = us

	api.Http = echo.New()
	api.Http.Logger.SetLevel(log.Lvl(conf.LogLevel))
	api.apiInfo.Address = ":" + strconv.Itoa(api.conf.Api.HttpPort)
	api.Http.HideBanner = true
	api.Http.Pre(middleware.RemoveTrailingSlash())

	if conf.Api.Logging {
		api.Http.Use(middleware.Logger())
		api.apiInfo.MW = append(api.apiInfo.MW, "Logger")
	}

	api.Http.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		user, err := api.us.GetUserByAccessToken(key)
		if user != nil && err == nil {
			api.user = user
			return true, nil
		}
		return false, err
	}))

	api.apiInfo.MW = append(api.apiInfo.MW, "KeyAuth")

	api.Http.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: conf.GzipLevel,
	}))

	// Status
	api.Http.GET("/v1", api.index)

	// API specification
	api.Http.Static("/v1/spec", "spec")

	// Chains
	api.Http.POST("/v1/chains", api.createChain)

	// Entries
	api.Http.POST("/v1/entries", api.createEntry)
	//api.Http.GET("/v1/entries/:entryhash", api.getEntry)

	// User

	return api
}

// Start API server
func (api *Api) Start() error {
	return api.Http.Start(":" + strconv.Itoa(api.conf.Api.HttpPort))
}

// Returns API information
func (api *Api) GetApiInfo() ApiInfo {
	return api.apiInfo
}

// Returns API specification (Swagger)
func (api *Api) index(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "/spec/api.json")
}

func (api *Api) spec(c echo.Context) error {
	return c.Inline("spec/api.json", "api.json")
}

// Success API response
func (api *Api) SuccessResponse(res interface{}, c echo.Context) error {
	return c.JSON(http.StatusOK, &SuccessResponse{Result: res})
}

// Custom API response in case of error
func (api *Api) ErrorResponse(err error, c echo.Context) error {
	resp := &ErrorResponse{
		Result: false,
		Code:   http.StatusBadRequest,
		Error:  err.Error(),
	}
	api.Http.Logger.Error(resp.Error)
	return c.JSON(resp.Code, resp)
}

// API functions

func (api *Api) createChain(c echo.Context) error {

	// Open API Chain struct
	req := &model.Chain{}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(err, c)
	}

	// validate ExtIDs, Content
	if err := api.validate.StructExcept(req, "ChainID"); err != nil {
		return api.ErrorResponse(err, c)
	}

	// check if first entry of chain fits into 10KB
	_, err := req.ConvertToEntryModel().Fit10KB()
	if err != nil {
		return api.ErrorResponse(err, c)
	}

	// calculate chainID
	req.ChainID = req.ID()

	// extend Chain
	resp := &ChainResponse{Chain: req}

	// calculate entryhash of the first entry
	resp.Links = append(resp.Links, "/chains/"+req.ChainID+"/entries/"+req.FirstEntryHash())

	// check for chain into local DB
	localChain, err := api.cs.GetChain(req)
	if err != nil {
		return err
	}

	// flag
	checkChainExistence := false

	// if chain doesn't exists in local DB
	if localChain == nil {
		// Create chain
		_, err = api.cs.CreateChain(req)
		if err != nil {
			return err
		}
		// need to check chain existence — set flag to true
		checkChainExistence = true
		// Just created chain has status "queue"
		resp.Status = model.ChainQueue
	} else {
		// Chain status is not completed
		if localChain.Status != model.ChainCompleted {
			checkChainExistence = true
		}
		resp.Status = localChain.Status
	}

	// if need to check chain existence
	if checkChainExistence == true {
		if req.Exists() == false {
			// if chain doesn't exist on the blockchain — add it to queue
			log.Info("ADD TO QUEUE")
		} else {
			// if chain exists on the blockchain - get it's status
			resp.Status = req.GetStatusFromFactom()
		}
	}

	// If we are here, so no errors occured and we force bind chain to API user
	err = api.cs.BindChainToUser(req, api.user)
	if err != nil {
		return err
	}

	return api.SuccessResponse(resp, c)
}

// swagger:operation POST /entries createEntry
// ---
// description: Create entry in chain
// parameters:
// - name: chainid
//   in: body
//   description: Chain ID of the Factom chain where to add new entry.
//   required: true
//   type: string
// - name: content
//   in: body
//   description: The content of new entry.
//   required: true
//   type: string
// - name: extids
//   in: body
//   description: One or many external ids identifying new entry. Should be sent as array of strings.
//   required: false
//   type: array
func (api *Api) createEntry(c echo.Context) error {

	// Open API Entry struct
	req := &model.Entry{}

	// bind input data
	if err := c.Bind(req); err != nil {
		return api.ErrorResponse(err, c)
	}

	// validate ChainID, ExtID (if exists), Content
	if err := api.validate.StructExcept(req, "EntryHash"); err != nil {
		return api.ErrorResponse(err, c)
	}

	// check if entry fits into 10KB
	_, err := req.Fit10KB()
	if err != nil {
		return api.ErrorResponse(err, c)
	}

	// extend Entry
	resp := &EntryResponse{Entry: req}

	// calculate entryhash
	resp.EntryHash = req.Hash()
	resp.Status = "completed"

	// Create entry
	_, err = api.es.CreateEntry(req)
	if err != nil {
		return err
	}

	// send to factomd
	//	factom.ComposeTransaction()

	// increase user's usage
	api.user.Usage += 1
	api.us.UpdateUser(api.user)
	//	log.Info(api.user)

	return api.SuccessResponse(resp, c)
}

/*
func (api *Api) getEntry(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param `id`")
	}
	cat, err := api.es.GetEntry(id)
	if err != nil {
		return err
	}
	if cat == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Category `id` = ", id, " not found")
	}
	return c.JSON(http.StatusOK, cat)
}
*/
