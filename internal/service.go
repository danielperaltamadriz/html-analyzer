package internal

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/danielperaltamadriz/home24/internal/models"
	"golang.org/x/net/html"
)

type linkVerifier interface {
	IsAccessible(url string) bool
}

type HTMLService struct {
	LinkVerifier linkVerifier
}

func NewHTMLService() *HTMLService {
	return &HTMLService{}
}

func (a *HTMLService) SetLinkVerifier(linkVerifier linkVerifier) {
	a.LinkVerifier = linkVerifier
}

func (a *HTMLService) Get(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get url: %w", err)
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("invalid content type")
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	return doc, nil
}

func (a *HTMLService) GetDetails(doc *html.Node) models.HTMLDetails {
	analyzer := NewAnalyzer()
	analyzer.WithSearchSingleElements(
		analyzer.HTMLVersion,
		analyzer.Title,
	)
	analyzer.WithSearchManyElements(
		analyzer.Headings,
		analyzer.Links,
	)
	return analyzer.Run(doc)
}
