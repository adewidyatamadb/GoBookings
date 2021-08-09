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
	return false, nil
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
	} else {
		rooms = append(rooms, models.Room{
			ID:       1,
			RoomName: "General's Quarter",
		})

		return rooms, nil
	}

}

// GetRoomByID get a room by id
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("some error")
	}
	return room, nil
}
