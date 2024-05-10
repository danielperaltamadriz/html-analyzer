package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/danielperaltamadriz/home24/internal"
	"github.com/danielperaltamadriz/home24/templates"
)

func main() {
	component := templates.Layout()

	http.Handle("/", templ.Handler(component))

	http.HandleFunc("/details", func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")
		apiURL := fmt.Sprintf("http://localhost:8080/v1/analyzes?url=%s", url)

		resp, err := http.Get(apiURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		component.Render(context.Background(), w)

	})

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)

}
