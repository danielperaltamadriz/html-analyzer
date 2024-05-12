package api

import (
	"errors"
	"net/http"

	"github.com/danielperaltamadriz/home24/analyze/models"
)

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
	Title        string           `json:"title"`
	Version      *VersionResponse `json:"version,omitempty"`
	Headings     HeadingResponse  `json:"headings"`
	Links        LinksResponse    `json:"links"`
	HasLoginForm bool             `json:"hasLoginForm"`
}

func mapError(w http.ResponseWriter, err error) {
	var e *models.Error
	errUnwrap := errors.Unwrap(err)
	if errUnwrap != nil {
		err = errUnwrap
	}
	if errors.As(err, &e) {
		switch e.Type {
		case models.ErrInvalidRequest:
			w.WriteHeader(e.ResponseStatusCode)
			return
		case models.ErrTypeInvalidURL:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusInternalServerError)

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
			Total:             links.CountInternalLinks(),
			TotalAccessible:   links.CountInternalLinksAccessible(),
			TotalInaccessible: links.CountInternalLinksInaccessible(),
			LinkDetails:       mapLinkDetailsResponse(links.GetInternalLinks()),
		},
		External: LinkTypeResponse{
			Total:             links.CountExternalLinks(),
			TotalAccessible:   links.CountExternalLinksAccessible(),
			TotalInaccessible: links.CountExternalLinksInaccessible(),
			LinkDetails:       mapLinkDetailsResponse(links.GetExternalLinks()),
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
