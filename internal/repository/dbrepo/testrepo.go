package dbrepo

import (
	"errors"
	"log"
	"time"

	"github.com/tanishqv/bnb-bookings/internal/models"
)

func (tr *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (tr *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	if res.RoomID == 2 {
		return 0, errors.New("insert reservation failed")
	}
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (tr *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	if r.RoomID == 1000 {
		return errors.New("insert restriction failed")
	}
	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for the roomID, and false if availability doesn't exist
func (tr *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	layout := "2006-01-02"
	naDate := "2039-12-31"
	failDate := "2060-01-01"

	noAvailabiltyDate, err := time.Parse(layout, naDate)
	if err != nil {
		log.Println(err)
	}

	dateToFail, err := time.Parse(layout, failDate)
	if err != nil {
		log.Println(err)
	}

	// Fail the query
	if start == dateToFail {
		return false, errors.New("query processing failed")
	}

	// No availability
	if start.After(noAvailabiltyDate) {
		return false, nil
	}

	return true, nil
}

// SearchAvailabilityForAllRoomsByDates returns a slice of available rooms, if any, for any given date range
func (tr *testDBRepo) SearchAvailabilityForAllRoomsByDates(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	layout := "2006-01-02"
	naDate := "2039-12-31"
	noAvailabiltyDate, err := time.Parse(layout, naDate)
	if err != nil {
		log.Println(err)
	}

	failDate := "2060-01-01"
	dateToFail, err := time.Parse(layout, failDate)
	if err != nil {
		log.Println(err)
	}

	// Fail the query
	if start == dateToFail {
		return rooms, errors.New("query processing failed")
	}

	// No availability
	if start.After(noAvailabiltyDate) {
		return rooms, nil
	}

	room := models.Room{
		ID: 1,
	}
	rooms = append(rooms, room)

	return rooms, nil
}

// GetRoomByID gets a room based on its ID
func (tr *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("error while getting room")
	}

	return room, nil
}

// GetUserByID returns a user by ID
func (tr *testDBRepo) GetUserByID(int) (models.User, error) {
	var u models.User

	return u, nil
}

// UpdateUser updates a user in the database
func (tr *testDBRepo) UpdateUser(models.User) error {
	return nil
}

// Authenticate authenticates a user
func (tr *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	return 0, "", nil
}

// AllReservations returns a slice of all the reservations
func (tr *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

// AllNewReservations returns a slice of all the reservations
func (tr *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

// GetReservationByID returns one reservation by ID
func (tr *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var res models.Reservation

	return res, nil
}

// UpdateReservation updates a reservation in the database
func (tr *testDBRepo) UpdateReservation(r models.Reservation) error {
	return nil
}

// DeleteReservation deletes a reservation in the database
func (tr *testDBRepo) DeleteReservation(id int) error {
	return nil
}

// UpdateProcessedForReservation updates processed for a reservation by ID
func (tg *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	return nil
}
