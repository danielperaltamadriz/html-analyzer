package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/a-h/templ"
	"github.com/danielperaltamadriz/home24/internal"
	"github.com/danielperaltamadriz/home24/templates"
)

const (
	_defaultAPIHost = "http://localhost:8080"
	_defaultPort    = 3000
)

type config struct {
	apiHost string
	Port    int
}

func loadConfig() config {
	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = _defaultAPIHost
	}

	port := os.Getenv("PORT")
	portInt, err := strconv.Atoi(port)
	if err != nil {
		portInt = _defaultPort
	}

	return config{
		apiHost: apiHost,
		Port:    portInt,
	}
}

func main() {
	component := templates.Layout()

	cfg := loadConfig()

	http.Handle("/", templ.Handler(component))

	http.HandleFunc("/details", func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")
		apiURL := fmt.Sprintf(cfg.apiHost+"/v1/analyzes?url=%s", url)

		resp, err := http.Get(apiURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode >= 400 {
			var errMessage templates.ErrorMessage
			switch resp.StatusCode {
			case http.StatusBadRequest:
				errMessage = templates.ErrorMessage{
					Message: "The URL requested is not valid",
				}
			case http.StatusNotFound:
				errMessage = templates.ErrorMessage{
					Message: "The URL requested was not found",
				}
			default:
				errMessage = templates.ErrorMessage{
					Message: "An error occurred while processing the request with the status code: " + strconv.Itoa(resp.StatusCode),
				}
			}
			errMessage.URL = url
			component = templates.ErrorsTemplate(errMessage)
			err = component.Render(context.Background(), w)
			if err != nil {
				fmt.Printf("Failed to render component: %v", err)
			}
			return
		}
		defer resp.Body.Close()
		var details *internal.DetailsResponse
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(body, &details)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		headings := []templates.Heading{
			{Key: "H1", Count: strconv.Itoa(details.Headings.H1)},
			{Key: "H2", Count: strconv.Itoa(details.Headings.H2)},
			{Key: "H3", Count: strconv.Itoa(details.Headings.H3)},
			{Key: "H4", Count: strconv.Itoa(details.Headings.H4)},
			{Key: "H5", Count: strconv.Itoa(details.Headings.H5)},
			{Key: "H6", Count: strconv.Itoa(details.Headings.H6)},
		}

		d := templates.Details{
			URL:         url,
			Title:       details.Title,
			HTMLVersion: details.Version.Number,
			Headings:    headings,
			Links: templates.Links{
				InternalTotal:     details.Links.Internal.Total,
				ExternalTotal:     details.Links.External.Total,
				InaccessibleTotal: details.Links.Internal.TotalInaccessible + details.Links.External.TotalInaccessible,
			},
			HasLoginForm: details.HasLoginForm,
		}

		component = templates.DetailsTemplate(d)
		err = component.Render(context.Background(), w)
		if err != nil {
			fmt.Printf("Failed to render component: %v", err)
		}

	})

	port := strconv.Itoa(cfg.Port)
	fmt.Println("Listening on :" + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Failed to start server: %v", err)
	}
}
