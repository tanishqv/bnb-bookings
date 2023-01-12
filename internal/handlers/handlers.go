package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tanishqv/bnb-bookings/internal/config"
	"github.com/tanishqv/bnb-bookings/internal/driver"
	"github.com/tanishqv/bnb-bookings/internal/forms"
	"github.com/tanishqv/bnb-bookings/internal/models"
	"github.com/tanishqv/bnb-bookings/internal/render"
	"github.com/tanishqv/bnb-bookings/internal/repository"
	"github.com/tanishqv/bnb-bookings/internal/repository/dbrepo"
)

// Handlers may not use template cache, but the config may be updated with things that makes the application run better
// Repository pattern to swap components out of the application with minimal changes erquired to code base

// Repo is the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewTestRepo creates a new repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Colonels renders the room page
func (m *Repository) Colonels(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "colonels.page.tmpl", &models.TemplateData{})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start-date"] = sd
	stringMap["end-date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.FirstName = r.Form.Get("first-name")
	reservation.LastName = r.Form.Get("last-name")
	reservation.Phone = r.Form.Get("phone-number")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)

	form.Required("first-name", "last-name", "email", "phone-number")
	form.MinLength("first-name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "Invalid form", http.StatusSeeOther)
		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot insert reservation into the database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1, // Type: Reservation
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot insert room restriction")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Send email notification to guest
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong>
		<hr>
		Dear %s, <br>
		This is to inform you about your reservation at Fort Smythe BnB. Please find the necessary details mentioned below: <br>
		<div style="text-align:center !important;">
			<strong>Room</strong>: %s <br>
			<strong>Duration</strong>: %s to %s <br>
		</div>
	`,
		reservation.FirstName+reservation.LastName,
		reservation.Room.RoomName,
		reservation.StartDate.Format("2006-01-02"),
		reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:      reservation.Email,
		From:    "manager@fsbnb.com",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg

	// Send email notification to property owner
	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Notification</strong>
		<hr>
		Mr/Ms Property Owner <br>
		A reservation has been made at Fort Smythe BnB. Please find the necessary details mentioned below: <br>
		<div style="text-align:center !important;">
			<strong>Room</strong>: %s <br>
			<strong>Duration</strong>: %s to %s <br>
		</div>
	`,
		reservation.Room.RoomName,
		reservation.StartDate.Format("2006-01-02"),
		reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To:      "property-owner@fsbnb.com",
		From:    "manager@fsbnb.com",
		Subject: "Reservation Notification",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Availability renders the search availability form
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability renders the search availability form
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	start := r.Form.Get("start")
	end := r.Form.Get("end")

	// Date Format: 01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"

	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.ErrorLog.Println("Cannot parse start date")
		m.App.Session.Put(r.Context(), "error", "Can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.ErrorLog.Println("Cannot parse end date")
		m.App.Session.Put(r.Context(), "error", "Can't parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRoomsByDates(startDate, endDate)
	if err != nil {
		m.App.ErrorLog.Println("Can't get availability for rooms")
		m.App.Session.Put(r.Context(), "error", "Can't get availability for rooms")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// No availability
	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability!")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.RenderTemplate(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles req for availability and send JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error, unable to parse form",
		}

		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	// Date Format: 01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error, unable to parse start date",
		}

		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error, unable to parse end date",
		}

		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room-id"))
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error, unable to get/convert room id",
		}

		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error, error getting available room(s) by date",
		}

		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary shows the reservation summary
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Cannot get item from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start-date"] = sd
	stringMap["end-date"] = ed

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// ChooseRoom displays list of available rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.ErrorLog.Println("missing/malformed URL parameter \"id\"")
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("cannot get item from session")
		m.App.Session.Put(r.Context(), "error", "Cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID
	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes URL parameters and builds a sessional variable, and takes user to make reservation page
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot get room from database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	s := r.URL.Query().Get("s")
	e := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, s)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, e)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName
	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
