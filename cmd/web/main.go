package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adewidyatamadb/GoBookings/internal/config"
	"github.com/adewidyatamadb/GoBookings/internal/driver"
	"github.com/adewidyatamadb/GoBookings/internal/handlers"
	"github.com/adewidyatamadb/GoBookings/internal/helpers"
	"github.com/adewidyatamadb/GoBookings/internal/models"
	"github.com/adewidyatamadb/GoBookings/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var server = "localhost"
var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

//main is the main application function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()
	// http.HandleFunc("/", handlers.Repo.Home)
	// http.HandleFunc("/about", handlers.Repo.About)

	fmt.Printf("Starting application on %s%s", server, portNumber)
	// _ = http.ListenAndServe(server+portNumber, nil)

	srv := &http.Server{
		Addr:    server + portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// what am I going to put in the session
	gob.Register(models.Reservation{})

	// change this to true when in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to database
	log.Println("Connecting to the database...")

	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=root")
	if err != nil {
		log.Fatal("cannot connect to the database! dying...")
	}

	log.Println("Connected to the database!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewTemplates(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
