package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tanishqv/bnb-bookings/internal/models"
)

var tests = []struct {
	tcName             string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"generals quarters", "/generals-quarters", "GET", http.StatusOK},
	{"colonels suite", "/colonels-suite", "GET", http.StatusOK},
	{"search availability", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	// {"make reservation", "/make-reservation", "GET", []postData{}, http.StatusOK},
	// {"post search availability", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "15-12-2022"},
	// 	{key: "end", value: "18-12-2022"},
	// }, http.StatusOK},
	// {"post search availability JSON", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "15-12-2022"},
	// 	{key: "end", value: "18-12-2022"},
	// }, http.StatusOK},
	// {"make reservation", "/make-reservation", "POST", []postData{
	// 	{key: "first-name", value: "John"},
	// 	{key: "last-name", value: "Doe"},
	// 	{key: "email", value: "johndoe@anonymous.com"},
	// 	{key: "phone-number", value: "0000000000"},
	// }, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close()

	for _, e := range tests {
		resp, err := testServer.Client().Get(testServer.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("%s: expected %d but got %d", e.tcName, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	respRecorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(respRecorder, req)

	if respRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", respRecorder.Code, http.StatusOK)
	}

	// Test case where reservation is not in session
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test case where room does not exist
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	respRecorder = httptest.NewRecorder()

	reservation.RoomID = 100

	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		app.ErrorLog.Println(err)
	}
	return ctx
}
