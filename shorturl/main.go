package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"embed"
	"log"
)

//go:embed pb_public/*
var f embed.FS

// Naive approach to prefixing https to the url
// if it's missing from the original url
func addPrefix(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

func main() {
	app := pocketbase.New()

	corsCfg := middleware.DefaultCORSConfig
	corsCfg.AllowOrigins = []string{"http://127.0.1.1:8090/*"}
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		// Serve the index.html page for the root path
		// to enable new URL creation
		e.Router.AddRoute(echo.Route{
			Method:  http.MethodGet,
			Path:    "/",
			Handler: indexHandler(app),
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				// middleware.CORS(),
				middleware.CORSWithConfig(corsCfg),
			},
		})

		// Redirect from short url (the url passphrase) to full url
		e.Router.AddRoute(echo.Route{
			Method:  http.MethodGet,
			Path:    "/:url",
			Handler: fetchUrlHandler(app),
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				middleware.CORS(),
			},
		})

		// Create a new short url from full original url
		e.Router.AddRoute(echo.Route{
			Method:  http.MethodPost,
			Path:    "/api/createurl",
			Handler: createUrlHandler(app),
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				// middleware.CORS(),
				middleware.CORSWithConfig(corsCfg),
			},
		})

		return nil
	})
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
