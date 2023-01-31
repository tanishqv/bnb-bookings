package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tanishqv/bnb-bookings/internal/driver"
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
	{"non existent route", "/non-existent/route", "GET", http.StatusNotFound},
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"admin dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"admin new reservations", "/admin/reservations-new", "GET", http.StatusOK},
	{"admin all reservations", "/admin/reservations-all", "GET", http.StatusOK},
	{"admin show reservation from all", "/admin/reservations/all/1/show", "GET", http.StatusOK},
	{"admin show reservation form new", "/admin/reservations/new/1/show", "GET", http.StatusOK},
	{"admin show reservation from calendar", "/admin/reservations/cal/1/show", "GET", http.StatusOK},
	{"admin show reservations calendar", "/admin/reservations-calendar", "GET", http.StatusOK},
	{"admin show reservations calendar with params", "/admin/reservations-calendar?y=2023&m=3", "GET", http.StatusOK},
}

// TestHandlers tests all GET routes
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

// TestNewRepo tests the type of object returned when initializing a new repository
func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type form NewRepo(): got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

// reservationTests is the test data for the Reservation handler
var reservationTests = []struct {
	tcName             string
	reservation        models.Reservation
	expectedStatusCode int
	expectedURL        string
	expectedHTML       string
}{
	{
		tcName: "reservation in session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		expectedStatusCode: http.StatusOK,
		expectedHTML:       `action="/make-reservation"`,
	},
	{
		tcName:             "reservation not in session",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "",
	},
	{
		tcName: "non existent room",
		reservation: models.Reservation{
			RoomID: 3,
			Room: models.Room{
				ID:       3,
				RoomName: "General's Quarters",
			},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
}

// TestRepository_Reservation tests the Reservation handler
func TestRepository_Reservation(t *testing.T) {
	for _, e := range reservationTests {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		respRecorder := httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.Reservation)
		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", e.tcName, respRecorder.Code, e.expectedStatusCode)
		}

		if e.expectedHTML != "" {
			html := respRecorder.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.tcName, e.expectedHTML)
			}
		}

		if e.expectedURL != "" {
			actualLoc, _ := respRecorder.Result().Location()
			if actualLoc.String() != e.expectedURL {
				t.Errorf("failed %s: expected location %s, but got location %s", e.tcName, e.expectedURL, actualLoc.String())
			}
		}
	}
}

