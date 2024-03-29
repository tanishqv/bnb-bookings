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
	AllRooms() ([]models.Room, error)
	GetRoomByID(int) (models.Room, error)

	GetUserByID(int) (models.User, error)
	UpdateUser(models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	AllReservations() ([]models.Reservation, error)
	AllNewReservations() ([]models.Reservation, error)
	GetReservationByID(int) (models.Reservation, error)
	UpdateReservation(models.Reservation) error
	DeleteReservation(int) error
	UpdateProcessedForReservation(int, int) error
	GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)
	InsertBlockForRoom(int, time.Time) error
	DeleteBlockByID(int) error
}
