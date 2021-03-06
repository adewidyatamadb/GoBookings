package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/adewidyatamadb/GoBookings/internal/config"
	"github.com/adewidyatamadb/GoBookings/internal/driver"
	"github.com/adewidyatamadb/GoBookings/internal/models"
	"github.com/go-chi/chi/v5"
)

func TestNewRepo(t *testing.T) {
	var app config.AppConfig
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=root")
	if err != nil {
		log.Fatal("cannot connect to the database! dying...")
	}

	NewRepo(&app, db)
}

func TestGetHandlers(t *testing.T) {
	var tableTest = []struct {
		name               string
		url                string
		method             string
		expectedStatusCode int
	}{
		{"home page", "/", "GET", http.StatusOK},
		{"about page", "/about", "GET", http.StatusOK},
		{"general quarters page", "/generals-quarters", "GET", http.StatusOK},
		{"majors suite page", "/majors-suite", "GET", http.StatusOK},
		{"search availability page", "/search-availability", "GET", http.StatusOK},
		{"contact page", "/contact", "GET", http.StatusOK},
		{"non-existent", "/not-exist", "GET", http.StatusNotFound},
		//admin route
		{"login", "/user/login", "GET", http.StatusOK},
		{"logout", "/user/logout", "GET", http.StatusOK},
		{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
		{"new reservation", "/admin/reservations-new", "GET", http.StatusOK},
		{"all reservation", "/admin/reservations-all", "GET", http.StatusOK},
		{"show reservation", "/admin/reservations/new/1/show", "GET", http.StatusOK},
	}

	routes := getRoutes()

	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, test := range tableTest {
		resp, err := ts.Client().Get(ts.URL + test.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != test.expectedStatusCode {
			t.Errorf("for %s, expected status code %d  but got %d", test.name, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	var tableTest = []struct {
		name               string
		params             models.Reservation
		expectedStatusCode int
	}{
		{"intended case", models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		}, http.StatusOK},
		{"reservation not in session", models.Reservation{
			RoomID: 1000,
		}, http.StatusSeeOther},
		{"room not exist", models.Reservation{
			RoomID: 99,
			Room: models.Room{
				ID:       99,
				RoomName: "Not Exist",
			},
		}, http.StatusSeeOther},
	}

	for _, test := range tableTest {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if test.params.RoomID != 1000 {
			session.Put(ctx, "reservation", test.params)
		}

		handler := http.HandlerFunc(Repo.Reservation)

		handler.ServeHTTP(rr, req)

		if rr.Code != test.expectedStatusCode {
			t.Errorf("%s: reservation handler returned wrong response code: got %d, wanted %d", test.name, rr.Code, test.expectedStatusCode)
		}
	}
}

func TestRepository_PostReservation(t *testing.T) {
	type postData struct {
		key   string
		value string
	}

	var tableTest = []struct {
		name               string
		params             []postData
		expectedStatusCode int
	}{
		{"intended case", []postData{
			{key: "start_date", value: "11-10-2021"},
			{key: "end_date", value: "12-10-2021"},
			{key: "room_id", value: "1"},
			{key: "first_name", value: "John"},
			{key: "last_name", value: "Smith"},
			{key: "email", value: "john@smith.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusSeeOther},
		{"invalid arrival date", []postData{
			{key: "start_date", value: "invalid"},
			{key: "end_date", value: "12-10-2021"},
			{key: "room_id", value: "1"},
			{key: "first_name", value: "John"},
			{key: "last_name", value: "Smith"},
			{key: "email", value: "john@smith.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusSeeOther},
		{"invalid departure date", []postData{
			{key: "start_date", value: "12-10-2021"},
			{key: "end_date", value: "invalid"},
			{key: "room_id", value: "1"},
			{key: "first_name", value: "John"},
			{key: "last_name", value: "Smith"},
			{key: "email", value: "john@smith.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusSeeOther},
		{"invalid room id", []postData{
			{key: "start_date", value: "12-10-2021"},
			{key: "end_date", value: "13-10-2021"},
			{key: "room_id", value: "invalid"},
			{key: "first_name", value: "John"},
			{key: "last_name", value: "Smith"},
			{key: "email", value: "john@smith.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusSeeOther},
		{"room not exist", []postData{
			{key: "start_date", value: "12-10-2021"},
			{key: "end_date", value: "13-10-2021"},
			{key: "room_id", value: "99"},
			{key: "first_name", value: "John"},
			{key: "last_name", value: "Smith"},
			{key: "email", value: "john@smith.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusSeeOther},
		{"invalid user data", []postData{
			{key: "start_date", value: "12-10-2021"},
			{key: "end_date", value: "13-10-2021"},
			{key: "room_id", value: "1"},
			{key: "first_name", value: "a"},
			{key: "last_name", value: "Smith"},
			{key: "email", value: "john@smith.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusOK},
		{"failed to insert reservation data", []postData{
			{key: "start_date", value: "12-10-2021"},
			{key: "end_date", value: "13-10-2021"},
			{key: "room_id", value: "2"},
			{key: "first_name", value: "John"},
			{key: "last_name", value: "Smith"},
			{key: "email", value: "john@smith.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusSeeOther},
		{"failed to insert room restriction data", []postData{
			{key: "start_date", value: "12-10-2021"},
			{key: "end_date", value: "13-10-2021"},
			{key: "room_id", value: "100"},
			{key: "first_name", value: "John"},
			{key: "last_name", value: "Smith"},
			{key: "email", value: "john@smith.com"},
			{key: "phone", value: "555-555-5555"},
		}, http.StatusSeeOther},
		{"body missing", []postData{}, http.StatusSeeOther},
	}

	for _, test := range tableTest {
		postedData := url.Values{}
		var r *http.Request
		if test.name != "body missing" {
			postedData.Add("start_date", test.params[0].value)
			postedData.Add("end_date", test.params[1].value)
			postedData.Add("room_id", test.params[2].value)
			postedData.Add("first_name", test.params[3].value)
			postedData.Add("last_name", test.params[4].value)
			postedData.Add("email", test.params[5].value)
			postedData.Add("phone", test.params[6].value)
			req, err := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
			if err != nil {
				t.Error(err)
			}
			r = req
		} else {
			req, err := http.NewRequest("POST", "/make-reservation", nil)
			if err != nil {
				t.Error(err)
			}
			r = req
		}

		ctx := getCTX(r)
		r = r.WithContext(ctx)

		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostReservation)

		handler.ServeHTTP(w, r)

		if w.Code != test.expectedStatusCode {
			t.Errorf("%s: reservation handler returned wrong response code: got %d, wanted %d", test.name, w.Code, test.expectedStatusCode)
		}
	}
}

func TestRepository_ReservationSummary(t *testing.T) {
	var tableTest = []struct {
		name               string
		session            bool
		expectedStatusCode int
	}{
		{"intended case", true, http.StatusOK},
		{"cannot get session", false, http.StatusSeeOther},
	}

	for _, test := range tableTest {
		start := "11-11-2021"
		end := "12-11-2021"

		layout := "02-01-2006"
		startDate, _ := time.Parse(layout, start)
		endDate, _ := time.Parse(layout, end)

		reservation := models.Reservation{
			StartDate: startDate,
			EndDate:   endDate,
		}

		req, _ := http.NewRequest("GET", "/reservation-summary", nil)
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if test.session {
			session.Put(ctx, "reservation", reservation)
		}

		handler := http.HandlerFunc(Repo.ReservationSummary)

		handler.ServeHTTP(rr, req)
		if rr.Code != test.expectedStatusCode {
			t.Errorf("%s: reservation handler returned wrong response code: got %d, wanted %d", test.name, rr.Code, test.expectedStatusCode)
		}
	}
}

func TestRepository_PostAvailability(t *testing.T) {
	type postData struct {
		key   string
		value string
	}
	var tableTest = []struct {
		name               string
		params             []postData
		expectedStatusCode int
	}{
		{"intended case", []postData{
			{key: "start", value: "11-11-2021"},
			{key: "end", value: "12-11-2021"},
		}, http.StatusOK},
		{"cannot retrieve rooms data", []postData{
			{key: "start", value: "01-01-2021"},
			{key: "end", value: "12-11-2021"},
		}, http.StatusSeeOther},
		{"there are no room available", []postData{
			{key: "start", value: "30-12-2020"},
			{key: "end", value: "02-01-2021"},
		}, http.StatusSeeOther},
		{"failed parsing arrival date", []postData{
			{key: "start", value: "a"},
			{key: "end", value: "02-01-2021"},
		}, http.StatusSeeOther},
		{"failed parsing departure date", []postData{
			{key: "start", value: "30-12-2020"},
			{key: "end", value: "b"},
		}, http.StatusSeeOther},
		{"missing request body", []postData{}, http.StatusSeeOther},
	}

	for _, test := range tableTest {
		postedData := url.Values{}
		var r *http.Request
		if test.name != "missing request body" {
			postedData.Add("start", test.params[0].value)
			postedData.Add("end", test.params[1].value)

			req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(postedData.Encode()))
			r = req
		} else {
			req, _ := http.NewRequest("POST", "/search-availability", nil)
			r = req
		}

		ctx := getCTX(r)
		r = r.WithContext(ctx)

		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostAvailability)

		handler.ServeHTTP(w, r)

		if w.Code != test.expectedStatusCode {
			t.Errorf("case - %s: reservation handler returened wrong response code: got %d, wanted %d", test.name, w.Code, test.expectedStatusCode)
		}
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	var tableTest = []struct {
		name               string
		roomID             string
		sessionParams      models.Reservation
		session            bool
		expectedStatusCode int
	}{
		{"intended case", "1", models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		}, true, http.StatusSeeOther},
		{"room id is not number", "a", models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		}, true, http.StatusSeeOther},
		{"reservation not in the session", "1", models.Reservation{}, false, http.StatusSeeOther},
	}

	for _, test := range tableTest {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/choose-room", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", test.roomID)

		ctx := getCTX(r)

		r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		if test.session {
			session.Put(r.Context(), "reservation", test.sessionParams)
		}

		handler := http.HandlerFunc(Repo.ChooseRoom)

		handler.ServeHTTP(w, r)
		if w.Code != test.expectedStatusCode {
			t.Errorf("case - %s: reservation handler returned wrong response code: got %d, wanted %d", test.name, w.Code, http.StatusSeeOther)
		}
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	type postData struct {
		key   string
		value string
	}
	var tableTest = []struct {
		name       string
		params     []postData
		expectedOK bool
	}{
		{"room available", []postData{
			{key: "start", value: "11-10-2021"},
			{key: "end", value: "12-10-2021"},
			{key: "room_id", value: "1"},
		}, true},
		{"room not available", []postData{
			{key: "start", value: "11-10-2021"},
			{key: "end", value: "12-10-2021"},
			{key: "room_id", value: "3"},
		}, false},
		{"room id invalid", []postData{
			{key: "start", value: "11-10-2021"},
			{key: "end", value: "12-10-2021"},
			{key: "room_id", value: "a"},
		}, false},
		{"cannot retrieve data from the database", []postData{
			{key: "start", value: "11-10-2021"},
			{key: "end", value: "12-10-2021"},
			{key: "room_id", value: "100"},
		}, false},
		{"invalid form", []postData{}, false},
	}

	for _, test := range tableTest {
		// case - rooms are not available
		postedData := url.Values{}
		var r *http.Request
		if test.name != "invalid form" {
			postedData.Add("start", test.params[0].value)
			postedData.Add("end", test.params[1].value)
			postedData.Add("room_id", test.params[2].value)
			req, err := http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
			if err != nil {
				t.Error(err)
			}
			r = req
		} else {
			req, err := http.NewRequest("POST", "/search-availability-json", nil)
			if err != nil {
				t.Error(err)
			}
			r = req
		}

		ctx := getCTX(r)
		r = r.WithContext(ctx)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AvailabilityJSON)

		handler.ServeHTTP(w, r)

		var j jsonResponse
		err := json.Unmarshal(w.Body.Bytes(), &j)
		if err != nil {
			t.Error("failed to parse json")
		}
		if j.OK != test.expectedOK {
			t.Errorf("case - %s: reservation handler returned wrong response: got %v, wanted %v", test.name, j.OK, test.expectedOK)
		}
	}
}

func TestRepository_BookRoom(t *testing.T) {
	type parameters struct {
		key   string
		value string
	}
	var tableTest = []struct {
		name               string
		params             []parameters
		expectedStatusCode int
	}{
		{"intended case", []parameters{
			{key: "id", value: "1"},
			{key: "start_date", value: "11-11-2021"},
			{key: "end_date", value: "12-11-2021"},
		}, http.StatusSeeOther},
		{"invalid id", []parameters{
			{key: "id", value: "3"},
			{key: "start_date", value: "11-11-2021"},
			{key: "end_date", value: "12-11-2021"},
		}, http.StatusSeeOther},
	}

	for _, test := range tableTest {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/book-room?id="+test.params[0].value+"&s="+test.params[1].value+"&e="+test.params[2].value, nil)
		ctx := getCTX(r)
		r = r.WithContext(ctx)

		handler := http.HandlerFunc(Repo.BookRoom)

		handler.ServeHTTP(w, r)

		if w.Code != test.expectedStatusCode {
			t.Errorf("case - %s: reservation handler returned wrong response code: got %d, wanted %d", test.name, w.Code, test.expectedStatusCode)
		}
	}
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{"valid credential", "me@here.com", http.StatusSeeOther, "", "/"},
	{"invalid credential", "jack@here.com", http.StatusSeeOther, "", "/user/login"},
	{"invalid data", "j", http.StatusOK, `action="/user/login"`, ""},
}

func TestLogin(t *testing.T) {
	// range through all tests
	for _, test := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", test.email)
		postedData.Add("password", "password")

		// create a request
		req := httptest.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		// set the req header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(w, req)

		if w.Code != test.expectedStatusCode {
			t.Errorf("case - %s: expected code %d but got %d", test.name, test.expectedStatusCode, w.Code)
		}

		if test.expectedLocation != "" {
			// get the url from test
			actualLoc, _ := w.Result().Location()
			if actualLoc.String() != test.expectedLocation {
				t.Errorf("case - %s: expected location %s, but got location %s", test.name, test.expectedLocation, actualLoc.String())
			}
		}

		//checking for expected values in HTML
		if test.expectedHTML != "" {
			// read the response body into string
			html := w.Body.String()
			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf("case - %s: expected to find %s but did not", test.name, test.expectedHTML)
			}
		}
	}
}

func TestRepository_AdminPostShowReservation(t *testing.T) {
	var tableTest = []struct {
		name               string
		year               string
		month              string
		src                string
		expectedStatusCode int
		expectedLocation   string
	}{
		{"update from all reservations page", "", "", "all", http.StatusSeeOther, "/admin/reservations-all"},
		{"update from new reservations page", "", "", "new", http.StatusSeeOther, "/admin/reservations-new"},
		{"update from new reservations page", "2021", "11", "cal", http.StatusSeeOther, "/admin/reservations-calendar?y=2021&m=11"},
	}

	for _, test := range tableTest {
		postedData := url.Values{}
		postedData.Add("first_name", "John")
		postedData.Add("last_name", "John")
		postedData.Add("email", "John")
		postedData.Add("phone", "John")
		postedData.Add("month", test.month)
		postedData.Add("year", test.year)

		// create a request
		req := httptest.NewRequest("POST", fmt.Sprintf("/admin/reservations/%s/1", test.src), strings.NewReader(postedData.Encode()))
		ctx := getCTX(req)
		req = req.WithContext(ctx)

		// set the req header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostShowReservation)
		handler.ServeHTTP(w, req)

		if w.Code != test.expectedStatusCode {
			t.Errorf("case - %s: expected code %d but got %d", test.name, test.expectedStatusCode, w.Code)
		}

		if test.expectedLocation != "" {
			// get the url from test
			actualLoc, _ := w.Result().Location()
			if actualLoc.String() != test.expectedLocation {
				t.Errorf("case - %s: expected location %s, but got location %s", test.name, test.expectedLocation, actualLoc.String())
			}
		}
	}
}

func TestRepository_AdminReservationsCalendar(t *testing.T) {
	var tableTest = []struct {
		name               string
		month              string
		year               string
		expectedStatusCode int
	}{
		{"empty year and month", "", "", http.StatusOK},
		{"with year and month", "02", "2021", http.StatusOK},
	}

	for _, test := range tableTest {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", test.year, test.month), nil)
		ctx := getCTX(r)
		r = r.WithContext(ctx)

		handler := http.HandlerFunc(Repo.AdminReservationsCalendar)

		handler.ServeHTTP(w, r)
		if w.Code != test.expectedStatusCode {
			t.Errorf("case - %s: reservation handler returned wrong response code: got %d, wanted %d", test.name, w.Code, test.expectedStatusCode)
		}
	}
}

func TestRepository_AdminPostReservationsCalendar(t *testing.T) {
	// example session.Put(r.Context(), "reservation", test.sessionParams)
	var tableTest = []struct {
		name               string
		month              string
		year               string
		expectedStatusCode int
		expectedLocation   string
		removeBlock        bool
		addBlock           bool
	}{
		{"no block and reservations", "2", "2021", http.StatusSeeOther, "/admin/reservations-calendar?y=2021&m=2", false, false},
		{"removing block and adding block", "2", "2021", http.StatusSeeOther, "/admin/reservations-calendar?y=2021&m=2", true, true},
	}

	for _, test := range tableTest {
		blockMap := make(map[string]int)
		blockMap["11-2-2021"] = 1
		postedData := url.Values{}
		postedData.Add("m", test.month)
		postedData.Add("y", test.year)

		if test.removeBlock {
			postedData.Add("remove_block_1_11-2-2021", "")
		}

		if test.addBlock {
			postedData.Add("add_block_1_12-2-2021", "1")
		}

		// create a request
		req := httptest.NewRequest("POST", "/reservation-calendar", strings.NewReader(postedData.Encode()))
		ctx := getCTX(req)
		req = req.WithContext(ctx)
		session.Put(req.Context(), "block_map_1", blockMap)

		// set the req header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		handler.ServeHTTP(w, req)
		if w.Code != test.expectedStatusCode {
			t.Errorf("case - %s: expected code %d but got %d", test.name, test.expectedStatusCode, w.Code)
		}

		if test.expectedLocation != "" {
			// get the url from test
			actualLoc, _ := w.Result().Location()
			if actualLoc.String() != test.expectedLocation {
				t.Errorf("case - %s: expected location %s, but got location %s", test.name, test.expectedLocation, actualLoc.String())
			}
		}
	}
}

func TestRepository_AdminProcessReservation(t *testing.T) {
	var tableTest = []struct {
		name               string
		year               string
		month              string
		src                string
		expectedStatusCode int
		expectedLocation   string
	}{
		{"update from new reservations page", "", "", "new", http.StatusSeeOther, "/admin/reservations-new"},
		{"update from all reservations page", "", "", "all", http.StatusSeeOther, "/admin/reservations-all"},
		{"update from reservations calendar page", "2021", "2", "cal", http.StatusSeeOther, "/admin/reservations-calendar?y=2021&m=2"},
	}

	for _, test := range tableTest {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/%s/1/do?y=%s&m=%s", test.src, test.year, test.month), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		rctx.URLParams.Add("src", test.src)
		ctx := getCTX(r)
		r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		handler := http.HandlerFunc(Repo.AdminProcessReservation)

		handler.ServeHTTP(w, r)
		if w.Code != test.expectedStatusCode {
			t.Errorf("case - %s: reservation handler returned wrong response code: got %d, wanted %d", test.name, w.Code, test.expectedStatusCode)
		}

		if test.expectedLocation != "" {
			// get the url from test
			actualLoc, _ := w.Result().Location()
			if actualLoc.String() != test.expectedLocation {
				t.Errorf("case - %s: expected location %s, but got location %s", test.name, test.expectedLocation, actualLoc.String())
			}
		}
	}
}
func TestRepository_AdminDeleteReservation(t *testing.T) {
	var tableTest = []struct {
		name               string
		year               string
		month              string
		src                string
		expectedStatusCode int
		expectedLocation   string
	}{
		{"delete from new reservations page", "", "", "new", http.StatusSeeOther, "/admin/reservations-new"},
		{"delete from all reservations page", "", "", "all", http.StatusSeeOther, "/admin/reservations-all"},
		{"delete from reservations calendar page", "2021", "2", "cal", http.StatusSeeOther, "/admin/reservations-calendar?y=2021&m=2"},
	}

	for _, test := range tableTest {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("/admin/delete-reservation/%s/1/do?y=%s&m=%s", test.src, test.year, test.month), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		rctx.URLParams.Add("src", test.src)
		ctx := getCTX(r)
		r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		handler := http.HandlerFunc(Repo.AdminDeleteReservation)

		handler.ServeHTTP(w, r)
		if w.Code != test.expectedStatusCode {
			t.Errorf("case - %s: reservation handler returned wrong response code: got %d, wanted %d", test.name, w.Code, test.expectedStatusCode)
		}

		if test.expectedLocation != "" {
			// get the url from test
			actualLoc, _ := w.Result().Location()
			if actualLoc.String() != test.expectedLocation {
				t.Errorf("case - %s: expected location %s, but got location %s", test.name, test.expectedLocation, actualLoc.String())
			}
		}
	}
}

// getCTX get the context
func getCTX(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}
