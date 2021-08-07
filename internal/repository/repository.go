package repository

import "github.com/adewidyatamadb/GoBookings/internal/models"

type DatabaseRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) error
}
