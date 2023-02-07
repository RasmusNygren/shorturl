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

	"log"
)

// TODO: Handle collisions when the generated phrase already exists.
// TODO: All links has to be prepended with http if not provided as
// redirects otherwise fail.

// Naive approach to prefixing https to the url if it's missing from the
// original url
func addPrefix(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// Redirect from short url (the url passphrase) to full url
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:url",
			Handler: func(c echo.Context) error {
				short_url := c.PathParam("url")
				record, err := app.Dao().FindFirstRecordByData("links", "short_url", short_url)

				if err != nil {
					log.Fatal(err)
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

				collection, err := app.Dao().FindCollectionByNameOrId("links")
				if err != nil {
					log.Fatal(err)
				}

				wordlist := passphrase.EffSmallShortWords()
				url_phrase := passphrase.GeneratePhrase(wordlist, 2)
				record, err := app.Dao().FindFirstRecordByData("links", "short_url", url_phrase)

				// We do not allow duplicate url phrases so regenerate a new phrase
				// if the previous one already exists.
				for record != nil {
					url_phrase = passphrase.GeneratePhrase(wordlist, 2)
					record, err = app.Dao().FindFirstRecordByData("links", "short_url", url_phrase)
					if err != nil {
						log.Fatal(err)
					}
				}

				record = models.NewRecord(collection)
				form := forms.NewRecordUpsert(app, record)

				// Ensure that all long urls begin with https:// or http://
				long_url := addPrefix(url)
				form.LoadData(map[string]any{
					"long_url":  long_url,
					"short_url": url_phrase,
				})
				if err := form.Submit(); err != nil {
					return err
				}
				if err := app.Dao().SaveRecord(record); err != nil {
					return err
				}
				return c.String(http.StatusOK, url_phrase)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		return nil
	})
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
