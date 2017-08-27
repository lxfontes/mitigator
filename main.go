package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/handlers"
)

type TagMatcher map[string]bool

func (tf TagMatcher) Hit(input string) []string {
	ret := []string{}
	for txt, _ := range tf {
		if strings.Contains(input, txt) {
			ret = append(ret, txt)
		}
	}

	return ret
}

type DB struct {
	tags map[string]TagMatcher
}

type Score struct {
	Likely string              `json:"likely"`
	Hits   map[string][]string `json:"hits"`
}

func NewDB() *DB {
	return &DB{
		tags: map[string]TagMatcher{},
	}
}

func (db *DB) LoadFile(tag string, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	words := TagMatcher{}
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		bword := s.Text()
		words[strings.ToLower(bword)] = true
	}

	if err := s.Err(); err != nil {
		return err
	}

	db.tags[tag] = TagMatcher(words)
	return nil
}

func (db *DB) Score(input string) (*Score, error) {
	score := &Score{
		Hits: map[string][]string{},
	}

	totalHits := int64(0)
	tagHits := map[string]int{}

	for tag, tr := range db.tags {
		tagMatches := tr.Hit(input)
		if len(tagMatches) > 0 {
			for _, m := range tagMatches {
				score.Hits[m] = append(score.Hits[m], tag)
			}
			totalHits++
			tagHits[tag]++
		}
	}

	var highestHist int
	for tag, hits := range tagHits {
		if hits > highestHist {
			score.Likely = tag
			highestHist = hits
		}
	}

	return score, nil
}

type scoreHandler struct {
	db *DB
}

var _ http.Handler = &scoreHandler{}

type scoreRequest struct {
	Input string `json:"txt"`
}

func (sh *scoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req scoreRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "why u no json?", http.StatusBadRequest)
		return
	}

	score, err := sh.db.Score(req.Input)
	if err != nil {
		log.Println(err)
		http.Error(w, "too hard", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(score)
}

func main() {
	var addr string
	var staticDir string

	flag.StringVar(&addr, "addr", ":8080", "Listen address")
	flag.StringVar(&staticDir, "static-dir", "html", "HTML directory")
	flag.Parse()

	sh := &scoreHandler{db: NewDB()}
	sh.db.LoadFile("buzzword", "./data/buzzword")
	//	sh.db.LoadFile("english", "./data/all")

	r := http.NewServeMux()
	r.Handle("/score", sh)
	r.Handle("/", http.FileServer(http.Dir(staticDir)))

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		http.ListenAndServe(addr, loggedRouter)
		close(c)
	}()

	<-c
}
