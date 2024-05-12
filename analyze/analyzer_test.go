package analyze_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/danielperaltamadriz/home24/analyze"
	"github.com/danielperaltamadriz/home24/analyze/models"
	"github.com/stretchr/testify/suite"
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

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			analyzer := analyze.NewAnalyzer()
			result, err := analyzer.RunFromURL(tc.inputURL)
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

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			fakeServer := suite.httptestSetup(tc.fakeServerSetup)
			defer fakeServer.Close()
			analyzer := analyze.NewAnalyzer()
			result, err := analyzer.RunFromURL(fakeServer.URL)
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
		{
			name: "Get HTML from not found URL",
			fakeServerSetup: &setupHTTPTest{
				statusCode:   http.StatusNotFound,
				htmlFilePath: "testdata/html.html",
				contentType:  "text/html",
			},
			errMessageString: "invalid status code",
		},
		{
			name:             "Get HTML from invalid URL",
			fakeServerSetup:  nil,
			errMessageString: "invalid url",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			fakeServer := suite.httptestSetup(tc.fakeServerSetup)
			url := "/url"
			if tc.fakeServerSetup != nil {
				defer fakeServer.Close()
				url = fakeServer.URL
			}
			analyzer := analyze.NewAnalyzer()
			result, err := analyzer.RunFromURL(url)
			suite.Nil(result)
			suite.ErrorContains(err, tc.errMessageString)
		})
	}
}

type getDetailsTestCase struct {
	name     string
	htmlPath string

	expectedDetails models.HTMLDetails
}

func (suite *serviceTestSuite) TestGetTitle() {
	testCases := []getDetailsTestCase{
		{
			name:     " get title from html",
			htmlPath: "testdata/title.html",

			expectedDetails: models.HTMLDetails{
				Title: "Title",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			analyzer := analyze.NewAnalyzer()
			analyzer.WithSearchSingleElements(analyzer.Title)
			suite.setupTestGetDetails(tc, analyzer)
		})
	}
}

func (suite *serviceTestSuite) TestGetHTMLVersion() {
	testCases := []getDetailsTestCase{
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
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			analyzer := analyze.NewAnalyzer()
			analyzer.WithSearchSingleElements(analyzer.HTMLVersion)
			suite.setupTestGetDetails(tc, analyzer)
		})
	}

}

func (suite *serviceTestSuite) TestGetHeadings() {
	testCases := []getDetailsTestCase{
		{
			name:     "get html with many headings",
			htmlPath: "testdata/headings.html",

			expectedDetails: models.HTMLDetails{
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
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			analyzer := analyze.NewAnalyzer()
			analyzer.WithSearchManyElements(analyzer.Headings)
			suite.setupTestGetDetails(tc, analyzer)
		})
	}
}

func (suite *serviceTestSuite) TestGetLinks() {
	testCases := []getDetailsTestCase{
		{
			name:     "get all internal links",
			htmlPath: "testdata/internal_links.html",

			expectedDetails: models.HTMLDetails{
				Links: models.Links{
					"#link1": {
						URL:        "#link1",
						Count:      1,
						Type:       models.LinkTypeInternal,
						Accessible: true,
					},
					"/link2": {
						URL:        "/link2",
						Count:      1,
						Type:       models.LinkTypeInternal,
						Accessible: true,
					},
					"/link3": {
						URL:        "/link3",
						Count:      1,
						Type:       models.LinkTypeInternal,
						Accessible: false,
					},
				},
			},
		},
		{
			name:     "get all external links",
			htmlPath: "testdata/external_links.html",
			expectedDetails: models.HTMLDetails{
				Links: models.Links{
					"https://google.com": {
						URL:        "https://google.com",
						Count:      1,
						Type:       models.LinkTypeExternal,
						Accessible: false,
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

	fakeLinkVerifierFunc := func(l *models.Link) bool {
		if strings.Contains(l.URL, "google") {
			return false
		}
		if strings.Contains(l.URL, "/link3") {
			return false
		}
		return true
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			analyzer := analyze.NewAnalyzer()
			analyzer.WithSearchManyElements(analyzer.Links)
			analyzer.WithLinkVerifierFunc(fakeLinkVerifierFunc)
			fakeServer := suite.httptestSetup(&setupHTTPTest{
				statusCode:   http.StatusOK,
				htmlFilePath: tc.htmlPath,
				contentType:  "text/html",
			})
			defer fakeServer.Close()
			details, err := analyzer.RunFromURL(fakeServer.URL + "/path")
			suite.NoError(err)
			if _, ok := tc.expectedDetails.Links["#link1"]; ok {
				tc.expectedDetails.Links[fakeServer.URL+"/path#link1"] = &models.Link{
					URL:        fakeServer.URL + "/path#link1",
					Count:      tc.expectedDetails.Links["#link1"].Count,
					Type:       models.LinkTypeInternal,
					Accessible: true,
				}
				delete(tc.expectedDetails.Links, "#link1")
			}
			if _, ok := tc.expectedDetails.Links["/link2"]; ok {
				tc.expectedDetails.Links[fakeServer.URL+"/link2"] = &models.Link{
					URL:        fakeServer.URL + "/link2",
					Count:      tc.expectedDetails.Links["/link2"].Count,
					Type:       models.LinkTypeInternal,
					Accessible: true,
				}
				delete(tc.expectedDetails.Links, "/link2")
			}
			if _, ok := tc.expectedDetails.Links["/link3"]; ok {
				tc.expectedDetails.Links[fakeServer.URL+"/path/link3"] = &models.Link{
					URL:        fakeServer.URL + "/path/link3",
					Count:      tc.expectedDetails.Links["/link3"].Count,
					Type:       models.LinkTypeInternal,
					Accessible: false,
				}
				delete(tc.expectedDetails.Links, "/link3")
			}
			suite.Equal(&tc.expectedDetails, details)
		})
	}
}

func (suite *serviceTestSuite) TestGetLoginForm() {
	testCases := []getDetailsTestCase{
		{
			name:     "get html with login form",
			htmlPath: "testdata/login.html",

			expectedDetails: models.HTMLDetails{
				HasLoginForm: true,
			},
		},
		{
			name:     "get html with login form",
			htmlPath: "testdata/real_login.html",

			expectedDetails: models.HTMLDetails{
				HasLoginForm: true,
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			analyzer := analyze.NewAnalyzer()
			analyzer.WithSearchSingleElements(analyzer.HasLoginForm)
			suite.setupTestGetDetails(tc, analyzer)
		})
	}
}

func (suite *serviceTestSuite) setupTestGetDetails(tc getDetailsTestCase, analyzer *analyze.Analyzer) {
	fakeServer := suite.httptestSetup(&setupHTTPTest{
		statusCode:   http.StatusOK,
		htmlFilePath: tc.htmlPath,
		contentType:  "text/html",
	})
	defer fakeServer.Close()
	details, err := analyzer.RunFromURL(fakeServer.URL)
	suite.NoError(err)
	suite.Equal(&tc.expectedDetails, details)
}

type setupHTTPTest struct {
	statusCode   int
	htmlFilePath string
	contentType  string
}

func (suite *serviceTestSuite) httptestSetup(setup *setupHTTPTest) *httptest.Server {
	if setup == nil {
		return nil
	}
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
