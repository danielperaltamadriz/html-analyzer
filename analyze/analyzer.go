package analyze

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/danielperaltamadriz/home24/analyze/models"
	"golang.org/x/net/html"
)

const (
	_defaultTimeout = 2 * time.Second
)

type SearchElement func(n *html.Node) bool
type Analyzer struct {
	result models.HTMLDetails
	url    string
	ctx    context.Context
	req    *http.Request
	node   *html.Node

	searchSingleElements []SearchElement
	singleSearchesDone   map[int]bool
	searchManyElements   []SearchElement

	verifyLinkFunc func(l *models.Link) bool
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		singleSearchesDone: make(map[int]bool),
	}
}

func (a *Analyzer) WithSearchSingleElements(searchElements ...SearchElement) {
	a.searchSingleElements = searchElements
}

func (a *Analyzer) WithSearchManyElements(searchElements ...SearchElement) {
	a.searchManyElements = searchElements
}

func (a *Analyzer) run(node *html.Node) models.HTMLDetails {
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
	a.ValidateLinks()
	return a.result
}

func (a *Analyzer) ValidateLinks() {
	var wg sync.WaitGroup
	for _, link := range a.result.Links {
		if link == nil {
			continue
		}
		wg.Add(1)
		go func(l *models.Link) {
			defer wg.Done()
			if a.verifyLinkFunc != nil {
				l.Accessible = a.verifyLinkFunc(l)
				return
			}
			l.VerifyLink()
		}(link)
	}
	wg.Wait()
}

func (a *Analyzer) RunFromURL(url string) (*models.HTMLDetails, error) {
	a.url = url
	if err := a.requestHTML(); err != nil {
		return nil, err
	}
	details := a.run(a.node)
	return &details, nil
}

func (a *Analyzer) requestHTML() error {
	ctx := a.ctx
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), _defaultTimeout)
		defer cancel()
	}
	resp, err := a.doRequest(ctx)
	if err != nil {
		return fmt.Errorf("doRequest: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return models.NewErrorWithStatusCode(models.ErrInvalidRequest, "invalid status code", resp.StatusCode)
	}

	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return fmt.Errorf("invalid content type")
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse html: %w", err)
	}
	a.node = doc
	return nil
}

func (a *Analyzer) doRequest(ctx context.Context) (*http.Response, error) {
	url, err := url.ParseRequestURI(a.url)
	if err != nil {
		return nil, models.NewError(models.ErrTypeInvalidURL, "invalid url")
	}
	if url.Scheme == "" || url.Host == "" {
		return nil, models.NewError(models.ErrTypeInvalidURL, "invalid url")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, models.NewError(models.ErrTypeInvalidURL, "invalid request")
	}
	a.req = req
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, models.NewError(models.ErrTypeInvalidURL, "failed to get html file")
	}
	return resp, nil
}

func (a *Analyzer) HTMLVersion(n *html.Node) bool {
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

func (a *Analyzer) Title(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "title" {
		a.result.Title = n.FirstChild.Data
		return true
	}
	return false
}

func (a *Analyzer) Headings(n *html.Node) bool {
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

func (a *Analyzer) HasLoginForm(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "form" {
		formAnalyzer := NewAnalyzer()
		var hasPassword bool
		hasPasswordFunc := func(n *html.Node) bool {
			if n.Type == html.ElementNode && n.Data == "input" {
				for _, attr := range n.Attr {
					if attr.Key == "type" && attr.Val == "password" {
						hasPassword = true
						return true
					}
				}
			}
			return false
		}

		formAnalyzer.WithSearchSingleElements(hasPasswordFunc)
		formAnalyzer.run(n)

		if hasPassword {
			a.result.HasLoginForm = true
			return true
		}
	}
	return false
}

func (a *Analyzer) Links(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				if strings.HasPrefix(attr.Val, "http") {
					a.result.Links = a.result.Links.AddExternalLink(attr.Val)
					return true
				}
				url := a.req.URL.String() + "/" + attr.Val
				if strings.HasPrefix(attr.Val, "#") {
					url = a.req.URL.String() + attr.Val
				}
				if strings.HasPrefix(attr.Val, "/") {
					url = a.req.URL.Scheme + "://" + a.req.Host + attr.Val
				}
				a.result.Links = a.result.Links.AddInternalLink(url)
				return true
			}
		}
	}
	return false
}

func (a *Analyzer) WithLinkVerifierFunc(verifyFunc func(l *models.Link) bool) {
	a.verifyLinkFunc = verifyFunc
}
