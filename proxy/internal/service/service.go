package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"test/proxy/internal/responder"

	"github.com/go-chi/jwtauth/v5"
	"github.com/ptflp/godecoder"
	"golang.org/x/crypto/bcrypt"
)

var TokenAuth *jwtauth.JWTAuth

var (
	TestEnabled    = false
	TestGeoHost    = "http://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address"
	TestSearchHost = "https://cleaner.dadata.ru/api/v1/clean/address"
)

type GeoServicer interface {
	IsUserExist(login string) bool
	Register(login, pasw string) string
	Login(login, pasw string) (string, bool)
	GetGeoResp(lat, lon string) (*responder.GeocodeResponse, error)
	GetSearchResp(query string) (*responder.SearchResponse, error)
}

type User map[string]interface{}

type GeoService struct {
	Users map[string]User
	godecoder.Decoder
}

func NewGeoService(decoder godecoder.Decoder) GeoServicer {
	return &GeoService{Users: make(map[string]User), Decoder: decoder}
}

func (g *GeoService) IsUserExist(login string) bool {
	_, ok := g.Users[login]
	return ok
}

func (g *GeoService) Register(login, passw string) string {
	pass, _ := bcrypt.GenerateFromPassword([]byte(passw), 0)

	g.Users[login] = User{
		"login":    login,
		"password": string(pass),
	}

	_, tokenString, _ := TokenAuth.Encode(g.Users[login])
	return tokenString
}

func (g *GeoService) Login(login, passw string) (string, bool) {
	user, ok := g.Users[login]
	if !ok {
		return "", false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user["password"].(string)), []byte(passw)); err != nil {
		return "", false
	}

	_, tokenString, _ := TokenAuth.Encode(user)
	return tokenString, true
}

const searchHost = "https://cleaner.dadata.ru/api/v1/clean/address"

func (g *GeoService) GetSearchResp(query string) (*responder.SearchResponse, error) {
	client := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf(`[ "%s" ]`, query))

	host := searchHost
	if TestEnabled {
		host = TestSearchHost
	}

	req, _ := http.NewRequest("POST", host, data)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token 62221a61a6c6f89397432e67dc434135ebda706e")
	req.Header.Set("X-Secret", "3298c7039948814bf8fdcd051e300983a5a3c000")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error request dadata.ru/api: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status %v dadata.ru/api", resp.StatusCode)
	}

	addrS := make(Addresses, 0)
	err = g.Decode(resp.Body, &addrS)
	if err != nil {
		return nil, fmt.Errorf("error decode response dadata.ru/api: %v", err)
	}

	addrSearch := &responder.SearchResponse{Addresses: make([]*responder.Address, len(addrS))}
	for i, v := range addrS {
		tempAddr := responder.Address{Address: v.Result}
		tempAddr.Lat, _ = strconv.ParseFloat(v.GeoLat, 64)

		tempAddr.Lon, _ = strconv.ParseFloat(v.GeoLon, 64)

		addrSearch.Addresses[i] = &tempAddr
	}

	return addrSearch, nil
}

const geoHost = "http://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address"

func (g *GeoService) GetGeoResp(lat, lon string) (*responder.GeocodeResponse, error) {
	client := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf(`{ "lat": %v, "lon": %v }`, lat, lon))

	host := geoHost
	if TestEnabled {
		host = TestGeoHost
	}

	req, _ := http.NewRequest("POST", host, data)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token 62221a61a6c6f89397432e67dc434135ebda706e")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error request dadata.ru/api: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status %v dadata.ru/api", resp.StatusCode)
	}

	addrS := GeoAddresses{}
	err = g.Decode(resp.Body, &addrS)
	if err != nil {
		return nil, fmt.Errorf("error decode response dadata.ru/api: %v", err)
	}

	addrSearch := &responder.GeocodeResponse{Addresses: make([]*responder.Address, len(addrS.Suggestions))}
	for i, v := range addrS.Suggestions {
		tempAddr := responder.Address{Address: v.Value}
		tempAddr.Lat, _ = strconv.ParseFloat(v.Data.GeoLat, 64)

		tempAddr.Lon, _ = strconv.ParseFloat(v.Data.GeoLon, 64)

		addrSearch.Addresses[i] = &tempAddr
	}

	return addrSearch, nil
}

type Addresses []respSearch

