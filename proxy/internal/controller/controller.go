package controller

import (
	"html/template"
	"net/http"
	"time"

	"test/proxy/internal/responder"
	"test/proxy/internal/service"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/ptflp/godecoder"
)

type Controllerer interface {
	Register(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
	GeoSearch(http.ResponseWriter, *http.Request)
	GeoCode(http.ResponseWriter, *http.Request)
	Authenticator(*jwtauth.JWTAuth) func(http.Handler) http.Handler
	SwaggerUI(http.ResponseWriter, *http.Request)
}

type Controller struct {
	service service.GeoServicer
	responder.Responder
	godecoder.Decoder
}

func NewController(resp responder.Responder, decod godecoder.Decoder, service service.GeoServicer) Controllerer {
	return &Controller{Responder: resp, Decoder: decod, service: service}
}

func (c *Controller) Authenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())

			if err != nil || token == nil || jwt.Validate(token, ja.ValidateOptions()...) != nil {
				c.ErrorForbidden(w)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.ErrorNotAllowed(w)
		return
	}

	userInput := &UserRequest{}
	err := c.Decode(r.Body, userInput)
	if err != nil {
		c.ErrorBadRequest(w, err)
		return
	}

	if ok := c.service.IsUserExist(userInput.Login); ok {
		c.ErrorUserConflict(w)
		return
	}

	tokenString := c.service.Register(userInput.Login, userInput.Password)

	token := responder.TokenResponse{AccessToken: "Bearer " + tokenString}
	c.OutputJSON(w, token)
}

// swagger:model userRequest
type UserRequest struct {
	// user login
	//
	// example: user1
	Login string `json:"login"`
	// user password
	//
	// example: qwerty
	Password string `json:"password"`
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.ErrorNotAllowed(w)
		return
	}

	userInput := &UserRequest{}
	err := c.Decode(r.Body, userInput)
	if err != nil {
		c.ErrorBadRequest(w, err)
		return
	}

	tokenString, ok := c.service.Login(userInput.Login, userInput.Password)
	if !ok {
		c.ErrorUserNotFound(w)
		return
	}

	token := responder.TokenResponse{AccessToken: "Bearer " + tokenString}
	c.OutputJSON(w, token)
}

func (c *Controller) GeoSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.ErrorNotAllowed(w)
		return
	}

	reqInput := &SearchRequest{}
	err := c.Decode(r.Body, reqInput)
	if err != nil {
		c.ErrorBadRequest(w, err)
		return
	}

	addrSearch, err := c.service.GetSearchResp(reqInput.Query)
	if err != nil {
		c.ErrorInternal(w, err)
		return
	}

	c.OutputJSON(w, addrSearch)
}

func (c *Controller) GeoCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.ErrorNotAllowed(w)
		return
	}

	reqInput := &GeocodeRequest{}
	err := c.Decode(r.Body, reqInput)
	if err != nil {
		c.ErrorBadRequest(w, err)
		return
	}

	addrGeoCode, err := c.service.GetGeoResp(reqInput.Lat, reqInput.Lng)
	if err != nil {
		c.ErrorInternal(w, err)
		return
	}

	c.OutputJSON(w, addrGeoCode)
}

// swagger:model searchRequest
type SearchRequest struct {
	// searching address query
	//
	// required: true
	// min length: 2
	// example: Москва
	Query string `json:"query"`
}

// swagger:model geocodeRequest
type GeocodeRequest struct {
	// point latitude
	//
	// required: true
	// example: 55.7522
	Lat string `json:"lat"`
	// point longitude
	//
	// required: true
	// example: 37.6156
	Lng string `json:"lng"`
}

const (
	swaggerTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <script src="//unpkg.com/swagger-ui-dist@3/swagger-ui-standalone-preset.js"></script>
    <!-- <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.22.1/swagger-ui-standalone-preset.js"></script> -->
    <script src="//unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
    <!-- <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.22.1/swagger-ui-bundle.js"></script> -->
    <link rel="stylesheet" href="//unpkg.com/swagger-ui-dist@3/swagger-ui.css" />
    <!-- <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.22.1/swagger-ui.css" /> -->
	<style>
		body {
			margin: 0;
		}
	</style>
    <title>Swagger</title>
</head>
<body>
    <div id="swagger-ui"></div>
    <script>
        window.onload = function() {
          SwaggerUIBundle({
            url: "/docs/swagger.json?{{.Time}}",
            dom_id: '#swagger-ui',
            presets: [
              SwaggerUIBundle.presets.apis,
              SwaggerUIStandalonePreset
            ],
            layout: "StandaloneLayout"
          })
        }
    </script>
</body>
</html>
`
)

func (c *Controller) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, err := template.New("swagger").Parse(swaggerTemplate)
	if err != nil {
		return
	}
	err = tmpl.Execute(w, struct {
		Time int64
	}{
		Time: time.Now().Unix(),
	})
	if err != nil {
		return
	}
}
