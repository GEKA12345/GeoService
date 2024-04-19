// GeoService
//
// # This is a Geo Service API
//
// info:
//
//	Version: 0.0.1
//	title: GeoService
//	description: This is a Geo Service API
//
// Schemes: http
//
//	Host: localhost:8080
//	BasePath:
//
// Consumes:
// - application/json
// Produces:
// - application/json
//
// swagger:meta
package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"test/proxy/internal/controller"
	"test/proxy/internal/responder"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/ptflp/godecoder"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate swagger generate spec -o ./docs/swagger.json --scan-models
type Router struct {
	r *chi.Mux
	c controller.Controllerer
}

func init() {
	controller.TokenAuth = jwtauth.New("HS256", []byte("salt_01"), nil)
}

func (router *Router) handleRoutes() {

	// swagger:operation POST /api/login user postLoginUser
	//
	// Login user
	//
	// ---
	// parameters:
	//   - name: userRequest
	//     in: body
	//     required: true
	//     schema:
	//       $ref: "#/definitions/userRequest"
	// responses:
	//   "200":
	//     description: successfully logged in
	//     in: body
	//     schema:
	//       $ref: "#/definitions/tokenResponse"
	//   "400":
	//     description: bad request
	//     in: body
	//     schema:
	//       $ref: "#/definitions/errorResponse"
	//   "404":
	//     description: user not found or wrong password
	//     in: body
	//     schema:
	//       $ref: "#/definitions/errorResponse"
	//   "500":
	//     description: internal server error
	//     in: body
	//     schema:
	//       $ref: "#/definitions/errorResponse"
	router.r.HandleFunc("/api/login", router.c.Login)

	// swagger:operation POST /api/register user postRegisterUser
	//
	// Register user
	//
	// ---
	// parameters:
	//   - name: userRequest
	//     in: body
	//     required: true
	//     schema:
	//       $ref: "#/definitions/userRequest"
	// responses:
	//   "200":
	//     description: successfully registered
	//     in: body
	//     schema:
	//       $ref: "#/definitions/tokenResponse"
	//   "400":
	//     description: bad request
	//     in: body
	//     schema:
	//       $ref: "#/definitions/errorResponse"
	//   "409":
	//     description: user already exists
	//     in: body
	//     schema:
	//       $ref: "#/definitions/errorResponse"
	//   "500":
	//     description: internal server error
	//     in: body
	//     schema:
	//       $ref: "#/definitions/errorResponse"
	router.r.HandleFunc("/api/register", router.c.Register)

	router.r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(controller.TokenAuth))

		r.Use(router.c.Authenticator(controller.TokenAuth))

		// swagger:operation POST /api/address/search search postSearch
		//
		// Search for addresses by query string
		//
		// ---
		// parameters:
		//   - name: query
		//     in: body
		//     required: true
		//     schema:
		//       $ref: "#/definitions/searchRequest"
		//   - name: Authorization
		//     in: header
		//     type: string
		//     required: true
		//     description: Bearer token for user authentication
		//     example: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ
		// responses:
		//   "200":
		//     description: search results
		//     in: body
		//     schema:
		//       $ref: "#/definitions/searchResponse"
		//   "400":
		//     description: bad request
		//     in: body
		//     schema:
		//       $ref: "#/definitions/errorResponse"
		//   "403":
		//     description: forbidden
		//     in: body
		//     schema:
		//       $ref: "#/definitions/errorResponse"
		//   "500":
		//     description: internal server error
		//     in: body
		//     schema:
		//       $ref: "#/definitions/errorResponse"
		r.HandleFunc("/api/address/search", router.c.GeoSearch)

		// swagger:operation POST /api/address/geocode geoCode postGeo
		//
		// Search for addresses by longitude and latitude
		//
		// ---
		// parameters:
		//   - name: query
		//     in: body
		//     required: true
		//     schema:
		//       $ref: "#/definitions/geocodeRequest"
		//   - name: Authorization
		//     in: header
		//     type: string
		//     required: true
		//     description: Bearer token for user authentication
		//     example: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ
		// responses:
		//   "200":
		//     description: geoCode results
		//     in: body
		//     schema:
		//       $ref: "#/definitions/geocodeResponse"
		//   "400":
		//     description: bad request
		//     in: body
		//     schema:
		//       $ref: "#/definitions/errorResponse"
		//   "403":
		//     description: forbidden
		//     in: body
		//     schema:
		//       $ref: "#/definitions/errorResponse"
		//   "500":
		//     description: internal server error
		//     in: body
		//     schema:
		//       $ref: "#/definitions/errorResponse"
		r.HandleFunc("/api/address/geocode", router.c.GeoCode)
	})
}

func main() {
	host := "http://hugo"
	port := ":1313"
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))
	decoder := godecoder.NewDecoder()
	respond := responder.NewResponder(decoder, logger)
	contrl := controller.NewController(respond, decoder)

	r := getProxyRouter(host, port, contrl)
	http.ListenAndServe(":8080", r.r)
}

func getProxyRouter(host, port string, contrl controller.Controllerer) *Router {
	router := &Router{r: chi.NewRouter(), c: contrl}

	router.r.Use(NewReverseProxy(host, port, contrl).ReverseProxy)

	router.handleRoutes()

	return router
}

type ReverseProxy struct {
	host string
	port string
	c    controller.Controllerer
}

func NewReverseProxy(host, port string, c controller.Controllerer) *ReverseProxy {
	return &ReverseProxy{
		host: host,
		port: port,
		c:    c,
	}
}

func (rp *ReverseProxy) ReverseProxy(next http.Handler) http.Handler {
	reverseProxyURL, _ := url.Parse(rp.host + rp.port)
	proxy := httputil.NewSingleHostReverseProxy(reverseProxyURL)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/docs") {
			http.ServeFile(w, r, "./docs/swagger.json")
			return
		}
		if strings.HasPrefix(r.URL.Path, "/swagger") {
			rp.c.SwaggerUI(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api") {
			next.ServeHTTP(w, r)
			return
		}
		proxy.ServeHTTP(w, r)
	})
}
