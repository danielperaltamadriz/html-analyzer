package main_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/danielperaltamadriz/home24/internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func TestAcceptanceTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AcceptanceTest Suite")
}

var _ = Describe("Analyze HTML", func() {

	var server *ghttp.Server
	BeforeEach(func() {
		server = ghttp.NewServer()
		apiServer := internal.NewAPI(internal.APIConfig{})
		server.AppendHandlers(
			apiServer.HTMLHandler,
		)
	})
	AfterEach(func() {
		server.Close()
	})

	Context("Given a not found URL", func() {
		When("the HTML is requested", func() {
			var statusCode int
			BeforeEach(func() {
				ts := httptestSetup(setupHTTPTest{
					statusCode:   http.StatusNotFound,
					htmlFilePath: "./testdata/file.html",
				})
				defer ts.Close()
				resp, _ := http.Get(server.URL() + "?url=" + ts.URL)
				statusCode = resp.StatusCode
			})

			It("should return a 404 status code", func() {
				Expect(statusCode).To(Equal(http.StatusNotFound))
			})
		})
	})

	Context("Given an invalid URL", func() {
		When("the HTML is requested", func() {
			var statusCode int
			BeforeEach(func() {
				resp, err := http.Get(server.URL() + "?url=invalid_url")
				Expect(err).To(BeNil())
				statusCode = resp.StatusCode
			})

			It("should return a 400 status code", func() {
				Expect(statusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Context("Given a valid URL", func() {
		When("a simple HTML is requested", func() {
			var details *internal.DetailsResponse
			var url string
			BeforeEach(func() {
				ts := httptestSetup(setupHTTPTest{
					statusCode:   http.StatusOK,
					htmlFilePath: "./testdata/file.html",
				})
				defer ts.Close()
				url = ts.URL
				resp, _ := http.Get(server.URL() + "?url=" + ts.URL)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				json.Unmarshal(body, &details) // nolint: errcheck
			})

			It("should return HTML version 5", func() {
				Expect(details.Version.Number).To(Equal("5"))
			})

			It("should return the title", func() {
				Expect(details.Title).To(Equal("My HTML File"))
			})

			It("should return the headings", func() {
				Expect(details.Headings.H1).To(Equal(1))
				Expect(details.Headings.H2).To(Equal(2))
				Expect(details.Headings.H3).To(Equal(2))
				Expect(details.Headings.H4).To(Equal(1))
				Expect(details.Headings.H5).To(Equal(1))
				Expect(details.Headings.H6).To(Equal(1))
			})

			It("should return the internal links", func() {
				Expect(details.Links.Internal.Total).To(Equal(2))
				Expect(mapLinkDetailsToMap(details.Links.Internal.LinkDetails)).To(Equal(mapLinkDetailsToMap([]internal.LinkDetailResponse{
					{
						URL:          url + "#link1",
						Count:        1,
						IsAccessible: true,
					},
					{
						URL:          url + "#link2",
						Count:        1,
						IsAccessible: true,
					},
				})))
			})

			It("should return the external links counter", func() {
				Expect(details.Links.External.Total).To(Equal(2))
				Expect(mapLinkDetailsToMap(details.Links.External.LinkDetails)).To(Equal(mapLinkDetailsToMap([]internal.LinkDetailResponse{
					{
						URL:          "https://www.google.com",
						Count:        1,
						IsAccessible: true,
					},
					{
						URL:          "https://www.home24.de",
						Count:        1,
						IsAccessible: true,
					},
				})))
			})

			It("should return it has a login form", func() {
				Expect(details.HasLoginForm).To(Equal(true))
			})

		})
		When("a complex HTML is requested", func() {
			It("should return the title and valid status code", func() {
				ts := httptestSetup(setupHTTPTest{
					statusCode:   http.StatusOK,
					htmlFilePath: "./testdata/scrapeme.html",
				})
				defer ts.Close()

				resp, _ := http.Get(server.URL() + "?url=" + ts.URL)
				var details *internal.DetailsResponse
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				json.Unmarshal(body, &details) // nolint: errcheck

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(details.Title).To(Equal("Products – ScrapeMe"))
			})
		})
	})
})

func mapLinkDetailsToMap(links []internal.LinkDetailResponse) map[string]internal.LinkDetailResponse {
	linkDetails := make(map[string]internal.LinkDetailResponse)
	for _, link := range links {
		linkDetails[link.URL] = link
	}
	return linkDetails
}

type setupHTTPTest struct {
	statusCode   int
	htmlFilePath string
}

func httptestSetup(setup setupHTTPTest) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html, err := openFile(setup.htmlFilePath)
		if err != nil {
			http.Error(w, "server failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(setup.statusCode)
		w.Write([]byte(html)) // nolint: errcheck
	}))
	return ts
}

func openFile(filename string) (string, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(file), nil
}