// postReservationTests is the test data for the PostReservation handler
var postReservationTests = []struct {
	tcName             string
	reservation        models.Reservation
	postedData         url.Values
	expectedStatusCode int
	expectedURL        string
	expectedHTML       string
}{
	{
		tcName: "valid data",
		reservation: models.Reservation{
			RoomID: 1,
		},
		postedData: url.Values{
			"start-date":   {"2024-01-01"},
			"end-date":     {"2024-01-02"},
			"first-name":   {"John"},
			"last-name":    {"Smith"},
			"email":        {"john@smith.com"},
			"phone-number": {"123456789"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedURL:        "/reservation-summary",
	},
	{
		tcName:      "reservation not in session",
		reservation: models.Reservation{},
		postedData: url.Values{
			"start-date":   {"2050-01-01"},
			"end-date":     {"2050-01-02"},
			"first-name":   {"John"},
			"last-name":    {"Smith"},
			"email":        {"john@smith.com"},
			"phone-number": {"123456789"},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName: "missing post body",
		reservation: models.Reservation{
			RoomID: 1,
		},
		postedData:         nil,
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName: "invalid form data - first name",
		reservation: models.Reservation{
			RoomID: 1,
		},
		postedData: url.Values{
			"start-date":   {"2050-01-01"},
			"end-date":     {"2050-01-02"},
			"first-name":   {"S"},
			"last-name":    {"Smith"},
			"email":        {"john@smith.com"},
			"phone-number": {"123456789"},
		},
		expectedStatusCode: http.StatusOK,
		expectedHTML:       `action="/make-reservation"`,
	},
	{
		tcName: "non existent room",
		reservation: models.Reservation{
			RoomID: 3,
			Room: models.Room{
				ID:       3,
				RoomName: "General's Quarters",
			},
		},
		postedData: url.Values{
			"start-date":   {"2050-01-01"},
			"end-date":     {"2050-01-02"},
			"first-name":   {"John"},
			"last-name":    {"Smith"},
			"email":        {"john@smith.com"},
			"phone-number": {"123456789"},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName: "database insert failure for reservation",
		reservation: models.Reservation{
			RoomID: 2,
		},
		postedData: url.Values{
			"start-date":   {"2050-01-01"},
			"end-date":     {"2050-01-02"},
			"first-name":   {"John"},
			"last-name":    {"Smith"},
			"email":        {"john@smith.com"},
			"phone-number": {"123456789"},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName: "database insert failure for room restriction",
		reservation: models.Reservation{
			RoomID: 1000,
		},
		postedData: url.Values{
			"start-date":   {"2050-01-01"},
			"end-date":     {"2050-01-02"},
			"first-name":   {"John"},
			"last-name":    {"Smith"},
			"email":        {"john@smith.com"},
			"phone-number": {"123456789"},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
}

// TestRepository_PostReservation tests the PostReservation handler
func TestRepository_PostReservation(t *testing.T) {
	// Date: 2022-12-28 -- 01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"
	var req *http.Request
	var respRecorder *httptest.ResponseRecorder

	for _, e := range postReservationTests {
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		respRecorder = httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			var startDate, endDate time.Time

			if e.postedData != nil {
				startDate, _ = time.Parse(layout, e.postedData.Get("start-date"))
				endDate, _ = time.Parse(layout, e.postedData.Get("end-date"))
			} else {
				startDate, _ = time.Parse(layout, "2050-01-01")
				endDate, _ = time.Parse(layout, "2050-01-02")
			}
			e.reservation.StartDate = startDate
			e.reservation.EndDate = endDate

			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.PostReservation)
		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", e.tcName, respRecorder.Code, e.expectedStatusCode)
		}

		if e.expectedURL != "" {
			actualLoc, _ := respRecorder.Result().Location()
			if actualLoc.String() != e.expectedURL {
				t.Errorf("failed %s: expected location %s, but got location %s", e.tcName, e.expectedURL, actualLoc.String())
			}
		}

		if e.expectedHTML != "" {
			html := respRecorder.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.tcName, e.expectedHTML)
			}
		}
	}
}

// postAvailabilityTests is the test data for the PostAvailability handler
var postAvailabilityTests = []struct {
	tcName             string
	postedData         url.Values
	expectedStatusCode int
	expectedURL        string
}{
	{
		tcName: "rooms are available",
		postedData: url.Values{
			"start": {"2024-01-01"},
			"end":   {"2024-01-02"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		tcName:             "unable to parse form",
		postedData:         nil,
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName: "invalid start date",
		postedData: url.Values{
			"start": {"invalid"},
			"end":   {"2050-01-02"},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName: "invalid end date",
		postedData: url.Values{
			"start": {"2050-01-01"},
			"end":   {"invalid"},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName: "rooms are not available",
		postedData: url.Values{
			"start": {"2040-01-01"},
			"end":   {"2040-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		tcName: "database query failure",
		postedData: url.Values{
			"start": {"2060-01-01"},
			"end":   {"2060-01-02"},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
}

// TestRepository_PostAvailability tests the PostAvailability handler
func TestRepository_PostAvailability(t *testing.T) {
	var req *http.Request
	for _, e := range postAvailabilityTests {
		if e.postedData == nil {
			req, _ = http.NewRequest("POST", "/search-availability", nil)
		} else {
			req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(e.postedData.Encode()))
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		respRecorder := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.PostAvailability)
		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", e.tcName, respRecorder.Code, e.expectedStatusCode)
		}

		if e.expectedURL != "" {
			actualLoc, _ := respRecorder.Result().Location()
			if actualLoc.String() != e.expectedURL {
				t.Errorf("failed %s: expected location %s, but got location %s", e.tcName, e.expectedURL, actualLoc.String())
			}
		}
	}
}

// availabilityJSONTests is the test data for the AvailabilityJSON handler
var availabilityJSONTests = []struct {
	tcName          string
	postedData      url.Values
	expectedOK      bool
	expectedMessage string
}{
	{
		tcName: "room is available",
		postedData: url.Values{
			"start":   {"2024-01-01"},
			"end":     {"2024-01-02"},
			"room-id": {"1"},
		},
		expectedOK: true,
	},
	{
		tcName:          "unable to parse form",
		postedData:      nil,
		expectedOK:      false,
		expectedMessage: "Internal server error, unable to parse form",
	},
	{
		tcName: "invalid start date",
		postedData: url.Values{
			"start":   {"invalid"},
			"end":     {"2024-01-02"},
			"room-id": {"1"},
		},
		expectedOK:      false,
		expectedMessage: "Internal server error, unable to parse start date",
	},
	{
		tcName: "invalid end date",
		postedData: url.Values{
			"start":   {"2024-01-01"},
			"end":     {"invalid"},
			"room-id": {"1"},
		},
		expectedOK:      false,
		expectedMessage: "Internal server error, unable to parse end date",
	},
	{
		tcName: "invalid room id",
		postedData: url.Values{
			"start":   {"2024-01-01"},
			"end":     {"2024-01-02"},
			"room-id": {"invalid"},
		},
		expectedOK:      false,
		expectedMessage: "Internal server error, unable to get/convert room id",
	},
	{
		tcName: "room is not available",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"2040-01-02"},
			"room-id": {"1"},
		},
		expectedOK: false,
	},
	{
		tcName: "database query failure",
		postedData: url.Values{
			"start":   {"2060-01-01"},
			"end":     {"2060-01-02"},
			"room-id": {"1"},
		},
		expectedOK:      false,
		expectedMessage: "Internal server error, error getting available room by date",
	},
}

// TestRepository_AvailabilityJSON tests the AvailabilityJSON handler
func TestRepository_AvailabilityJSON(t *testing.T) {
	var req *http.Request
	for _, e := range availabilityJSONTests {
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/search-availability-json", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		respRecorder := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(respRecorder, req)

		var j jsonResponse
		err := json.Unmarshal(respRecorder.Body.Bytes(), &j)
		if err != nil {
			t.Error("Failed to parse JSON")
		}

		if j.OK != e.expectedOK {
			t.Errorf("%s: expected %v but got %v", e.tcName, e.expectedOK, j.OK)
		}

		if j.Message != e.expectedMessage {
			t.Errorf("%s: expected message \"%v\" but got \"%v\"", e.tcName, e.expectedMessage, j.Message)
		}
	}
}

// reservationSummaryTests is the test data for the ReservationSummary handler
var reservationSummaryTests = []struct {
	tcName             string
	reservation        models.Reservation
	expectedStatusCode int
	expectedURL        string
}{
	{
		tcName: "reservation in session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		tcName:             "reservation not in session",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
}

// TestRepository_ReservationSummary tests the ReservationSummary handler
func TestRepository_ReservationSummary(t *testing.T) {
	for _, e := range reservationSummaryTests {
		req, _ := http.NewRequest("GET", "/reservation-summary", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		respRecorder := httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.ReservationSummary)
		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", e.tcName, respRecorder.Code, e.expectedStatusCode)
		}

		if e.expectedURL != "" {
			actualLoc, _ := respRecorder.Result().Location()
			if actualLoc.String() != e.expectedURL {
				t.Errorf("failed %s: expected location %s, but got location %s", e.tcName, e.expectedURL, actualLoc.String())
			}
		}
	}
}

// chooseRoomTests is the test data for the ChooseRoom handler
var chooseRoomTests = []struct {
	tcName             string
	reservation        models.Reservation
	url                string
	expectedStatusCode int
	expectedURL        string
}{
	{
		tcName: "reservation in session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		url:                "/choose-room/1",
		expectedStatusCode: http.StatusSeeOther,
		expectedURL:        "/make-reservation",
	},
	{
		tcName:             "reservation not in session",
		reservation:        models.Reservation{},
		url:                "/choose-room/1",
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName:             "malformed or missing parameter in URL",
		reservation:        models.Reservation{},
		url:                "/choose-room/invalid",
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
}

// TestRepository_ChooseRoom tests the ChooseRoom handler
func TestRepository_ChooseRoom(t *testing.T) {
	for _, e := range chooseRoomTests {
		req, _ := http.NewRequest("GET", e.url, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.RequestURI = e.url

		respRecorder := httptest.NewRecorder()
		if e.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.ChooseRoom)
		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", e.tcName, respRecorder.Code, e.expectedStatusCode)
		}

		if e.expectedURL != "" {
			actualLoc, _ := respRecorder.Result().Location()
			if actualLoc.String() != e.expectedURL {
				t.Errorf("failed %s: expected location %s, but got location %s", e.tcName, e.expectedURL, actualLoc.String())
			}
		}
	}
}

// bookRoomTests is the test data for the BookRoom handler
var bookRoomTests = []struct {
	tcName             string
	url                string
	expectedStatusCode int
	expectedURL        string
}{
	{
		tcName:             "room exists",
		url:                "/book-room?s=2024-01-01&e=2024-01-02&id=1",
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		tcName:             "room does not exist",
		url:                "/book-room?s=2024-01-01&e=2024-01-02&id=3",
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName:             "room does not exist - malformed URL",
		url:                "/book-room?s=2024-01-01&e=2024-01-02&id=invalid",
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName:             "invalid start date",
		url:                "/book-room?s=invalid&e=2024-01-02&id=1",
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
	{
		tcName:             "invalid end date",
		url:                "/book-room?s=2024-01-01&e=invalid&id=1",
		expectedStatusCode: http.StatusTemporaryRedirect,
		expectedURL:        "/",
	},
}

// TestRepository_BookRoom tests the BookRoom handler
func TestRepository_BookRoom(t *testing.T) {
	for _, e := range bookRoomTests {
		req, _ := http.NewRequest("GET", e.url, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.RequestURI = e.url

		respRecorder := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.BookRoom)
		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", e.tcName, respRecorder.Code, e.expectedStatusCode)
		}

		if e.expectedURL != "" {
			actualLoc, _ := respRecorder.Result().Location()
			if actualLoc.String() != e.expectedURL {
				t.Errorf("failed %s: expected location %s, but got location %s", e.tcName, e.expectedURL, actualLoc.String())
			}
		}
	}
}

// postShowLoginTests is the test data for the PostShowLogin handler
var postShowLoginTests = []struct {
	tcName             string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedURL        string
}{
	{
		"missing post body",
		"",
		http.StatusTemporaryRedirect,
		"",
		"/",
	},
	{
		"valid credentials",
		"admin@fsbnb.com",
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"invalid credentials",
		"someone@somewhere.co",
		http.StatusSeeOther,
		"",
		"/user/login",
	},
	{
		"invalid data",
		"invalid-email",
		http.StatusOK,
		`action="/user/login"`,
		"",
	},
}

// TestRepository_PostShowLogin tests the PostShowLogin handler
func TestRepository_PostShowLogin(t *testing.T) {
	var req *http.Request
	for _, e := range postShowLoginTests {
		postedData := url.Values{}
		if e.email == "" {
			req, _ = http.NewRequest("POST", "/user/login", nil)
		} else {
			postedData.Add("email", e.email)
			postedData.Add("passwd", "admin")

			req, _ = http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handler := http.HandlerFunc(Repo.PostShowLogin)
		respRecorder := httptest.NewRecorder()

		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.tcName, e.expectedStatusCode, respRecorder.Code)
		}

		if e.expectedURL != "" {
			actualLoc, _ := respRecorder.Result().Location()
			if actualLoc.String() != e.expectedURL {
				t.Errorf("failed %s: expected location %s, but got location %s", e.tcName, e.expectedURL, actualLoc.String())
			}
		}

		if e.expectedHTML != "" {
			html := respRecorder.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s, expected to find %s but did not", e.tcName, e.expectedHTML)
			}
		}
	}
}

// adminPostShowReservationTests is the test data for the AdminPostShowReservation handler
var adminPostShowReservationTests = []struct {
	tcName             string
	url                string
	postedData         url.Values
	expectedStatusCode int
	expectedURL        string
}{
	{
		tcName: "valid reservation from new",
		url:    "/admin/reservations/new/1/show",
		postedData: url.Values{
			"first-name":   {"John"},
			"last-name":    {"Smith"},
			"email":        {"john@smith.com"},
			"phone-number": {"1234567890"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedURL:        "/admin/reservations-new",
	},
	{
		tcName: "valid reservation from all",
		url:    "/admin/reservations/all/1/show",
		postedData: url.Values{
			"first-name":   {"John"},
			"last-name":    {"Smith"},
			"email":        {"john@smith.com"},
			"phone-number": {"1234567890"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedURL:        "/admin/reservations-all",
	},
	{
		tcName: "valid reservation from calendar",
		url:    "/admin/reservations/cal/1/show",
		postedData: url.Values{
			"first-name":   {"John"},
			"last-name":    {"Smith"},
			"email":        {"john@smith.com"},
			"phone-number": {"1234567890"},
			"year":         {"2023"},
			"month":        {"02"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedURL:        "/admin/reservations-calendar?y=2023&m=02",
	},
}

// TestRepository_AdminPostShowReservation tests the AdminPostShowReservation handler
func TestRepository_AdminPostShowReservation(t *testing.T) {
	var req *http.Request
	for _, e := range adminPostShowReservationTests {
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", e.url, strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", e.url, nil)
		}

		req.RequestURI = e.url

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handler := http.HandlerFunc(Repo.AdminPostShowReservation)
		respRecorder := httptest.NewRecorder()

		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.tcName, e.expectedStatusCode, respRecorder.Code)
		}

		if e.expectedURL != "" {
			actualLoc, _ := respRecorder.Result().Location()
			if actualLoc.String() != e.expectedURL {
				t.Errorf("failed %s: expected location %s, but got location %s", e.tcName, e.expectedURL, actualLoc.String())
			}
		}
	}
}

// adminProcessReservationTests is the test data for the AdminProcessReservation handler
var adminProcessReservationTests = []struct {
	tcName      string
	from        string
	queryParams string
}{
	{
		tcName: "process reservation back to new",
		from:   "new",
	},
	{
		tcName: "process reservation back to all",
		from:   "all",
	},
	{
		tcName:      "process reservation back to calendar",
		from:        "cal",
		queryParams: "y=2023&m=02",
	},
}

// TestRepository_AdminProcessReservation tests the AdminProcessReservation handler
func TestRepository_AdminProcessReservation(t *testing.T) {
	for _, e := range adminProcessReservationTests {
		url := fmt.Sprintf("/admin/process-reservation/%s/1/process?%s", e.from, e.queryParams)

		req, _ := http.NewRequest("GET", url, nil)
		req.RequestURI = url

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		respRecorder := httptest.NewRecorder()

		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != http.StatusSeeOther {
			t.Errorf("failed %s: expected code %d, but got %d", e.tcName, http.StatusSeeOther, respRecorder.Code)
		}
	}
}

// adminDeleteReservationTests is the test data for the AdminDeleteReservation handler
var adminDeleteReservationTests = []struct {
	tcName             string
	from               string
	roomID             int
	queryParams        string
	expectedStatusCode int
}{
	{
		tcName:             "delete reservation back to new",
		from:               "new",
		roomID:             1,
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		tcName:             "delete reservation back to all",
		from:               "all",
		roomID:             1,
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		tcName:             "delete reservation back to calendar",
		from:               "cal",
		roomID:             1,
		queryParams:        "y=2023&m=02",
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		tcName:             "delete reservation error from new",
		from:               "new",
		roomID:             1000,
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
}

// TestRepository_AdminDeleteReservation tests the AdminDeleteReservation handler
func TestRepository_AdminDeleteReservation(t *testing.T) {
	for _, e := range adminDeleteReservationTests {
		url := fmt.Sprintf("/admin/delete-reservation/%s/%d/delete?%s", e.from, e.roomID, e.queryParams)

		req, _ := http.NewRequest("GET", url, nil)
		req.RequestURI = url

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		respRecorder := httptest.NewRecorder()

		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.tcName, e.expectedStatusCode, respRecorder.Code)
		}
	}
}

// adminPostReservationsCalendarTests is the test data for the AdminPostReservationsCalendar handler
var adminPostReservationsCalendarTests = []struct {
	tcName             string
	postedData         url.Values
	roomID             int
	blocks             int
	reservations       int
	expectedStatusCode int
}{
	{
		tcName: "Adding block",
		postedData: url.Values{
			"y": {time.Now().Format("2006")},
			"m": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		roomID:             1,
		blocks:             1,
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		tcName: "Error while adding block",
		postedData: url.Values{
			"y": {time.Now().Format("2006")},
			"m": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1000_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		roomID:             1,
		blocks:             1,
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		tcName: "Removing block",
		postedData: url.Values{
			"y": {time.Now().Format("2006")},
			"m": {time.Now().Format("01")},
			fmt.Sprintf("remove_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		roomID:             1,
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		tcName: "Adding reservation",
		postedData: url.Values{
			"y": {time.Now().Format("2006")},
			"m": {time.Now().Format("01")},
		},
		roomID:             1,
		reservations:       1,
		expectedStatusCode: http.StatusSeeOther,
	},
}

// TestRepository_AdminPostReservationsCalendar tests the AdminPostReservationsCalendar handler
func TestRepository_AdminPostReservationsCalendar(t *testing.T) {
	var req *http.Request
	for _, e := range adminPostReservationsCalendarTests {
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		blockMap := make(map[string]int)
		reservationMap := make(map[string]int)

		now := time.Now()
		currYear, currMonth, _ := now.Date()
		currLoc := now.Location()

		firstOfMonth := time.Date(currYear, currMonth, 1, 0, 0, 0, 0, currLoc)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		if e.blocks > 0 {
			blockMap[firstOfMonth.Format("2006-01-2")] = e.blocks
		}

		if e.reservations > 0 {
			reservationMap[lastOfMonth.Format("2006-01-2")] = e.reservations
		}

		session.Put(ctx, fmt.Sprintf("block_map_%d", e.roomID), blockMap)
		session.Put(ctx, fmt.Sprintf("reservation_map_%d", e.roomID), reservationMap)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		respRecorder := httptest.NewRecorder()

		handler.ServeHTTP(respRecorder, req)

		if respRecorder.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.tcName, e.expectedStatusCode, respRecorder.Code)
		}
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		app.ErrorLog.Println(err)
	}
	return ctx
}
