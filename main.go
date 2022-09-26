package main

import (
	"context"
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"ruz-ics/gruz"
	"strconv"
	"strings"
	"time"
)

var client = gruz.NewClient(http.DefaultClient)

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "templates/html/main.html")
}

func Calendar(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	id := ps.ByName("id")
	if len(id) > 10 {
		http.Error(w, "id too long", http.StatusBadRequest)
		return
	}
	userId, err := strconv.Atoi(id)
	if err != nil {
		return
	}
	lessons, err := client.GetSchedule(context.Background(), int64(userId), gruz.StudentPerson, time.Now().AddDate(0, 0, -3), time.Now().AddDate(0, 0, 14), gruz.RussianLanguage)
	if err != nil {
		log.Println(err)
	}
	cal := ics.NewCalendar()
	cal.SetName("Расписание занятий ВШЭ")
	cal.SetXPublishedTTL("PT7D")
	cal.SetRefreshInterval("PT7D")
	for _, l := range lessons {
		event := cal.AddEvent(strconv.Itoa(l.LessonOid))
		start, err := time.Parse("2006.01.02-15:04", l.Date+"-"+l.BeginLesson)
		if err != nil {
			log.Println(err)
		}
		end, err := time.Parse("2006.01.02-15:04", l.Date+"-"+l.EndLesson)
		if err != nil {
			log.Println(err)
		}
		event.SetSummary(fmt.Sprintf("%s • %s • %s", l.Discipline, l.KindOfWork, l.LecturerTitle))
		event.SetStartAt(start)
		event.SetEndAt(end)
		event.SetLocation(l.Auditorium + ", " + l.Building)
	}
	w.Header().Set("Content-Type", "text/calendar")
	w.Header().Set("Content-Disposition", "attachment; filename=ruz.ics")
	fmt.Fprint(w, cal.Serialize())
}

func Info(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if !strings.HasSuffix(email, "@edu.hse.ru") {
		http.Error(w, "email too long", http.StatusBadRequest)
		return
	}
	userId, err := client.GetMailInfo(context.Background(), email)
	if err != nil {
		log.Println(err)
	}
	userJson, err := json.Marshal(userId)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(w, string(userJson))
}

func main() {
	SentryDsn := os.Getenv("SENTRY_DSN")
	if SentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              SentryDsn,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}
	sentryHandler := sentryhttp.New(sentryhttp.Options{})
	router := httprouter.New()
	router.ServeFiles("/css/*filepath", http.Dir("templates/css"))
	router.HandlerFunc("GET", "/", sentryHandler.HandleFunc(Index))
	router.HandlerFunc("GET", "/cal/:id", sentryHandler.HandleFunc(Calendar))
	router.HandlerFunc("GET", "/info", sentryHandler.HandleFunc(Info))
	log.Fatal(http.ListenAndServe(":8001", router))

}
