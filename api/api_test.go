package api

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/user"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/service"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	"github.com/DeFacto-Team/Factom-Open-API/wallet"
	"github.com/FactomProject/factom"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func NewTestAPI() *API {

	usr, _ := user.Current()

	configFile := usr.HomeDir + "/.foa/config_test.yaml"
	conf, _ := config.NewConfig(configFile)

	store, _ := store.NewStore(conf, false)
	wallet, _ := wallet.NewWallet(conf)

	s := service.NewService(store, wallet)

	// Create factom
	if conf.Factom.URL != "" {
		factom.SetFactomdServer(conf.Factom.URL)
	}
	if conf.Factom.User != "" && conf.Factom.Password != "" {
		factom.SetFactomdRpcConfig(conf.Factom.User, conf.Factom.Password)
	}

	return NewAPI(conf, s, configFile)

}

func TestLogin(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Setup echo context
	f := make(url.Values)
	f.Set("user", testAPI.conf.Admin.User)
	f.Set("password", testAPI.conf.Admin.Password)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.login(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestAdminIndex(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.adminIndex(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestAdminGetQueue(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create test user and queue item
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	tc := &model.Chain{}
	tc.ExtIDs = []string{strconv.FormatInt(time.Now().UnixNano(), 10)}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}
	tq := testAPI.service.GetQueue(&model.Queue{})
	var tqid int
	for _, tqi := range tq {
		tqid = tqi.ID
		break
	}

	// Assertions
	if assert.NoError(t, testAPI.adminGetQueue(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test queue item and user
	testAPI.service.DeleteQueue(&model.Queue{ID: tqid})
	testAPI.service.DeleteUser(tu)

}

func TestAdminDeleteQueue(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and queue item
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	tc := &model.Chain{}
	tc.ExtIDs = []string{strconv.FormatInt(time.Now().UnixNano(), 10)}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}
	tq := testAPI.service.GetQueue(&model.Queue{})
	var tqid int
	for _, tqi := range tq {
		tqid = tqi.ID
		break
	}

	// Setup echo context
	f := make(url.Values)
	f.Set("id", strconv.Itoa(tqid))
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.adminDeleteQueue(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestAdminGetUsers(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create test user
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}

	// Assertions
	if assert.NoError(t, testAPI.adminGetUsers(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestAdminCreateUser(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Generate test user
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)

	// Setup echo context
	f := make(url.Values)
	f.Set("name", tu.Name)
	f.Set("accessToken", tu.AccessToken)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.adminCreateUser(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestAdminDeleteUser(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	f := make(url.Values)
	f.Set("id", strconv.Itoa(tu.ID))
	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.adminDeleteUser(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestAdminGetUser(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create test user
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(tu.ID))

	// Assertions
	if assert.NoError(t, testAPI.adminGetUser(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestAdminUpdateUser(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	f := make(url.Values)
	f.Set("usage", "1000")
	f.Set("status", "0")
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(tu.ID))

	// Assertions
	if assert.NoError(t, testAPI.adminUpdateUser(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestAdminRotateUserToken(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create test user
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(tu.ID))

	// Assertions
	if assert.NoError(t, testAPI.adminRotateUserToken(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestAdminLogout(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.adminLogout(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestAdminGetSettings(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.adminGetSettings(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestAdminUpdateSettings(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Setup echo context
	f := make(url.Values)
	f.Set("adminUser", "test")
	f.Set("adminPassword", "test")
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.adminUpdateSettings(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestAdminRestartAPI(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.adminRestartAPI(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestAdminRandomEC(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.adminRandomEC(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestAdminGetEC(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create random EC wallet
	ecAddress := model.GenerateEC()

	// Setup echo context
	c.SetParamNames("esaddress")
	c.SetParamValues(ecAddress.EsAddress)

	// Assertions
	if assert.NoError(t, testAPI.adminGetEC(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestIndex(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.index(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}

func TestCreateChain(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	// Setup echo context
	f := make(url.Values)
	extId := strconv.FormatInt(time.Now().UnixNano(), 10)
	extId = base64.StdEncoding.EncodeToString([]byte(extId))
	f.Set("extIds", extId)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.createChain(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestGetChains(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and chain
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	tc := &model.Chain{}
	tc.ExtIDs = []string{strconv.FormatInt(time.Now().UnixNano(), 10)}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.getChains(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestGetChain(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and chain
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	tc := &model.Chain{}
	tc.ExtIDs = []string{strconv.FormatInt(time.Now().UnixNano(), 10)}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames("chainid")
	c.SetParamValues(tc.ChainID)

	// Assertions
	if assert.NoError(t, testAPI.getChain(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestSearchChains(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and chain
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	tc := &model.Chain{}
	extId := strconv.FormatInt(time.Now().UnixNano(), 10)
	tc.ExtIDs = []string{extId}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	f := make(url.Values)
	extId = base64.StdEncoding.EncodeToString([]byte(extId))
	f.Set("extIds", extId)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.searchChains(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestGetChainEntries(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and chain
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	tc := &model.Chain{}
	tc.ExtIDs = []string{strconv.FormatInt(time.Now().UnixNano(), 10)}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames("chainid")
	c.SetParamValues(tc.ChainID)

	// Assertions
	if assert.NoError(t, testAPI.getChainEntries(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestSearchChainEntries(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and chain
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	tc := &model.Chain{}
	extId := strconv.FormatInt(time.Now().UnixNano(), 10)
	tc.ExtIDs = []string{extId}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	f := make(url.Values)
	extId = base64.StdEncoding.EncodeToString([]byte(extId))
	f.Set("extIds", extId)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames("chainid")
	c.SetParamValues(tc.ChainID)

	// Assertions
	if assert.NoError(t, testAPI.searchChainEntries(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestGetChainFirstOrLastEntry(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and chain
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	tc := &model.Chain{}
	tc.ExtIDs = []string{strconv.FormatInt(time.Now().UnixNano(), 10)}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames("chainid", "item")
	c.SetParamValues(tc.ChainID, "first")

	// Assertions
	if assert.NoError(t, testAPI.getChainFirstOrLastEntry(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestCreateEntry(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and chain
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	tc := &model.Chain{}
	extId := strconv.FormatInt(time.Now().UnixNano(), 10)
	tc.ExtIDs = []string{extId}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	f := make(url.Values)
	extId = strconv.FormatInt(time.Now().UnixNano(), 10)
	extId = base64.StdEncoding.EncodeToString([]byte(extId))
	f.Set("extIds", extId)
	f.Set("chainId", tc.ChainID)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.createEntry(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestGetEntry(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and chain
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	tc := &model.Chain{}
	tc.ExtIDs = []string{strconv.FormatInt(time.Now().UnixNano(), 10)}
	tc, err = testAPI.service.CreateChain(tc.Base64Encode(), tu)
	if err != nil {
		t.Error(err)
	}

	// Setup echo context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames("entryhash")
	c.SetParamValues(tc.Base64Decode().FirstEntryHash())

	// Assertions
	if assert.NoError(t, testAPI.getEntry(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}

func TestGetUser(t *testing.T) {

	// Setup
	testAPI := NewTestAPI()
	e := echo.New()

	// Create test user and chain
	tu := &model.User{}
	tu.Name = "Test"
	tu.AccessToken = tu.GenerateAccessToken(32)
	tu, err := testAPI.service.CreateUser(tu)
	if err != nil {
		t.Error(err)
	}
	testAPI.user = tu

	// Setup echo context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, testAPI.getUser(c)) {
		t.Logf(rec.Body.String())
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Delete test user
	testAPI.service.DeleteUser(tu)

}