type respSearch struct {
	Source       string `json:"source"`
	Result       string `json:"result"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	Region       string `json:"region"`
	CityArea     string `json:"city_area"`
	CityDistrict string `json:"city_district"`
	Street       string `json:"street"`
	House        string `json:"house"`
	GeoLat       string `json:"geo_lat"`
	GeoLon       string `json:"geo_lon"`
	QcGeo        int64  `json:"qc_geo"`
}

type GeoAddresses struct {
	Suggestions []Suggestion `json:"suggestions"`
}

type Suggestion struct {
	Value             string `json:"value"`
	UnrestrictedValue string `json:"unrestricted_value"`
	Data              Data   `json:"data"`
}

type Data struct {
	Area                 interface{} `json:"area"`
	AreaFiasID           interface{} `json:"area_fias_id"`
	AreaKladrID          interface{} `json:"area_kladr_id"`
	AreaType             interface{} `json:"area_type"`
	AreaTypeFull         interface{} `json:"area_type_full"`
	AreaWithType         interface{} `json:"area_with_type"`
	BeltwayDistance      interface{} `json:"beltway_distance"`
	BeltwayHit           interface{} `json:"beltway_hit"`
	Block                interface{} `json:"block"`
	BlockType            interface{} `json:"block_type"`
	BlockTypeFull        interface{} `json:"block_type_full"`
	CapitalMarker        string      `json:"capital_marker"`
	City                 string      `json:"city"`
	CityArea             string      `json:"city_area"`
	CityDistrict         interface{} `json:"city_district"`
	CityDistrictFiasID   interface{} `json:"city_district_fias_id"`
	CityDistrictKladrID  interface{} `json:"city_district_kladr_id"`
	CityDistrictType     interface{} `json:"city_district_type"`
	CityDistrictTypeFull interface{} `json:"city_district_type_full"`
	CityDistrictWithType interface{} `json:"city_district_with_type"`
	CityFiasID           string      `json:"city_fias_id"`
	CityKladrID          string      `json:"city_kladr_id"`
	CityType             string      `json:"city_type"`
	CityTypeFull         string      `json:"city_type_full"`
	CityWithType         string      `json:"city_with_type"`
	Country              string      `json:"country"`
	CountryIsoCode       string      `json:"country_iso_code"`
	Divisions            interface{} `json:"divisions"`
	Entrance             interface{} `json:"entrance"`
	FederalDistrict      string      `json:"federal_district"`
	FiasActualityState   string      `json:"fias_actuality_state"`
	FiasCode             interface{} `json:"fias_code"`
	FiasID               string      `json:"fias_id"`
	FiasLevel            string      `json:"fias_level"`
	Flat                 interface{} `json:"flat"`
	FlatArea             interface{} `json:"flat_area"`
	FlatCadnum           interface{} `json:"flat_cadnum"`
	FlatFiasID           interface{} `json:"flat_fias_id"`
	FlatPrice            interface{} `json:"flat_price"`
	FlatType             interface{} `json:"flat_type"`
	FlatTypeFull         interface{} `json:"flat_type_full"`
	Floor                interface{} `json:"floor"`
	GeoLat               string      `json:"geo_lat"`
	GeoLon               string      `json:"geo_lon"`
	GeonameID            string      `json:"geoname_id"`
	HistoryValues        interface{} `json:"history_values"`
	House                string      `json:"house"`
	HouseCadnum          interface{} `json:"house_cadnum"`
	HouseFiasID          string      `json:"house_fias_id"`
	HouseKladrID         string      `json:"house_kladr_id"`
	HouseType            string      `json:"house_type"`
	HouseTypeFull        string      `json:"house_type_full"`
	KladrID              string      `json:"kladr_id"`
	Metro                interface{} `json:"metro"`
	Okato                string      `json:"okato"`
	Oktmo                string      `json:"oktmo"`
	PostalBox            interface{} `json:"postal_box"`
	PostalCode           string      `json:"postal_code"`
	Qc                   interface{} `json:"qc"`
	QcComplete           interface{} `json:"qc_complete"`
	QcGeo                string      `json:"qc_geo"`
	QcHouse              interface{} `json:"qc_house"`
	Region               string      `json:"region"`
	RegionFiasID         string      `json:"region_fias_id"`
	RegionIsoCode        string      `json:"region_iso_code"`
	RegionKladrID        string      `json:"region_kladr_id"`
	RegionType           string      `json:"region_type"`
	RegionTypeFull       string      `json:"region_type_full"`
	RegionWithType       string      `json:"region_with_type"`
	Room                 interface{} `json:"room"`
	RoomCadnum           interface{} `json:"room_cadnum"`
	RoomFiasID           interface{} `json:"room_fias_id"`
	RoomType             interface{} `json:"room_type"`
	RoomTypeFull         interface{} `json:"room_type_full"`
	Settlement           interface{} `json:"settlement"`
	SettlementFiasID     interface{} `json:"settlement_fias_id"`
	SettlementKladrID    interface{} `json:"settlement_kladr_id"`
	SettlementType       interface{} `json:"settlement_type"`
	SettlementTypeFull   interface{} `json:"settlement_type_full"`
	SettlementWithType   interface{} `json:"settlement_with_type"`
	Source               interface{} `json:"source"`
	SquareMeterPrice     interface{} `json:"square_meter_price"`
	Stead                interface{} `json:"stead"`
	SteadCadnum          interface{} `json:"stead_cadnum"`
	SteadFiasID          interface{} `json:"stead_fias_id"`
	SteadType            interface{} `json:"stead_type"`
	SteadTypeFull        interface{} `json:"stead_type_full"`
	Street               string      `json:"street"`
	StreetFiasID         string      `json:"street_fias_id"`
	StreetKladrID        string      `json:"street_kladr_id"`
	StreetType           string      `json:"street_type"`
	StreetTypeFull       string      `json:"street_type_full"`
	StreetWithType       string      `json:"street_with_type"`
	TaxOffice            string      `json:"tax_office"`
	TaxOfficeLegal       string      `json:"tax_office_legal"`
	Timezone             interface{} `json:"timezone"`
	UnparsedParts        interface{} `json:"unparsed_parts"`
}
