package internal_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/danielperaltamadriz/home24/internal"
	"github.com/danielperaltamadriz/home24/internal/models"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/html"
)

var (
	HTMLVersion5 = &models.HTMLVersion{
		Number: models.HTMLVersion5,
	}

	HTMLVersion401_STRICT = &models.HTMLVersion{
		Number: models.HTMLVersion401,
		Strict: true,
	}
)

type serviceTestSuite struct {
	suite.Suite
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(serviceTestSuite))
}

func (suite *serviceTestSuite) TestGetValidHTMLUsingRealServer() {
	testCases := []struct {
		name     string
		inputURL string
	}{
		{
			name:     "Get HTML from URL",
			inputURL: "https://scrapeme.live/shop/",
		},
	}

	service := internal.NewHTMLService()
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result, err := service.Get(tc.inputURL)
			suite.NoError(err)
			suite.NotNil(result)
		})
	}
}

func (suite *serviceTestSuite) TestGetValidHTMLUsingFakeServer() {
	testCases := []struct {
		name            string
		fakeServerSetup *setupHTTPTest
	}{
		{
			name: "Get HTML from URL",
			fakeServerSetup: &setupHTTPTest{
				statusCode:   http.StatusOK,
				htmlFilePath: "testdata/html.html",
				contentType:  "text/html",
			},
		},
	}

	service := internal.NewHTMLService()
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			fakeServer := suite.httptestSetup(*tc.fakeServerSetup)
			defer fakeServer.Close()
			result, err := service.Get(fakeServer.URL)
			suite.NoError(err)
			suite.NotNil(result)
		})
	}
}

func (suite *serviceTestSuite) TestGetInvalidHTML() {
	testCases := []struct {
		name             string
		fakeServerSetup  *setupHTTPTest
		errMessageString string
	}{
		{
			name: "Get JSON from URL",
			fakeServerSetup: &setupHTTPTest{
				statusCode:   http.StatusOK,
				htmlFilePath: "testdata/file.json",
				contentType:  "application/json",
			},
			errMessageString: "invalid content type",
		},
	}

	service := internal.NewHTMLService()
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			fakeServer := suite.httptestSetup(*tc.fakeServerSetup)
			defer fakeServer.Close()
			result, err := service.Get(fakeServer.URL)
			suite.Nil(result)
			suite.ErrorContains(err, tc.errMessageString)
		})
	}
}

func (suite *serviceTestSuite) TestGetHTMLDetails() {
	testCases := []struct {
		name     string
		htmlPath string

		expectedDetails models.HTMLDetails
	}{
		{
			name:     " get title from html",
			htmlPath: "testdata/title.html",

			expectedDetails: models.HTMLDetails{
				Version: HTMLVersion5,
				Title:   "Title",
			},
		},
		{
			name:     "get html version 5",
			htmlPath: "testdata/html5.html",

			expectedDetails: models.HTMLDetails{
				Version: HTMLVersion5,
			},
		},
		{
			name:     "get html version 4.01 strict",
			htmlPath: "testdata/html401_strict.html",

			expectedDetails: models.HTMLDetails{
				Version: HTMLVersion401_STRICT,
			},
		},
		{
			name:     "get html with many headings",
			htmlPath: "testdata/headings.html",

			expectedDetails: models.HTMLDetails{
				Version: HTMLVersion5,
				HeadingsCounter: map[models.Heading]int{
					models.H1: 2,
					models.H2: 1,
					models.H3: 1,
					models.H4: 1,
					models.H5: 1,
					models.H6: 1,
				},
			},
		},
		{
			name:     "get all internal links",
			htmlPath: "testdata/internal_links.html",

			expectedDetails: models.HTMLDetails{
				Version: HTMLVersion5,
				Links: models.Links{
					"#link1": {
						URL:   "#link1",
						Count: 1,
						Type:  models.LinkTypeInternal,
					},
					"#link2": {
						URL:   "#link2",
						Count: 1,
						Type:  models.LinkTypeInternal,
					},
				},
			},
		},
		{
			name:     "get all external links",
			htmlPath: "testdata/external_links.html",
			expectedDetails: models.HTMLDetails{
				Version: HTMLVersion5,
				Links: models.Links{
					"https://google.com": {
						URL:        "https://google.com",
						Count:      1,
						Type:       models.LinkTypeExternal,
						Accessible: true,
					},
					"https://home24.de": {
						URL:        "https://home24.de",
						Count:      1,
						Type:       models.LinkTypeExternal,
						Accessible: true,
					},
				},
			},
		},
	}

	service := internal.NewHTMLService()
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			htmlFile, err := openFile(tc.htmlPath)
			if err != nil {
				suite.FailNow("failed to open html file")
			}
			suite.NotNil(htmlFile)
			doc, err := html.Parse(bytes.NewReader(htmlFile))
			suite.NoError(err)
			suite.NotNil(doc)
			details := service.GetDetails(doc)
			suite.Equal(tc.expectedDetails, details)
		})
	}
}

type setupHTTPTest struct {
	statusCode   int
	htmlFilePath string
	contentType  string
}

func (suite *serviceTestSuite) httptestSetup(setup setupHTTPTest) *httptest.Server {
	resp, err := openFile(setup.htmlFilePath)
	if err != nil {
		suite.T().Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", setup.contentType)

		w.WriteHeader(setup.statusCode)
		w.Write(resp) // nolint: errcheck

	}))
	return ts
}

func openFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
