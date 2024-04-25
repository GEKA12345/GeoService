package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"test/proxy/internal/responder"
	"test/proxy/internal/service"
	"testing"

	"github.com/go-chi/jwtauth/v5"
	"github.com/ptflp/godecoder"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestController_Authenticator(t *testing.T) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := responder.NewResponder(decoder, logger)
	serv := service.NewGeoService(decoder)
	contrl := NewController(respond, decoder, serv)

	mockJWTAuth := jwtauth.New("HS256", []byte("salt_01"), nil)

	// Создаем фэйковый http.Handler
	fakeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Тест 1: проверяем, что пользователь с правильным токеном проходит успешно
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6IlVzZXIxIiwicGFzc3dvcmQiOiIkMmEkMTAkaVUyaTQuTGdhdnEuZE9rdmtoZDZyZVY4QUVNbm1CLy5KWmJCOVpMZ2ZSYXVHY2ZnOHZWWmUifQ.h3iq4QISSdE1x4m7vVv9_U9fZukQagVlxvodMuwXaro"
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	handler := jwtauth.Verifier(mockJWTAuth)(contrl.Authenticator(mockJWTAuth)(fakeHandler))
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)

	// Тест 2: проверяем, что пользователь с неправильным токеном блокируется
	req, err = http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+"invalid_token_string")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusForbidden)

	// Тест 3: проверяем, что пользователь без токена блокируется
	req, err = http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusForbidden)
}

