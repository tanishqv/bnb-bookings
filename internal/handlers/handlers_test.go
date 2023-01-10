package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type form NewRepo(): got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
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

	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)
	respRecorder := httptest.NewRecorder()

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

func TestRepository_PostReservation(t *testing.T) {
	// Date: 2022-12-28 -- 01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"

	sd := "2050-01-01"
	ed := "2050-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	reqBody := fmt.Sprintf("%s=%s", "start-date", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end-date", ed)
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first-name=John")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone-number=123456789")

	reservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    1,
	}

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.PostReservation)
	respRecorder := httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, wanted %d", respRecorder.Code, http.StatusSeeOther)
	}

	// Test for missing reservation in session
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostReservation)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for missing post body
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostReservation)
	respRecorder = httptest.NewRecorder()

	reservation = models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    1,
	}
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid form data
	reqBody = fmt.Sprintf("%s=%s", "start-date", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end-date", ed)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "first-name", "S")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone-number=123456789")
	// reqBody = fmt.Sprintf("%s&%s=%d", reqBody, "room-id", 1)

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostReservation)
	respRecorder = httptest.NewRecorder()

	reservation = models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    1,
	}
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code for invalid data: got %d, wanted %d", respRecorder.Code, http.StatusSeeOther)
	}

	// Test for failure to insert reservation into database
	reqBody = fmt.Sprintf("%s=%s", "start-date", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end-date", ed)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "first-name", "John")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone-number=123456789")
	// reqBody = fmt.Sprintf("%s&%s=%d", reqBody, "room-id", 2)

	reservation = models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    2,
	}

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostReservation)
	respRecorder = httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed when trying to insert reservation: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for failure to insert restriction into database
	reqBody = fmt.Sprintf("%s=%s", "start-date", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end-date", ed)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "first-name", "John")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone-number=123456789")

	reservation = models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    1000,
	}

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostReservation)
	respRecorder = httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed when trying to insert reservation: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostAvailability(t *testing.T) {
	sd := "2024-01-01"
	ed := "2024-01-02"

	reqBody := fmt.Sprintf("%s=%s", "start", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", ed)

	req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(Repo.PostAvailability)
	respRecorder := httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code when rooms are available: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// No request body: unable to parse form
	req, _ = http.NewRequest("POST", "/search-availability", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostAvailability)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Invalid start date
	ed = "2024-01-02"

	reqBody = fmt.Sprintf("%s=%s", "start", "invalid")
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", ed)

	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostAvailability)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code for invalid start date: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Invalid end date
	sd = "2024-01-01"

	reqBody = fmt.Sprintf("%s=%s", "start", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", "invalid")

	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostAvailability)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code for invalid end date: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Rooms are not available: Threshold date for room availability - 2039-12-31
	sd = "2040-01-01"
	ed = "2040-01-02"
	reqBody = fmt.Sprintf("%s=%s", "start", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", ed)

	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostAvailability)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code when rooms are not available: got %d, wanted %d", respRecorder.Code, http.StatusSeeOther)
	}

	// Rooms are not available: Threshold date for querying room availability - 2060-01-01
	sd = "2060-01-01"
	ed = "2060-01-02"
	reqBody = fmt.Sprintf("%s=%s", "start", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", ed)

	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.PostAvailability)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code when database query fails: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	sd := "2024-01-01"
	ed := "2024-01-02"

	reqBody := fmt.Sprintf("%s=%s", "start", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", ed)
	reqBody = fmt.Sprintf("%s&%s=%d", reqBody, "room-id", 1)

	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(Repo.AvailabilityJSON)
	respRecorder := httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)

	var jsonResp jsonResponse
	err := json.Unmarshal(respRecorder.Body.Bytes(), &jsonResp)
	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if !jsonResp.OK {
		t.Error("Got no availability when some was expected in AvailabilityJSON")
	}

	// No request body: unable to parse form
	req, _ = http.NewRequest("POST", "/search-availability-json", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)

	err = json.Unmarshal(respRecorder.Body.Bytes(), &jsonResp)
	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if jsonResp.OK || jsonResp.Message != "Internal server error, unable to parse form" {
		t.Error("Got availability when request body was empty")
	}

	// Invalid start date
	reqBody = fmt.Sprintf("%s=%s", "start", "invalid")
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", ed)
	reqBody = fmt.Sprintf("%s&%s=%d", reqBody, "room-id", 1)

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)

	err = json.Unmarshal(respRecorder.Body.Bytes(), &jsonResp)
	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if jsonResp.OK || jsonResp.Message != "Internal server error, unable to parse start date" {
		t.Error("Got availability when start date was invalid")
	}

	// Invalid end date
	reqBody = fmt.Sprintf("%s=%s", "start", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", "invalid")
	reqBody = fmt.Sprintf("%s&%s=%d", reqBody, "room-id", 1)

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)

	err = json.Unmarshal(respRecorder.Body.Bytes(), &jsonResp)
	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if jsonResp.OK || jsonResp.Message != "Internal server error, unable to parse end date" {
		t.Error("Got availability when end date was invalid")
	}

	// Invalid room id
	reqBody = fmt.Sprintf("%s=%s", "start", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", ed)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "room-id", "invalid")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)

	err = json.Unmarshal(respRecorder.Body.Bytes(), &jsonResp)
	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if jsonResp.OK || jsonResp.Message != "Internal server error, unable to get/convert room id" {
		t.Error("Got availability when room id was invalid")
	}

	// Rooms are not available: Threshold date for room availability - 2039-12-31
	sd = "2040-01-01"
	ed = "2040-01-02"
	reqBody = fmt.Sprintf("%s=%s", "start", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", ed)
	reqBody = fmt.Sprintf("%s&%s=%d", reqBody, "room-id", 1)

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)

	err = json.Unmarshal(respRecorder.Body.Bytes(), &jsonResp)
	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if jsonResp.OK {
		t.Error("Got availability when none was expected")
	}

	// Rooms are not available: Threshold date for querying room availability - 2060-01-01
	sd = "2060-01-01"
	ed = "2060-01-02"
	reqBody = fmt.Sprintf("%s=%s", "start", sd)
	reqBody = fmt.Sprintf("%s&%s=%s", reqBody, "end", ed)
	reqBody = fmt.Sprintf("%s&%s=%d", reqBody, "room-id", 1)

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)

	err = json.Unmarshal(respRecorder.Body.Bytes(), &jsonResp)
	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if jsonResp.OK || jsonResp.Message != "Internal server error, error getting available room(s) by date" {
		t.Error("Got availability when simulating database error")
	}
}

