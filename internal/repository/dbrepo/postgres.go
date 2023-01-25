package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/tanishqv/bnb-bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (pgr *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (pgr *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int
	stmt := `INSERT INTO reservations (first_name, last_name, email, phone, start_date,
		end_date, room_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	err := pgr.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (pgr *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO room_restrictions (start_date, end_date, room_id, reservation_id,
		created_at, updated_at, restriction_id)
		VALUES
		($1, $2, $3, $4, $5, $6, $7)`

	_, err := pgr.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)
	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for the roomID, and false if availability doesn't exist
func (pgr *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := ` SELECT COUNT(id)
			   FROM room_restrictions
			   WHERE
			   room_id = $1
			   AND
			   $2 < end_date AND $3 > start_date;`

	row := pgr.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)

	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

// SearchAvailabilityForAllRoomsByDates returns a slice of available rooms, if any, for any given date range
func (pgr *postgresDBRepo) SearchAvailabilityForAllRoomsByDates(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := ` SELECT r.id, r.room_name
			   FROM rooms r
			   WHERE r.id NOT IN (
					SELECT rr.room_id
					FROM room_restrictions rr
					WHERE $1 < rr.end_date AND $2 > rr.start_date
			   );`

	rows, err := pgr.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var room models.Room

		err = rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// AllRooms returns a slice of all rooms in the database
func (pgr *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `SELECT id, room_name, created_at, updated_at
			  FROM rooms
			  ORDER BY room_name`

	rows, err := pgr.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}
	defer rows.Close()

	for rows.Next() {
		var rm models.Room

		err := rows.Scan(
			&rm.ID,
			&rm.RoomName,
			&rm.CreatedAt,
			&rm.UpdatedAt,
		)
		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, rm)
	}
	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRoomByID gets a room based on its ID
func (pgr *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `SELECT id, room_name, created_at, updated_at
			  FROM rooms
			  WHERE id = $1`

	row := pgr.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err != nil {
		return room, err
	}

	return room, nil
}

// GetUserByID returns a user by ID
func (pgr *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	query := `SELECT id, first_name, last_name, email, password, access_level, created_at, updated_at
			  FROM users
			  WHERE id = $1`

	row := pgr.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.AccessLevel,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return user, err
	}

	return user, nil
}

// UpdateUser updates a user in the database
func (pgr *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE users
			  SET
			  first_name = $1,
			  last_name = $2,
			  email = $3,
			  access_level = $4,
			  updated_at = $5`

	_, err := pgr.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

// Authenticate authenticates a user
func (pgr *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	query := `SELECT id, password
			  FROM users
			  WHERE email = $1`

	row := pgr.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

// AllReservations returns a slice of all the reservations
func (pgr *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone, 
					 r.start_date, r.end_date, r.room_id, r.created_at, 
					 r.updated_at, r.processed, 
					 rooms.id, rooms.room_name
			  FROM reservations r
			  LEFT JOIN rooms
			  ON r.room_id = rooms.id
			  ORDER BY r.start_date ASC`

	rows, err := pgr.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err = rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// AllNewReservations returns a slice of all the reservations
func (pgr *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone, 
					 r.start_date, r.end_date, r.room_id, r.created_at, 
					 r.updated_at, rooms.id, rooms.room_name
			  FROM reservations r
			  LEFT JOIN rooms
			  ON r.room_id = rooms.id
			  WHERE processed = 0
			  ORDER BY r.start_date ASC`

	rows, err := pgr.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err = rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// GetReservationByID returns one reservation by ID
func (pgr *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone,
			 r.start_date, r.end_date, r.room_id,
			 r.created_at, r.updated_at, r.processed,
			 rooms.id, rooms.room_name
			 FROM reservations r
			 LEFT JOIN rooms
			 ON r.room_id = rooms.id
			 WHERE r.id = $1`

	row := pgr.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Room.ID,
		&res.Room.RoomName,
	)
	if err != nil {
		return res, err
	}

	return res, nil
}

// UpdateReservation updates a reservation in the database
func (pgr *postgresDBRepo) UpdateReservation(r models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE reservations
			  SET
			  first_name = $1,
			  last_name = $2,
			  email = $3,
			  phone = $4,
			  updated_at = $5
			  WHERE id = $6`

	_, err := pgr.DB.ExecContext(ctx, query,
		r.FirstName,
		r.LastName,
		r.Email,
		r.Phone,
		time.Now(),
		r.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// DeleteReservation deletes a reservation in the database
func (pgr *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM reservations
			  WHERE id = $1`

	_, err := pgr.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProcessedForReservation updates processed for a reservation by ID
func (pgr *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE reservations
			  SET processed = $1
			  WHERE id = $2`

	_, err := pgr.DB.ExecContext(ctx, query, processed, id)
	if err != nil {
		return err
	}

	return nil
}

// GetRestrictionsForRoomByDate returns restrictions for a room by date range
func (pgr *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `SELECT id, COALESCE(reservation_id, 0), restriction_id, room_id, start_date, end_date
			  FROM room_restrictions
			  WHERE
			  $1 < end_date AND $2 >= start_date
			  AND
			  room_id = $3`

	rows, err := pgr.DB.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err = rows.Scan(
			&r.ID,
			&r.ReservationID,
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)
		if err != nil {
			return nil, err
		}

		restrictions = append(restrictions, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return restrictions, nil
}

// InsertBlockForRoom inserts a room restriction
func (pgr *postgresDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO room_restrictions
			  (start_date, end_date, room_id, restriction_id, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := pgr.DB.ExecContext(ctx, query, startDate, startDate.AddDate(0, 0, 1), id, 2, time.Now(), time.Now())
	if err != nil {
		return err
	}

	return nil
}

// DeleteBlockByID deletes a room restriction
func (pgr *postgresDBRepo) DeleteBlockByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM room_restrictions
			  WHERE id=$1`

	_, err := pgr.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
