package handlers

import (
	"context"
	"encoding/json"
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

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"/", "/", "GET", http.StatusOK},
	{"/about", "/about", "GET", http.StatusOK},
	{"/general-quarters", "/generals-quarters", "GET", http.StatusOK},
	{"/majors-suite", "/majors-suite", "GET", http.StatusOK},
	{"/search-availability", "/search-availability", "GET", http.StatusOK},
	{"/contact", "/contact", "GET", http.StatusOK},
}

func TestNewRepo(t *testing.T) {
	var app config.AppConfig
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=root")
	if err != nil {
		log.Fatal("cannot connect to the database! dying...")
	}

	NewRepo(&app, db)
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()

	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s, expected status code %d  but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCTX(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("reservation handler returened wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// test case where reservation is not in session(reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returened wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test with non existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCTX(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returened wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	postedData := url.Values{}
	postedData.Add("start_date", "11-10-2021")
	postedData.Add("end_date", "12-10-2021")
	postedData.Add("room_id", "1")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "555-555-5555")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx := getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("reservation handler returened wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// case if the post body is missing
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returened wrong response code for missing post body: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid arrival date
	postedData = url.Values{}
	postedData.Add("start_date", "invalid")
	postedData.Add("end_date", "12-10-2021")
	postedData.Add("room_id", "1")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "555-555-5555")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returened wrong response code for invalid arrival date: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid departure date
	postedData = url.Values{}
	postedData.Add("start_date", "11-10-2021")
	postedData.Add("end_date", "invalid")
	postedData.Add("room_id", "1")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "555-555-5555")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returened wrong response code for invalid departure date: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid room id
	postedData = url.Values{}
	postedData.Add("start_date", "11-10-2021")
	postedData.Add("end_date", "12-10-2021")
	postedData.Add("room_id", "a")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "555-555-5555")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returned wrong response code for invalid room id: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid data
	postedData = url.Values{}
	postedData.Add("start_date", "11-10-2021")
	postedData.Add("end_date", "12-10-2021")
	postedData.Add("room_id", "1")
	postedData.Add("first_name", "a")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "555-555-5555")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("reservation handler returned wrong response code for invalid data: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test for failed insert reservation data
	postedData = url.Values{}
	postedData.Add("start_date", "11-10-2021")
	postedData.Add("end_date", "12-10-2021")
	postedData.Add("room_id", "2")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "555-555-5555")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returned wrong response code for failed insert data: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for failed insert room restriction data
	postedData = url.Values{}
	postedData.Add("start_date", "11-10-2021")
	postedData.Add("end_date", "12-10-2021")
	postedData.Add("room_id", "100")
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@smith.com")
	postedData.Add("phone", "555-555-5555")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returned wrong response code for failed insert data: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_ReservationSummary(t *testing.T) {
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
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ReservationSummary)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// test case if data cannot get pulled from session
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returened wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

}

func TestRepository_PostAvailability(t *testing.T) {
	// test case valid condition
	postedData := url.Values{}
	postedData.Add("start", "11-11-2021")
	postedData.Add("end", "12-11-2021")

	req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(postedData.Encode()))
	ctx := getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("reservation handler returened wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// case if post missing request body
	req, _ = http.NewRequest("POST", "/search-availability", nil)
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returened wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// case cannot retrieve rooms data from the database
	postedData = url.Values{}
	postedData.Add("start", "01-01-2021")
	postedData.Add("end", "12-11-2021")

	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(postedData.Encode()))
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returened wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// case there is no room available
	postedData = url.Values{}
	postedData.Add("start", "30-12-2020")
	postedData.Add("end", "02-01-2021")

	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(postedData.Encode()))
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostAvailability)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("reservation handler returened wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	// intended case
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/choose-room", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")

	ctx := getCTX(r)

	r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

	session.Put(r.Context(), "reservation", reservation)

	handler := http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(w, r)
	if w.Code != http.StatusSeeOther {
		t.Errorf("reservation handler returned wrong response code: got %d, wanted %d", w.Code, http.StatusSeeOther)
	}

	//test case if id is not integer
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/choose-room", nil)

	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "a")

	ctx = getCTX(r)

	r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

	session.Put(r.Context(), "reservation", reservation)

	handler = http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(w, r)
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returned wrong response code: got %d, wanted %d", w.Code, http.StatusTemporaryRedirect)
	}

	// test case if there is no reservation data in the session
	//test case if id is not integer
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/choose-room", nil)

	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")

	ctx = getCTX(r)

	r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

	session.Put(r.Context(), "non-existent", reservation)

	handler = http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(w, r)
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returned wrong response code: got %d, wanted %d", w.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	// case - rooms are not available
	postedData := url.Values{}
	postedData.Add("start", "11-10-2021")
	postedData.Add("start", "12-10-2021")
	postedData.Add("room_id", "3")

	req, err := http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
	if err != nil {
		t.Error(err)
	}
	ctx := getCTX(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.AvailabilityJSON)

	handler.ServeHTTP(rr, req)

	var j jsonResponse
	err = json.Unmarshal(rr.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse json")
	}
	if j.OK {
		t.Error("room supposed to be unavailable")
	}

	// case - room is available
	postedData = url.Values{}
	postedData.Add("start", "11-10-2021")
	postedData.Add("start", "12-10-2021")
	postedData.Add("room_id", "1")

	req, err = http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
	if err != nil {
		t.Error(err)
	}
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse json")
	}

	if !j.OK {
		t.Error("room supposed to be available")
	}

	// case - invalid form
	req, _ = http.NewRequest("POST", "/search-availability-json", nil)
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse json")
	}

	if j.OK {
		t.Error("form supposed to be cannot be parsed")
	}

	// case - invalid room id
	postedData = url.Values{}
	postedData.Add("start", "11-10-2021")
	postedData.Add("start", "12-10-2021")
	postedData.Add("room_id", "a")

	req, err = http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
	if err != nil {
		t.Error(err)
	}
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse json")
	}

	if j.OK {
		t.Error("room id supposed to be cannot be converted")
	}

	// case - cannot retrieve data from the database
	postedData = url.Values{}
	postedData.Add("start", "11-10-2021")
	postedData.Add("start", "12-10-2021")
	postedData.Add("room_id", "100")

	req, err = http.NewRequest("POST", "/search-availability-json", strings.NewReader(postedData.Encode()))
	if err != nil {
		t.Error(err)
	}
	ctx = getCTX(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse json")
	}

	if j.OK {
		t.Error("data supposed to be cannot be retrieved from the database")
	}
}

func TestRepository_BookRoom(t *testing.T) {
	// intended case
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/book-room?id=1&s=11-11-2021&e=12-11-2021", nil)
	ctx := getCTX(r)
	r = r.WithContext(ctx)

	handler := http.HandlerFunc(Repo.BookRoom)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("reservation handler returned wrong response code: got %d, wanted %d", w.Code, http.StatusSeeOther)
	}

	// test case if id is invalid
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/book-room?id=3&s=11-11-2021&e=12-11-2021", nil)
	ctx = getCTX(r)
	r = r.WithContext(ctx)

	handler = http.HandlerFunc(Repo.BookRoom)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returned wrong response code: got %d, wanted %d", w.Code, http.StatusTemporaryRedirect)
	}
}

func getCTX(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}
