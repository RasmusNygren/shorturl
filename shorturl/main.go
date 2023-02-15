package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
    "github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"

	"github.com/RasmusNygren/go-passphrase/passphrase"

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

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		// Serve the index.html page for the root path
		// to enable new URL creation
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/",
			Handler: func(c echo.Context) error {
				html, err := f.ReadFile("pb_public/index.html")
				if err != nil {
					log.Fatal("Error reading embedded file, this likely depends on a pathing issue or build error")
				}
				return c.HTML(http.StatusOK, string(html))
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				middleware.CORS(),
			},
		})

		// Redirect from short url (the url passphrase) to full url
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:url",
			Handler: func(c echo.Context) error {
				short_url := c.PathParam("url")
				record, err := app.Dao().FindFirstRecordByData("links", "short_url", short_url)

				if err != nil {
					log.Println(err)
					html, err := f.ReadFile("pb_public/404.html")
					if err != nil {
						log.Fatal("Error reading embedded file, this likely depends on a pathing issue or build error")
					}
					return c.HTML(http.StatusNotFound, string(html))
				}

				long_url := record.GetString("long_url")
				return c.Redirect(http.StatusFound, long_url)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				middleware.CORS(),
			},
		})

		// Create a new short url from full original url
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/createurl",
			Handler: func(c echo.Context) error {
				url := c.FormValue("url")
                url = addPrefix(url)
				log.Println("Provided long url", url)

                {
                    record, err := app.Dao().FindFirstRecordByData("links", "long_url", url)
                    // Consider doing some logging if error is not null
                    log.Println("record", record, "error", err)
                    if record != nil && err == nil {
                        url_phrase := record.GetString("short_url")
                        return c.String(http.StatusOK, url_phrase)

                    }
                }


				wordlist := passphrase.EffSmallShortWords()
				url_phrase := passphrase.GeneratePhrase(wordlist, 2)

				// We do not allow duplicate url phrases so check that the 
                // url phrase is unique. If it is not unique, regenerate a new phrase
				// until it is.
				for {
                    record, err := app.Dao().FindFirstRecordByData("links", "short_url", url_phrase)
                    
					if err != nil {
						log.Println(err)
					}
                    if record == nil {
                        break
                    }
					url_phrase = passphrase.GeneratePhrase(wordlist, 2)
				}

				collection, err := app.Dao().FindCollectionByNameOrId("links")
				if err != nil {
					log.Fatal(err)
				}
                record := models.NewRecord(collection)
				form := forms.NewRecordUpsert(app, record)

				// Ensure that all long urls begin with https:// or http://
				form.LoadData(map[string]any{
					"long_url":  url,
					"short_url": url_phrase,
				})
				if err := form.Submit(); err != nil {
					log.Println(err)
					return c.String(http.StatusBadRequest, "Invalid URL")
				}
				return c.String(http.StatusOK, url_phrase)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				middleware.CORS(),
			},
		})

		return nil
	})
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
