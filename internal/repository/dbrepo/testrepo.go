package dbrepo

import (
	"errors"
	"time"

	"github.com/tanishqv/bnb-bookings/internal/models"
)

func (tr *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (tr *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (tr *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for the roomID, and false if availability doesn't exist
func (tr *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	return false, nil
}

// SearchAvailabilityForAllRoomsByDates returns a slice of available rooms, if any, for any given date range
func (tr *testDBRepo) SearchAvailabilityForAllRoomsByDates(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room
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
