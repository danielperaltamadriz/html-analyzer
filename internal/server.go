package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danielperaltamadriz/home24/internal/models"
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

type VersionResponse struct {
	Number   string `json:"number"`
	IsStrict bool   `json:"is_strict"`
}

type HeadingResponse struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}

type LinkDetailResponse struct {
	URL          string `json:"url"`
	Count        int    `json:"count"`
	IsAccessible bool   `json:"is_accessible"`
}

type LinkTypeResponse struct {
	Total             int `json:"total"`
	TotalAccessible   int `json:"total_accessible"`
	TotalInaccessible int `json:"total_inaccessible"`
	LinkDetails       []LinkDetailResponse
}

type LinksResponse struct {
	Internal LinkTypeResponse `json:"internal"`
	External LinkTypeResponse `json:"external"`
}

type DetailsResponse struct {
	Title    string           `json:"title"`
	Version  *VersionResponse `json:"version,omitempty"`
	Headings HeadingResponse  `json:"headings"`
	Links    LinksResponse    `json:"links"`
}

func (a *API) Shutdown() error {
	fmt.Println("Shutting down server")
	return a.server.Shutdown(context.Background())
}

func (a *API) Start() error {
	fmt.Println("Starting server on port 8080")
	http.HandleFunc("/v1/analyzes", a.HTMLHandler)
	err := a.server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (a *API) HTMLHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received")
	w.Header().Set("Content-Type", "application/json")
	analyzer := NewAnalyzer()
	analyzer.WithSearchSingleElements(
		analyzer.HTMLVersion,
		analyzer.Title,
	)
	analyzer.WithSearchManyElements(
		analyzer.Headings,
		analyzer.Links,
	)
	details, err := analyzer.RunFromURL(r.FormValue("url"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := DetailsResponse{
		Title:    details.Title,
		Version:  mapVersion(details.Version),
		Headings: mapHeadings(details.HeadingsCounter),
		Links:    mapLinks(details.Links),
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func mapVersion(version *models.HTMLVersion) *VersionResponse {
	if version == nil {
		return nil
	}
	return &VersionResponse{
		Number:   string(version.Number),
		IsStrict: version.Strict,
	}
}

func mapHeadings(headings map[models.Heading]int) HeadingResponse {
	var response HeadingResponse
	for tag, count := range headings {
		switch tag {
		case models.H1:
			response.H1 = count
		case models.H2:
			response.H2 = count
		case models.H3:
			response.H3 = count
		case models.H4:
			response.H4 = count
		case models.H5:
			response.H5 = count
		case models.H6:
			response.H6 = count
		}
	}
	return response
}

func mapLinks(links models.Links) LinksResponse {
	return LinksResponse{
		Internal: LinkTypeResponse{
			Total:       links.CountInternalLinks(),
			LinkDetails: mapLinkDetailsResponse(links.GetInternalLinks()),
		},
		External: LinkTypeResponse{
			Total:       links.CountExternalLinks(),
			LinkDetails: mapLinkDetailsResponse(links.GetExternalLinks()),
		},
	}
}

func mapLinkDetailsResponse(links []models.Link) []LinkDetailResponse {
	var response []LinkDetailResponse
	for _, link := range links {
		response = append(response, LinkDetailResponse{
			URL:          link.URL,
			Count:        link.Count,
			IsAccessible: link.Accessible,
		})
	}
	return response
}
