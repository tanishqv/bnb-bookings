package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

var tests = []struct {
	tcName             string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"generals quarters", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"colonels suite", "/colonels-suite", "GET", []postData{}, http.StatusOK},
	{"search availability", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"make reservation", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"post search availability", "/search-availability", "POST", []postData{
		{key: "start", value: "15-12-2022"},
		{key: "end", value: "18-12-2022"},
	}, http.StatusOK},
	{"post search availability JSON", "/search-availability-json", "POST", []postData{
		{key: "start", value: "15-12-2022"},
		{key: "end", value: "18-12-2022"},
	}, http.StatusOK},
	{"make reservation", "/make-reservation", "POST", []postData{
		{key: "first-name", value: "John"},
		{key: "last-name", value: "Doe"},
		{key: "email", value: "johndoe@anonymous.com"},
		{key: "phone-number", value: "0000000000"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close()

	for _, e := range tests {
		if e.method == "GET" {
			resp, err := testServer.Client().Get(testServer.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("%s: expected %d but got %d", e.tcName, e.expectedStatusCode, resp.StatusCode)
			}
		} else {
			vals := url.Values{}
			for _, kv := range e.params {
				vals.Add(kv.key, kv.value)
			}
			resp, err := testServer.Client().PostForm(testServer.URL+e.url, vals)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.tcName, e.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}
