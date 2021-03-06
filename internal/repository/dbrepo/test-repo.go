package dbrepo

import (
	"errors"
	"time"

	"github.com/adewidyatamadb/GoBookings/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation insert a reservation into the database
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// if the room id is not 1, then fail, otherwise, pass
	if res.RoomID == 2 {
		return 0, errors.New("some error")
	}
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *testDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	// if the room id is not 1, then fail, otherwise, pass
	if res.RoomID == 100 {
		return errors.New("some error")
	}
	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if room available and return false if room is not available
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	if roomID == 3 {
		return false, nil
	} else if roomID == 100 {
		return false, errors.New("some error")
	} else {
		return true, nil
	}
}

// SearchAvailabilityForAllRooms returns a slice of available rooms if any for given date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	layout := "02-01-2006"
	startDate, _ := time.Parse(layout, "01-01-2021")
	endDate, _ := time.Parse(layout, "02-01-2021")

	if start == startDate {
		return nil, errors.New("some error")
	} else if end == endDate {
		return rooms, nil
	}
	rooms = append(rooms, models.Room{
		ID:       1,
		RoomName: "General's Quarter",
	})

	return rooms, nil

}

// GetRoomByID get a room by id
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 && id < 100 {
		return room, errors.New("some error")
	}
	return room, nil
}

// GetUserByID retrieve user data from the database using id
func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User

	return u, nil
}

// UpdateUser update user data in the database
func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

// Authentice authenticates a user
func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	if email == "me@here.com" {
		return 1, "", nil
	}
	return 0, "", errors.New("some error")
}

// GetAllReservations returns a slice of all reservations
func (m *testDBRepo) GetAllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

// GetAllNewReservations returns a slice of all reservations
func (m *testDBRepo) GetAllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

// GetReservationByID returns reservation by id
func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var res models.Reservation
	return res, nil
}

// UpdateReservation update reservation data in the database
func (m *testDBRepo) UpdateReservation(u models.Reservation) error {
	return nil
}

// DeleteReservation deletes one reservation by id
func (m *testDBRepo) DeleteReservation(id int) error {
	return nil
}

// UpdateProcessedForReservation update processed for a reservation by id
func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	return nil
}

func (m *testDBRepo) GetAllRooms() ([]models.Room, error) {
	var rooms = []models.Room{
		{ID: 1},
	}

	return rooms, nil
}

// GetRestrictionForRoomByDate returns restrictions for a room by date range
func (m *testDBRepo) GetRestrictionForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {

	var restrictions = []models.RoomRestriction{
		{
			ID:            1,
			StartDate:     start.AddDate(0, 0, 2),
			EndDate:       end.AddDate(0, 0, 4),
			ReservationID: 1,
		},
		{
			ID:            2,
			StartDate:     start.AddDate(0, 0, 6),
			EndDate:       start.AddDate(0, 0, 6),
			ReservationID: 0,
		},
	}

	return restrictions, nil

}

// InsertBlockForRoom insert a room restriction data
func (m *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	return nil
}

// DeleteBlockByID deletes a room restriction
func (m *testDBRepo) DeleteBlockByID(id int) error {
	return nil
}
