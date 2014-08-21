package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"code.google.com/p/go.blog/content/context/userip"
)

type EventStruct struct {
	Id        string
	Event     string
	IP        string
	URI       string
	Host      string
	Path      string
	Query     string
	UserAgent string
	Timestamp time.Time
	Action    string
}

var ipStore map[string]time.Time
var findStripper *regexp.Regexp
var mngoConnection MngoConnection

func main() {
	mngoConnection = MngoConnection{}
	mngoConnection.open()
	mngoConnection.Opensession = true
	findStripper = regexp.MustCompile("ObjectId|\\W")
	ipStore = make(map[string]time.Time)
	http.HandleFunc("/getTotal", getTotal)
	http.HandleFunc("/getSeries", getSeries)
	http.HandleFunc("/", handleEvent)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func saveIPToBlockList(ip string, timestamp time.Time) {
	// dummy function
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	log.Printf("saved IP %s to Blocklist, time: %s", ip, time.Since(timestamp))
}

// "intelligenter" IP Blocker
func ipBlocker(ip string) bool {
	block := ipStore[ip].After(time.Now().Add(-1 * time.Second * 3))
	if block {
		go saveIPToBlockList(ip, time.Now())
	}
	return block
}

// Handle HTTP Request
func handleEvent(w http.ResponseWriter, req *http.Request) {
	log.Println("Handle event")
	userIP, err := userip.FromRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if IP is requesting too frequently
	if ipBlocker(userIP.String()) {
		http.Error(w, "too many request from "+userIP.String(), http.StatusBadRequest)
		return
	} else {
		ipStore[userIP.String()] = time.Now()
	}

	requrl, _ := url.QueryUnescape(req.FormValue("uri"))
	parsedURI, _ := url.Parse(requrl)

	event := EventStruct{
		Id:        req.FormValue("id"),
		Event:     req.FormValue("event"),
		IP:        userIP.String(),
		URI:       requrl,
		Host:      parsedURI.Host,
		Path:      parsedURI.Path,
		Query:     parsedURI.RawQuery,
		UserAgent: req.UserAgent(),
		Timestamp: time.Now(),
		Action:    req.FormValue("action"),
	}
	if event.Id == "" || event.Event == "" {
		http.Error(w, "no query", http.StatusBadRequest)
		return
	}

	// write back to the user, thus the HTTP(s) request
	// doesn't block anything at client-side
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(nil)

	// save Event to all our cool backends / services
	saveEvent(event)
}

func saveEvent(event EventStruct) {
	saverChannel := make(chan bool)
	go func() { saverChannel <- SaveToMongoDB(mngoConnection, event) }()
	go func() { saverChannel <- PrintToStdout(event) }()

	var savings []bool
	for i := 0; i < 2; i++ {
		checkBool := <-saverChannel
		if checkBool {
			savings = append(savings, checkBool)
		}
	}
	if len(savings) == 2 {
		log.Println("EVERYTHING SAVED!")
	}
}

func getTotal(w http.ResponseWriter, req *http.Request) {
	id := findStripper.ReplaceAllString(req.FormValue("id"), "")
	days := findStripper.ReplaceAllString(req.FormValue("days"), "")
	daysInt, _ := strconv.Atoi(days)
	if daysInt < 1 {
		daysInt = 30
	}
	GetTotalById(mngoConnection, w, id, time.Hour*time.Duration(daysInt*24))
}
func getSeries(w http.ResponseWriter, req *http.Request) {
	id := findStripper.ReplaceAllString(req.FormValue("id"), "")
	days := findStripper.ReplaceAllString(req.FormValue("days"), "")
	daysInt, _ := strconv.Atoi(days)
	if daysInt < 1 {
		daysInt = 30
	}
	GetSeriesById(mngoConnection, w, id, time.Hour*time.Duration(daysInt*24))
}
