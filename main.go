package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type counters struct {
	sync.Mutex
	key   string // 'key' is in the form [data]:[timestamp].
	view  int
	click int
}

// Used for implementing the global limit for the stats handler.
type limiter struct {
	sync.Mutex
	curr int
}

var (
	max     int = 50 // Global limit on requests for stats handler.
	content     = []string{"sports", "entertainment", "business", "education"}
	c           = make([]counters, cMax*len(content))
	limit       = limiter{}
	size    int = 0  // Start with no counters stored.
	cMax    int = 30 // Number of minutes worth of counters stored.
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	timeCurrent := []rune(time.Now().Format("2006.01.02 15:04:05"))

	fmt.Fprint(w, "Welcome to EQ Works 😎", "\n")
	fmt.Fprint(w, "current time: ", string(timeCurrent), "\n")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	data := content[rand.Intn(len(content))]

	var count *counters

	if size != 0 {
		for i := 0; i < len(content); i++ {
			if c[i].key[0:strings.Index(c[i].key, ":")] == data { //
				count = &c[i]
				break
			}
		}

		fmt.Fprint(w, data, " page!", "\n", "\n")

		count.Lock()
		count.view++
		count.Unlock()

		for i := 0; i < size*len(content); i++ {
			fmt.Fprint(w, c[i].key, ", views: ", c[i].view, ", clicks:", c[i].click, "\n")
		}

	} else {
		fmt.Fprint(w, "Error, no counters", "\n") // Displayed if no counters are initialized (should never happen).
	}

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

	var count *counters

	for i := 0; i < len(content); i++ {
		if c[i].key[0:strings.Index(c[i].key, ":")] == data { // check
			count = &c[i]
			break
		}
	}

	(*count).Lock()
	(*count).click++
	(*count).Unlock()

	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) { // mock store queries
	if !isAllowed() {
		w.WriteHeader(429)
		return
	} else {
		// display counters to the page.
		limit.Lock()
		limit.curr++
		limit.Unlock()

		fmt.Fprint(w, "Stats page", "\n", "\n")
		fmt.Fprint(w, "Requests: ", limit.curr, "/", max, "\n", "\n")

		for i := 0; i < size*len(content); i++ {
			fmt.Fprint(w, c[i].key, ", views: ", c[i].view, ", clicks:", c[i].click, "\n")
		}

	}
}

func isAllowed() bool {
	if limit.curr >= max {
		return false
	} else {
		return true
	}
}

func uploadCounters() error { // upload to mock store every 5 seconds

	for {
		file, err := os.Create("mockstore.txt")
		if err != nil {
			log.Fatal("Error uploading to mock store", err)
		}
		for i := 0; i < size*len(content); i++ {
			viewStr := strconv.Itoa(c[i].view)
			clickStr := strconv.Itoa(c[i].click)

			fmt.Fprintf(file, c[i].key+"|"+viewStr+","+clickStr+"\n")
		}
		file.Close()
		time.Sleep(5 * time.Second)
	}

}

// Resets the stats handler request count every minute.
func rateLimiter() {
	for {
		limit.curr = 0
		time.Sleep(time.Minute)
	}
}

// Counters are organized so that the first n-indexes (where n is the length of content) are the counters for the current minute, the next n-indexes are the next minute, etc...
// Older counters slowly get pushed to the later indexes and are eventually removed.
func cycleCounters() {
	timeCurrent := []rune(time.Now().Format("2006.01.02 15:04:05"))
	for {
		timeCurrent = []rune(time.Now().Format("2006.01.02 15:04:05"))
		// If size is 0 (no counters), create the first set of counters regardless of the current time.
		// Otherwise, wait until time = HH:MM:00.
		if size > 0 {
			if string(timeCurrent[16:19]) == ":00" {
				if size < cMax {
					size++
				}
				if string(timeCurrent[0:16]) != c[0].key[strings.Index(c[0].key, ":")+1:strings.Index(c[0].key, ":")+17] {
					for i := size; i > 1; i-- {
						for j := 0; j < len(content); j++ {
							c[i*len(content)-j-1] = c[i*len(content)-j-len(content)-1]
						}
					}

					for i := 0; i < len(content); i++ {
						c[i] = counters{key: content[i] + ":" + string(timeCurrent[0:16]), click: 0, view: 0}
					}
				}
			}
		} else {
			for i := 0; i < len(content); i++ {
				c[i] = counters{key: content[i] + ":" + string(timeCurrent[0:16]), click: 0, view: 0}
			}
			size++
		}
		time.Sleep(time.Second)
	}
}

func loadData() {
	file, err := os.Open("mockstore.txt")

	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		newKey := line[0:strings.Index(line, "|")]
		newView, err := strconv.Atoi(line[strings.Index(line, "|")+1 : strings.Index(line, ",")])
		newClick, err := strconv.Atoi(line[strings.Index(line, ",")+1:])

		if err == nil {
			c[count] = counters{key: newKey, view: newView, click: newClick}
			if count%len(content) == 0 {
				size++
			}
			count++
		}
	}

}

func main() {

	loadData()

	go cycleCounters()
	go uploadCounters()
	go rateLimiter()

	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/stats/", statsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