func TestController_Register(t *testing.T) {
	service.TokenAuth = jwtauth.New("HS256", []byte("salt_01"), nil)

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := responder.NewResponder(decoder, logger)
	serv := service.NewGeoService(decoder)
	contrl := NewController(respond, decoder, serv)

	reqGet := httptest.NewRequest(http.MethodGet, "/", nil)
	dataBad := strings.NewReader(`d`)
	reqBad := httptest.NewRequest(http.MethodPost, "/", dataBad)
	dataUser1 := strings.NewReader(`{"login":"User1", "password": "qwerty"}`)
	reqUser1 := httptest.NewRequest(http.MethodPost, "/", dataUser1)
	dataUserConflict := strings.NewReader(`{"login":"User1", "password": "qwerty"}`)
	reqUserConflict := httptest.NewRequest(http.MethodPost, "/", dataUserConflict)

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name  string
		cntrl Controllerer
		args  args
		want  int
	}{
		{"1", contrl, args{w: httptest.NewRecorder(), r: reqGet}, http.StatusMethodNotAllowed},
		{"2", contrl, args{w: httptest.NewRecorder(), r: reqBad}, http.StatusBadRequest},
		{"3", contrl, args{w: httptest.NewRecorder(), r: reqUser1}, http.StatusOK},
		{"4", contrl, args{w: httptest.NewRecorder(), r: reqUserConflict}, http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(tt.cntrl.Register)
			handler.ServeHTTP(tt.args.w, tt.args.r)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

func TestController_Login(t *testing.T) {
	service.TokenAuth = jwtauth.New("HS256", []byte("salt_01"), nil)

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := responder.NewResponder(decoder, logger)
	serv := service.NewGeoService(decoder)
	contrl := NewController(respond, decoder, serv)

	reqGet := httptest.NewRequest(http.MethodGet, "/", nil)
	dataBad := strings.NewReader(`d`)
	reqBad := httptest.NewRequest(http.MethodPost, "/", dataBad)
	dataUser1Register := strings.NewReader(`{"login":"User1", "password": "qwerty"}`)
	reqUser1Register := httptest.NewRequest(http.MethodPost, "/", dataUser1Register)
	handlerRegister := http.HandlerFunc(contrl.Register)
	handlerRegister.ServeHTTP(httptest.NewRecorder(), reqUser1Register)
	dataUser1 := strings.NewReader(`{"login":"User1", "password": "qwerty"}`)
	reqUser1 := httptest.NewRequest(http.MethodPost, "/", dataUser1)
	dataUser2 := strings.NewReader(`{"login":"User2", "password": "qwerty"}`)
	reqUser2 := httptest.NewRequest(http.MethodPost, "/", dataUser2)

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name  string
		cntrl Controllerer
		args  args
		want  int
	}{
		{"1", contrl, args{w: httptest.NewRecorder(), r: reqGet}, http.StatusMethodNotAllowed},
		{"2", contrl, args{w: httptest.NewRecorder(), r: reqBad}, http.StatusBadRequest},
		{"3", contrl, args{w: httptest.NewRecorder(), r: reqUser1}, http.StatusOK},
		{"4", contrl, args{w: httptest.NewRecorder(), r: reqUser2}, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(tt.cntrl.Login)
			handler.ServeHTTP(tt.args.w, tt.args.r)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

func TestController_GeoSearch(t *testing.T) {
	service.TestEnabled = true

	handlerGeo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, mockResSearch)
	})

	handler500 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	server500 := httptest.NewServer(handler500)
	defer server500.Close()

	serverGeo := httptest.NewServer(handlerGeo)
	defer serverGeo.Close()

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := responder.NewResponder(decoder, logger)
	serv := service.NewGeoService(decoder)
	contrl := NewController(respond, decoder, serv)

	reqGet := httptest.NewRequest(http.MethodGet, "/", nil)
	dataBad := strings.NewReader(`d`)
	reqBad := httptest.NewRequest(http.MethodPost, "/", dataBad)
	dataOK := strings.NewReader(`{"query":"Ленинский проспект, 118к1, Санкт-Петербург"}`)
	reqOK := httptest.NewRequest(http.MethodPost, "/", dataOK)
	data500 := strings.NewReader(`{"query":"Ленинский проспект, 118к1, Санкт-Петербург"}`)
	req500 := httptest.NewRequest(http.MethodPost, "/", data500)

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name      string
		cntrl     Controllerer
		serverAPI *httptest.Server
		args      args
		want      int
	}{
		{"1", contrl, serverGeo, args{w: httptest.NewRecorder(), r: reqGet}, http.StatusMethodNotAllowed},
		{"2", contrl, serverGeo, args{w: httptest.NewRecorder(), r: reqBad}, http.StatusBadRequest},
		{"3", contrl, serverGeo, args{w: httptest.NewRecorder(), r: reqOK}, http.StatusOK},
		{"4", contrl, server500, args{w: httptest.NewRecorder(), r: req500}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service.TestSearchHost = tt.serverAPI.URL
			handler := http.HandlerFunc(tt.cntrl.GeoSearch)
			handler.ServeHTTP(tt.args.w, tt.args.r)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

func TestController_GeoCode(t *testing.T) {
	service.TestEnabled = true

	handlerGeo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, mockResGeo)
	})

	handler500 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	server500 := httptest.NewServer(handler500)
	defer server500.Close()

	serverGeo := httptest.NewServer(handlerGeo)
	defer serverGeo.Close()

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := responder.NewResponder(decoder, logger)
	serv := service.NewGeoService(decoder)
	contrl := NewController(respond, decoder, serv)

	reqGet := httptest.NewRequest(http.MethodGet, "/", nil)
	dataBad := strings.NewReader(`d`)
	reqBad := httptest.NewRequest(http.MethodPost, "/", dataBad)
	dataOK := strings.NewReader(`{"lat":"59.93986890851519","lng":"30.26046752929688"}`)
	reqOK := httptest.NewRequest(http.MethodPost, "/", dataOK)
	data500 := strings.NewReader(`{"lat":"59.93986890851519","lng":"30.26046752929688"}`)
	req500 := httptest.NewRequest(http.MethodPost, "/", data500)

	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name      string
		cntrl     Controllerer
		serverAPI *httptest.Server
		args      args
		want      int
	}{
		{"1", contrl, serverGeo, args{w: httptest.NewRecorder(), r: reqGet}, http.StatusMethodNotAllowed},
		{"2", contrl, serverGeo, args{w: httptest.NewRecorder(), r: reqBad}, http.StatusBadRequest},
		{"3", contrl, serverGeo, args{w: httptest.NewRecorder(), r: reqOK}, http.StatusOK},
		{"4", contrl, server500, args{w: httptest.NewRecorder(), r: req500}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service.TestGeoHost = tt.serverAPI.URL
			handler := http.HandlerFunc(tt.cntrl.GeoCode)
			handler.ServeHTTP(tt.args.w, tt.args.r)
			assert.Equal(t, tt.args.w.Code, tt.want)
		})
	}
}

var mockResSearch = `[
	{
		"source": "москва сухонская 11",
		"result": "г Москва, ул Сухонская, д 11",
		"postal_code": "127642",
		"country": "Россия",
		"region": "Москва",
		"city_area": "Северо-восточный",
		"city_district": "Северное Медведково",
		"street": "Сухонская",
		"house": "11",
		"geo_lat": "55.8782557",
		"geo_lon": "37.65372",
		"qc_geo": 0
	}
	]`

