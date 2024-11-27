package main

import (
	"context"
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"ruz-ics/gruz"
	"strconv"
	"strings"
	"time"
)

// add http proxy
var client = gruz.NewClient(&http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	},
})

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
	loc, _ := time.LoadLocation("Europe/Moscow") //TODO: make it depend on building
	for _, l := range lessons {
		event := cal.AddEvent(strconv.Itoa(l.LessonOid))
		start, err := time.ParseInLocation("2006.01.02-15:04", l.Date+"-"+l.BeginLesson, loc)
		if err != nil {
			log.Println(err)
		}
		end, err := time.ParseInLocation("2006.01.02-15:04", l.Date+"-"+l.EndLesson, loc)
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
	log.Println(fmt.Sprintf("GET /cal/%s", id))
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
	log.Println(fmt.Sprintf("GET /info %s", email))
}

func main() {
	router := httprouter.New()
	router.ServeFiles("/css/*filepath", http.Dir("templates/css"))
	router.HandlerFunc("GET", "/", Index)
	router.HandlerFunc("GET", "/cal/:id", Calendar)
	router.HandlerFunc("GET", "/info", Info)
	log.Println("Listening on :8001")
	log.Fatal(http.ListenAndServe(":8001", router))

}
