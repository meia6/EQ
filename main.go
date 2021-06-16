package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type counters struct {
	//locker  sync.Mutex
	sync.Mutex
	view  int
	click int
}

type limiter struct {
	sync.Mutex
	curr int
}

var (
	//c           = counters{}
	max int = 500 // arbitrary global limit
	//curr    int = 0
	content     = []string{"sports", "entertainment", "business", "education"}
	c           = []counters{}
	limit       = limiter{}
	cMax    int = 60 // 60 minutes of counters stored
	store       = map[string]counters{}
)

func refreshCounters(c []counters) {
	c = []counters{}

	for i := 0; i < len(content); i++ {
		c = append(c, counters{})
	}
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	data := content[rand.Intn(len(content))]

	timeCurrent := []rune(time.Now().Format("2006.01.02 15:04:05"))
	conc := data + ":" + string(timeCurrent[0:16])
	count := store[conc] // <- counters
	count.Lock()

	count.view++
	fmt.Fprint(w, "Click count:", count.click, " View count:", count.view, "\n")
	fmt.Fprint(w, "current requests: ", limit.curr, "/", max, "\n")
	fmt.Fprint(w, conc) //take [0:16] to extract everything except for seconds
	/*
		for timestamp, counts := range store {
			if val, ok := dict[]; ok{

			}
		}
	*/

	count.Unlock()

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}
	// simulate random click call
	if rand.Intn(100) < 50 {
		processClick(data)
	}

}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {

	timeCurrent := []rune(time.Now().Format("2001.01.01 01:01:01"))
	count := store[data+string(timeCurrent[0:16])]
	//locker := count.locker

	count.Lock()

	count.click++

	count.Unlock()

	return nil
}

// Arbitrarily set max number of map values to 60? for 60 minutes data stored

func statsHandler(w http.ResponseWriter, r *http.Request) { // mock store queries
	if !isAllowed() {
		w.WriteHeader(429)
		return
	} else {
		// display counters to the page.
		for i := 0; i < len(c); i++ { // loop for up to 60 times, displaying all 4 categories

			for j := 0; j < len(content); j++ {

			}
		}
		limit.Lock()
		limit.curr++
		limit.Unlock()
	}
}

func isAllowed() bool { // change this block for global rate limiting
	if limit.curr >= max {
		return false
	} else {
		return true
	}

}

func uploadCounters() error { // upload to mock store every 5 seconds
	for {
		// upload counters to the mock store
		time.Sleep(5 * time.Second)
	}
	return nil
}

func rateLimiter() {
	for {

		limit.curr = 0
		time.Sleep(time.Minute) // max of 'limit' queries per minute
	}
}

func cycleCounters() {

	timeCurrent := []rune(time.Now().Format("2001.01.01 01:01:01"))

	for {
		timeCurrent = []rune(time.Now().Format("2001.01.01 01:01:01"))

		for i := 0; i < len(content); i++ {
			key := content[i] + ":" + string(timeCurrent[0:16]) // key contains the content name and the date/time by minute
			store[key] = counters{}
		}
		time.Sleep(time.Minute)
	}
}

func main() {

	//go cycleCounters()
	go uploadCounters()
	go rateLimiter()

	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/stats/", statsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