var mockResGeo = `{"suggestions":[{"value":"г Москва, ул Сухонская, д 11","unrestricted_value":"127642, г Москва, р-н Северное Медведково, ул Сухонская, д 11","data":{"postal_code":"127642","country":"Россия","country_iso_code":"RU","federal_district":"Центральный","region_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","region_kladr_id":"7700000000000","region_iso_code":"RU-MOW","region_with_type":"г Москва","region_type":"г","region_type_full":"город","region":"Москва","area_fias_id":null,"area_kladr_id":null,"area_with_type":null,"area_type":null,"area_type_full":null,"area":null,"city_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","city_kladr_id":"7700000000000","city_with_type":"г Москва","city_type":"г","city_type_full":"город","city":"Москва","city_area":"Северо-восточный","city_district_fias_id":null,"city_district_kladr_id":null,"city_district_with_type":null,"city_district_type":null,"city_district_type_full":null,"city_district":null,"settlement_fias_id":null,"settlement_kladr_id":null,"settlement_with_type":null,"settlement_type":null,"settlement_type_full":null,"settlement":null,"street_fias_id":"95dbf7fb-0dd4-4a04-8100-4f6c847564b5","street_kladr_id":"77000000000283600","street_with_type":"ул Сухонская","street_type":"ул","street_type_full":"улица","street":"Сухонская","stead_fias_id":null,"stead_cadnum":null,"stead_type":null,"stead_type_full":null,"stead":null,"house_fias_id":"5ee84ac0-eb9a-4b42-b814-2f5f7c27c255","house_kladr_id":"7700000000028360004","house_cadnum":null,"house_type":"д","house_type_full":"дом","house":"11","block_type":null,"block_type_full":null,"block":null,"entrance":null,"floor":null,"flat_fias_id":null,"flat_cadnum":null,"flat_type":null,"flat_type_full":null,"flat":null,"flat_area":null,"square_meter_price":null,"flat_price":null,"room_fias_id":null,"room_cadnum":null,"room_type":null,"room_type_full":null,"room":null,"postal_box":null,"fias_id":"5ee84ac0-eb9a-4b42-b814-2f5f7c27c255","fias_code":null,"fias_level":"8","fias_actuality_state":"0","kladr_id":"7700000000028360004","geoname_id":"524901","capital_marker":"0","okato":"45280583000","oktmo":"45362000","tax_office":"7715","tax_office_legal":"7715","timezone":null,"geo_lat":"55.878315","geo_lon":"37.65372","beltway_hit":null,"beltway_distance":null,"metro":null,"divisions":null,"qc_geo":"0","qc_complete":null,"qc_house":null,"history_values":null,"unparsed_parts":null,"source":null,"qc":null}},{"value":"г Москва, ул Сухонская, д 11А","unrestricted_value":"127642, г Москва, р-н Северное Медведково, ул Сухонская, д 11А","data":{"postal_code":"127642","country":"Россия","country_iso_code":"RU","federal_district":"Центральный","region_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","region_kladr_id":"7700000000000","region_iso_code":"RU-MOW","region_with_type":"г Москва","region_type":"г","region_type_full":"город","region":"Москва","area_fias_id":null,"area_kladr_id":null,"area_with_type":null,"area_type":null,"area_type_full":null,"area":null,"city_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","city_kladr_id":"7700000000000","city_with_type":"г Москва","city_type":"г","city_type_full":"город","city":"Москва","city_area":"Северо-восточный","city_district_fias_id":null,"city_district_kladr_id":null,"city_district_with_type":null,"city_district_type":null,"city_district_type_full":null,"city_district":null,"settlement_fias_id":null,"settlement_kladr_id":null,"settlement_with_type":null,"settlement_type":null,"settlement_type_full":null,"settlement":null,"street_fias_id":"95dbf7fb-0dd4-4a04-8100-4f6c847564b5","street_kladr_id":"77000000000283600","street_with_type":"ул Сухонская","street_type":"ул","street_type_full":"улица","street":"Сухонская","stead_fias_id":null,"stead_cadnum":null,"stead_type":null,"stead_type_full":null,"stead":null,"house_fias_id":"abc31736-35c1-4443-a061-b67c183b590a","house_kladr_id":"7700000000028360005","house_cadnum":null,"house_type":"д","house_type_full":"дом","house":"11А","block_type":null,"block_type_full":null,"block":null,"entrance":null,"floor":null,"flat_fias_id":null,"flat_cadnum":null,"flat_type":null,"flat_type_full":null,"flat":null,"flat_area":null,"square_meter_price":null,"flat_price":null,"room_fias_id":null,"room_cadnum":null,"room_type":null,"room_type_full":null,"room":null,"postal_box":null,"fias_id":"abc31736-35c1-4443-a061-b67c183b590a","fias_code":null,"fias_level":"8","fias_actuality_state":"0","kladr_id":"7700000000028360005","geoname_id":"524901","capital_marker":"0","okato":"45280583000","oktmo":"45362000","tax_office":"7715","tax_office_legal":"7715","timezone":null,"geo_lat":"55.878212","geo_lon":"37.652016","beltway_hit":null,"beltway_distance":null,"metro":null,"divisions":null,"qc_geo":"0","qc_complete":null,"qc_house":null,"history_values":null,"unparsed_parts":null,"source":null,"qc":null}},{"value":"г Москва, ул Сухонская, д 13","unrestricted_value":"127642, г Москва, р-н Северное Медведково, ул Сухонская, д 13","data":{"postal_code":"127642","country":"Россия","country_iso_code":"RU","federal_district":"Центральный","region_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","region_kladr_id":"7700000000000","region_iso_code":"RU-MOW","region_with_type":"г Москва","region_type":"г","region_type_full":"город","region":"Москва","area_fias_id":null,"area_kladr_id":null,"area_with_type":null,"area_type":null,"area_type_full":null,"area":null,"city_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","city_kladr_id":"7700000000000","city_with_type":"г Москва","city_type":"г","city_type_full":"город","city":"Москва","city_area":"Северо-восточный","city_district_fias_id":null,"city_district_kladr_id":null,"city_district_with_type":null,"city_district_type":null,"city_district_type_full":null,"city_district":null,"settlement_fias_id":null,"settlement_kladr_id":null,"settlement_with_type":null,"settlement_type":null,"settlement_type_full":null,"settlement":null,"street_fias_id":"95dbf7fb-0dd4-4a04-8100-4f6c847564b5","street_kladr_id":"77000000000283600","street_with_type":"ул Сухонская","street_type":"ул","street_type_full":"улица","street":"Сухонская","stead_fias_id":null,"stead_cadnum":null,"stead_type":null,"stead_type_full":null,"stead":null,"house_fias_id":"301be60e-97c6-4ac4-a45c-11efee1c200a","house_kladr_id":"7700000000028360006","house_cadnum":null,"house_type":"д","house_type_full":"дом","house":"13","block_type":null,"block_type_full":null,"block":null,"entrance":null,"floor":null,"flat_fias_id":null,"flat_cadnum":null,"flat_type":null,"flat_type_full":null,"flat":null,"flat_area":null,"square_meter_price":null,"flat_price":null,"room_fias_id":null,"room_cadnum":null,"room_type":null,"room_type_full":null,"room":null,"postal_box":null,"fias_id":"301be60e-97c6-4ac4-a45c-11efee1c200a","fias_code":null,"fias_level":"8","fias_actuality_state":"0","kladr_id":"7700000000028360006","geoname_id":"524901","capital_marker":"0","okato":"45280583000","oktmo":"45362000","tax_office":"7715","tax_office_legal":"7715","timezone":null,"geo_lat":"55.878666","geo_lon":"37.6524","beltway_hit":null,"beltway_distance":null,"metro":null,"divisions":null,"qc_geo":"0","qc_complete":null,"qc_house":null,"history_values":null,"unparsed_parts":null,"source":null,"qc":null}},{"value":"г Москва, ул Сухонская, д 9","unrestricted_value":"127642, г Москва, р-н Северное Медведково, ул Сухонская, д 9","data":{"postal_code":"127642","country":"Россия","country_iso_code":"RU","federal_district":"Центральный","region_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","region_kladr_id":"7700000000000","region_iso_code":"RU-MOW","region_with_type":"г Москва","region_type":"г","region_type_full":"город","region":"Москва","area_fias_id":null,"area_kladr_id":null,"area_with_type":null,"area_type":null,"area_type_full":null,"area":null,"city_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","city_kladr_id":"7700000000000","city_with_type":"г Москва","city_type":"г","city_type_full":"город","city":"Москва","city_area":"Северо-восточный","city_district_fias_id":null,"city_district_kladr_id":null,"city_district_with_type":null,"city_district_type":null,"city_district_type_full":null,"city_district":null,"settlement_fias_id":null,"settlement_kladr_id":null,"settlement_with_type":null,"settlement_type":null,"settlement_type_full":null,"settlement":null,"street_fias_id":"95dbf7fb-0dd4-4a04-8100-4f6c847564b5","street_kladr_id":"77000000000283600","street_with_type":"ул Сухонская","street_type":"ул","street_type_full":"улица","street":"Сухонская","stead_fias_id":null,"stead_cadnum":null,"stead_type":null,"stead_type_full":null,"stead":null,"house_fias_id":"c68ee16b-e36a-427f-a8b7-5762d3562cf8","house_kladr_id":"7700000000028360002","house_cadnum":null,"house_type":"д","house_type_full":"дом","house":"9","block_type":null,"block_type_full":null,"block":null,"entrance":null,"floor":null,"flat_fias_id":null,"flat_cadnum":null,"flat_type":null,"flat_type_full":null,"flat":null,"flat_area":null,"square_meter_price":null,"flat_price":null,"room_fias_id":null,"room_cadnum":null,"room_type":null,"room_type_full":null,"room":null,"postal_box":null,"fias_id":"c68ee16b-e36a-427f-a8b7-5762d3562cf8","fias_code":null,"fias_level":"8","fias_actuality_state":"0","kladr_id":"7700000000028360002","geoname_id":"524901","capital_marker":"0","okato":"45280583000","oktmo":"45362000","tax_office":"7715","tax_office_legal":"7715","timezone":null,"geo_lat":"55.877167","geo_lon":"37.652481","beltway_hit":null,"beltway_distance":null,"metro":null,"divisions":null,"qc_geo":"0","qc_complete":null,"qc_house":null,"history_values":null,"unparsed_parts":null,"source":null,"qc":null}},{"value":"г Москва","unrestricted_value":"101000, г Москва","data":{"postal_code":"101000","country":"Россия","country_iso_code":"RU","federal_district":"Центральный","region_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","region_kladr_id":"7700000000000","region_iso_code":"RU-MOW","region_with_type":"г Москва","region_type":"г","region_type_full":"город","region":"Москва","area_fias_id":null,"area_kladr_id":null,"area_with_type":null,"area_type":null,"area_type_full":null,"area":null,"city_fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","city_kladr_id":"7700000000000","city_with_type":"г Москва","city_type":"г","city_type_full":"город","city":"Москва","city_area":null,"city_district_fias_id":null,"city_district_kladr_id":null,"city_district_with_type":null,"city_district_type":null,"city_district_type_full":null,"city_district":null,"settlement_fias_id":null,"settlement_kladr_id":null,"settlement_with_type":null,"settlement_type":null,"settlement_type_full":null,"settlement":null,"street_fias_id":null,"street_kladr_id":null,"street_with_type":null,"street_type":null,"street_type_full":null,"street":null,"stead_fias_id":null,"stead_cadnum":null,"stead_type":null,"stead_type_full":null,"stead":null,"house_fias_id":null,"house_kladr_id":null,"house_cadnum":null,"house_type":null,"house_type_full":null,"house":null,"block_type":null,"block_type_full":null,"block":null,"entrance":null,"floor":null,"flat_fias_id":null,"flat_cadnum":null,"flat_type":null,"flat_type_full":null,"flat":null,"flat_area":null,"square_meter_price":null,"flat_price":null,"room_fias_id":null,"room_cadnum":null,"room_type":null,"room_type_full":null,"room":null,"postal_box":null,"fias_id":"0c5b2444-70a0-4932-980c-b4dc0d3f02b5","fias_code":null,"fias_level":"1","fias_actuality_state":"0","kladr_id":"7700000000000","geoname_id":"524901","capital_marker":"0","okato":"45000000000","oktmo":"45000000","tax_office":"7700","tax_office_legal":"7700","timezone":null,"geo_lat":"55.75396","geo_lon":"37.620393","beltway_hit":null,"beltway_distance":null,"metro":null,"divisions":null,"qc_geo":"4","qc_complete":null,"qc_house":null,"history_values":null,"unparsed_parts":null,"source":null,"qc":null}}]}`