func TestRepository_ReservationSummary(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/reservation-summary", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ReservationSummary)
	respRecorder := httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusOK {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", respRecorder.Code, http.StatusOK)
	}

	// Reservation not in session
	req, _ = http.NewRequest("GET", "/reservation-summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	handler = http.HandlerFunc(Repo.ReservationSummary)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}
	uri := "/choose-room/1"

	req, _ := http.NewRequest("GET", uri, nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.RequestURI = uri
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ChooseRoom)
	respRecorder := httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code for valid choice of room: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Reservation not in session
	uri = "/choose-room/1"

	req, _ = http.NewRequest("GET", uri, nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.RequestURI = uri

	handler = http.HandlerFunc(Repo.ChooseRoom)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseRoom handler returned wrong response code when reservation is not in session: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Malformed/Missing URL parameter
	uri = "/choose-room/invalid"
	req, _ = http.NewRequest("GET", uri, nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.RequestURI = uri

	handler = http.HandlerFunc(Repo.ChooseRoom)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseRoom handler returned wrong response code for missing/malform URL parameter: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_BookRoom(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	sd := "2024-01-01"
	ed := "2024-01-02"
	uri := fmt.Sprintf("/book-room?%s=%s", "s", sd)
	uri = fmt.Sprintf("%s&%s=%s", uri, "e", ed)
	uri = fmt.Sprintf("%s&%s=%d", uri, "id", 1)

	req, _ := http.NewRequest("GET", uri, nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.BookRoom)
	respRecorder := httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", respRecorder.Code, http.StatusSeeOther)
	}

	// Room does not exist
	sd = "2024-01-01"
	ed = "2024-01-02"
	uri = fmt.Sprintf("/book-room?%s=%s", "s", sd)
	uri = fmt.Sprintf("%s&%s=%s", uri, "e", ed)
	uri = fmt.Sprintf("%s&%s=%d", uri, "id", 1000)

	req = httptest.NewRequest("GET", uri, nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	handler = http.HandlerFunc(Repo.BookRoom)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("BookRoom handler returned wrong response code when room does not exist: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Room does not exist: Malformed URL
	sd = "2024-01-01"
	ed = "2024-01-02"
	uri = fmt.Sprintf("/book-room?%s=%s", "s", sd)
	uri = fmt.Sprintf("%s&%s=%s", uri, "e", ed)
	uri = fmt.Sprintf("%s&%s=%s", uri, "id", "invalid")

	req = httptest.NewRequest("GET", uri, nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	handler = http.HandlerFunc(Repo.BookRoom)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("BookRoom handler returned wrong response code when room id in URL is malformed: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Invalid start date
	ed = "2024-01-02"
	uri = fmt.Sprintf("/book-room?%s=%s", "s", "invalid")
	uri = fmt.Sprintf("%s&%s=%s", uri, "e", ed)
	uri = fmt.Sprintf("%s&%s=%d", uri, "id", 1)

	req = httptest.NewRequest("GET", uri, nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	handler = http.HandlerFunc(Repo.BookRoom)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("BookRoom handler returned wrong response code when start date is invalid: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Invalid end date
	sd = "2024-01-01"
	uri = fmt.Sprintf("/book-room?%s=%s", "s", sd)
	uri = fmt.Sprintf("%s&%s=%s", uri, "e", "invalid")
	uri = fmt.Sprintf("%s&%s=%d", uri, "id", 1)

	req = httptest.NewRequest("GET", uri, nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	handler = http.HandlerFunc(Repo.BookRoom)
	respRecorder = httptest.NewRecorder()

	handler.ServeHTTP(respRecorder, req)
	if respRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("BookRoom handler returned wrong response code when end date is invalid: got %d, wanted %d", respRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		app.ErrorLog.Println(err)
	}
	return ctx
}
