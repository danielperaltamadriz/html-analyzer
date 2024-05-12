package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danielperaltamadriz/home24/analyze"
)

const (
	_defaultPort = 8080
)

type API struct {
	server *http.Server
}

type APIConfig struct {
	Port int
}

func NewAPI(cfg APIConfig) *API {
	if cfg.Port == 0 {
		cfg.Port = _defaultPort
	}
	return &API{
		server: &http.Server{
			Addr: fmt.Sprintf(":%d", cfg.Port),
		},
	}
}

func (a *API) Shutdown() error {
	fmt.Println("Shutting down server")
	return a.server.Shutdown(context.Background())
}

func (a *API) Start() error {
	fmt.Println("Starting server on port " + a.server.Addr)
	http.HandleFunc("/v1/analyzes", a.HTMLHandler)
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server.ListenAndServe: %w", err)
	}
	return nil
}

func (a *API) HTMLHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received")
	w.Header().Set("Content-Type", "application/json")
	analyzer := analyze.NewAnalyzer()
	analyzer.WithSearchSingleElements(analyzer.HTMLVersion, analyzer.Title, analyzer.HasLoginForm)
	analyzer.WithSearchManyElements(analyzer.Headings, analyzer.Links)
	url := r.FormValue("url")
	details, err := analyzer.RunFromURL(url)
	if err != nil {
		fmt.Printf("analyzer.RunFromURL, url: %s, error: %s \n", url, err.Error())
		mapError(w, err)
		return
	}
	response := DetailsResponse{
		Title:        details.Title,
		Version:      mapVersion(details.Version),
		Headings:     mapHeadings(details.HeadingsCounter),
		Links:        mapLinks(details.Links),
		HasLoginForm: details.HasLoginForm,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println("failed to encode response: ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
