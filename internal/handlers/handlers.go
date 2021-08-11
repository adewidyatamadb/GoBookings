package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/adewidyatamadb/GoBookings/internal/config"
	"github.com/adewidyatamadb/GoBookings/internal/driver"
	"github.com/adewidyatamadb/GoBookings/internal/forms"
	"github.com/adewidyatamadb/GoBookings/internal/models"
	"github.com/adewidyatamadb/GoBookings/internal/render"
	"github.com/adewidyatamadb/GoBookings/internal/repository"
	"github.com/adewidyatamadb/GoBookings/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
)

// Repo the repository used by the handlers
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

// NewTestRepo creates a new test repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.html", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.html", &models.TemplateData{})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	// get the reservation from session and cast it into models.Reservation
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "cannot get reservation data")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot find the room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	res.Room.RoomName = room.RoomName

	sd := res.StartDate.Format("02-01-2006")
	ed := res.EndDate.Format("02-01-2006")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot convert room id into integer!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	start := r.Form.Get("start_date")
	end := r.Form.Get("end_date")

	layout := "02-01-2006"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse arrival date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse departure date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot find the room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
		Room: models.Room{
			ID:       roomID,
			RoomName: room.RoomName,
		},
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "form is not valid", http.StatusTemporaryRedirect)
		render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot insert reservation into the database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot insert room restriction into the database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send notifications - to guest
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong>
		<br>
		Dear %s: <br>
		This is to confirm your reservation from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format("02-Jan-2006"), reservation.EndDate.Format("02-Jan-2006"))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "fort@smythe.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	// send notifications - owner
	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Notification</strong>
		<br>
		A reservation has been made for %s from %s to %s.
	`, reservation.Room.RoomName, reservation.StartDate.Format("02-Jan-2006"), reservation.EndDate.Format("02-Jan-2006"))

	msg = models.MailData{
		To:      "me@here.com",
		From:    "fort@smythe.com",
		Subject: "Reservation Notification",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "cannot get reservation from the session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := reservation.StartDate.Format("02-Jan-2006")
	ed := reservation.EndDate.Format("02-Jan-2006")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	m.App.Session.Remove(r.Context(), "reservation")
	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// Reservation renders the generals room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.html", &models.TemplateData{})
}

// Reservation renders the majors room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.html", &models.TemplateData{})
}

// Reservation renders the renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.html", &models.TemplateData{})
}

// PostAvailability renders the renders the search availability page
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "02-01-2006"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse arrival date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot parse departure date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot retrieve rooms data from the database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No Rooms Available!")
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

	render.Template(w, r, "choose-room.page.html", &models.TemplateData{
		Data: data,
	})
}

// ChooseRoom displays list of available room
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot convert room id!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// get the reservation from session and cast it into models.Reservation
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "cannot get reservation from session!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and send JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	// parsing req body
	err := r.ParseForm()
	if err != nil {
		//cannot parse form, so return appropiate json
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")
	layout := "02-01-2006"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		//cannot convert room id, so return appropiate json
		resp := jsonResponse{
			OK:      false,
			Message: "Error parsing room id",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		//cannot retrieve data from database, so return appropiate json
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to the database",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
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

// BookRoom takes URL parameters, builds a sessional variable, and takes user to make reservation page
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "02-01-2006"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot retrieve room data from the database!")
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

// Contact renders the renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.html", &models.TemplateData{})
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging the user
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.html", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}
