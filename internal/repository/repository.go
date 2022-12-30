package repository

import (
	"time"

	"github.com/tanishqv/bnb-bookings/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(models.Reservation) (int, error)
	InsertRoomRestriction(models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(time.Time, time.Time, int) (bool, error)
	SearchAvailabilityForAllRoomsByDates(time.Time, time.Time) ([]models.Room, error)
	GetRoomByID(int) (models.Room, error)
}
