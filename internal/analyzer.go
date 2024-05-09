package internal

import (
	"strings"

	"github.com/danielperaltamadriz/home24/internal/models"
	"golang.org/x/net/html"
)

type SearchElement func(n *html.Node) bool
type analyzer struct {
	result models.HTMLDetails
	url    string

	searchSingleElements []SearchElement
	singleSearchesDone   map[int]bool
	searchManyElements   []SearchElement
}

func NewAnalyzer() *analyzer {
	return &analyzer{
		singleSearchesDone: make(map[int]bool),
	}
}

func (a *analyzer) WithURL(url string) {
	a.url = url
}

func (a *analyzer) WithSearchSingleElements(searchElements ...SearchElement) {
	a.searchSingleElements = searchElements
}

func (a *analyzer) WithSearchManyElements(searchElements ...SearchElement) {
	a.searchManyElements = searchElements
}

func (a *analyzer) Run(node *html.Node) models.HTMLDetails {
	var f func(*html.Node) bool
	f = func(n *html.Node) bool {
		for i, searchElement := range a.searchSingleElements {
			if a.singleSearchesDone[i] {
				continue
			}
			if searchElement(n) {
				a.singleSearchesDone[i] = true
			}
		}
		if len(a.singleSearchesDone) == len(a.searchSingleElements) && len(a.searchManyElements) == 0 {
			return true
		}
		for _, searchElement := range a.searchManyElements {
			searchElement(n)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if f(c) {
				return true
			}
		}
		return false
	}
	f(node)
	return a.result
}

func (a *analyzer) HTMLVersion(n *html.Node) bool {
	if n.Type != html.DoctypeNode {
		return false
	}
	if strings.ToLower(n.Data) != "html" {
		return false
	}
	if len(n.Attr) == 0 {
		a.result.Version = &models.HTMLVersion{
			Number: models.HTMLVersion5,
		}
		return true
	}
	var htmlVersion *models.HTMLVersion
	if strings.Contains(n.Attr[0].Val, "4.01") {
		htmlVersion = &models.HTMLVersion{
			Number: models.HTMLVersion401,
		}
	}
	if (strings.Contains(n.Attr[0].Val, "strict")) || (len(n.Attr) > 1 && strings.Contains(n.Attr[1].Val, "strict")) {
		htmlVersion.Strict = true
	}
	a.result.Version = htmlVersion
	return true
}

func (a *analyzer) Title(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "title" {
		a.result.Title = n.FirstChild.Data
		return true
	}
	return false
}

func (a *analyzer) Headings(n *html.Node) bool {
	if n.Type == html.ElementNode {
		var heading models.Heading
		switch models.Heading(n.Data) {
		case models.H1:
			heading = models.H1
		case models.H2:
			heading = models.H2
		case models.H3:
			heading = models.H3
		case models.H4:
			heading = models.H4
		case models.H5:
			heading = models.H5
		case models.H6:
			heading = models.H6
		default:
			return false
		}
		if a.result.HeadingsCounter == nil {
			a.result.HeadingsCounter = make(map[models.Heading]int)
		}
		a.result.HeadingsCounter[heading]++
		return true
	}
	return false
}

func (a *analyzer) Links(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				if strings.HasPrefix(attr.Val, "http") {
					a.result.Links = a.result.Links.AddExternalLink(attr.Val)
					return true
				}
				a.result.Links = a.result.Links.AddInternalLink(attr.Val)
				return true
			}
		}
	}
	return false
}
