package movie_server

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	introFolder   = "/movies/intro"
	counterFolder = "/movies/counter"
	printFolder   = "/movies/printing"
	shotsFolder   = "/movies/shots"
)

var (
	mov *moviesSet
)

type movieList struct {
	movies       []string
	currentMovie int
	queryTime    time.Time
}

type shotList struct {
	startMovies  []string
	endMovies    []string
	currentMovie int
}

type shotsSet struct {
	shot      []shotList
	queryTime time.Time
}

type moviesSet struct {
	introMovies    movieList
	printingMovies movieList
	counterMovies  movieList
	shotMovies     shotsSet
	mediaFolder    string
}

func init() {
	mov = new(moviesSet)
	mov.init()
}

func (m *moviesSet) init() {
	flag.StringVar(&m.mediaFolder, "media", "~/media", "The media folder path")
	flag.Parse()
	m.introMovies.init(path.Join(m.mediaFolder, introFolder))
	m.printingMovies.init(path.Join(m.mediaFolder, printFolder))
	m.counterMovies.init(path.Join(m.mediaFolder, counterFolder))
	m.shotMovies.init(path.Join(m.mediaFolder, shotsFolder))
}

func (m *movieList) init(folder string) {
	m.movies = getMoviesList(folder)
	m.queryTime = time.Now()
}

func (m *movieList) getMovie() string {
	if time.Since(m.queryTime) > 20*time.Millisecond {
		if m.currentMovie < len(m.movies)-1 {
			m.currentMovie += 1
		} else {
			m.currentMovie = 0
		}
	}
	m.queryTime = time.Now()
	return m.movies[m.currentMovie]
}

func (m *shotsSet) init(dir string) {
	m.shot = make([]shotList, 4)
	movies := getMoviesList(dir)
	for _, movie := range movies {
		_, name := path.Split(movie)
		parts := strings.Split(name, "_")
		shotNum, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Printf("Error parsing filename: %s\n", movie)
			continue
		}
		if shotNum <= len(m.shot) {
			switch {
			case parts[2] == "start.mp4":
				m.shot[shotNum-1].startMovies = append(m.shot[shotNum-1].startMovies, movie)
			case parts[2] == "end.mp4":
				m.shot[shotNum-1].endMovies = append(m.shot[shotNum-1].endMovies, movie)
			}
		}
	}
}

func (m *shotsSet) getMovie(shot int, movie string) string {
	if shot >= len(m.shot) {
		fmt.Printf("There is no this shot number: %d\n", shot)
		return ""
	}
	var list []string
	switch movie {
	case "start":
		list = m.shot[shot-1].startMovies
	case "end":
		list = m.shot[shot-1].endMovies
	default:
		fmt.Printf("Error: You can request only 'start' or 'end' movie.\n")
		return ""
	}
	if time.Since(m.queryTime) > 20*time.Millisecond {
		if m.shot[shot].currentMovie < len(list)-1 {
			m.shot[shot-1].currentMovie += 1
		} else {
			m.shot[shot-1].currentMovie = 0
		}
	}
	m.queryTime = time.Now()
	return list[m.shot[shot-1].currentMovie]
}

func getMoviesList(dir string) (movies []string) {
	f, err := os.Open(dir)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return nil
	}
	defer f.Close()
	files, err := f.Readdir(0)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return nil
	}
	for _, file := range files {
		if path.Ext(file.Name()) == ".mp4" {
			movies = append(movies, path.Join(dir, file.Name()))
		}
	}
	return movies
}

func VideoMovieServer(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("New movie request\n")
	var movie string
	vars := mux.Vars(r)
	switch vars["name"] {
	case "intro":
		movie = mov.introMovies.getMovie()
	case "printing":
		movie = mov.printingMovies.getMovie()
	case "counter":
		movie = mov.counterMovies.getMovie()
	default:
		fmt.Printf("Error: You can not request this movie: %s.\n", vars["name"])
		return
	}
	if b, err := ioutil.ReadFile(movie); err != nil {
		fmt.Printf("Error reading file: %s\n", err)
	} else {
		w.Write(b)
	}
}

func VideoShotServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shot, err := strconv.Atoi(vars["shot"])
	if err != nil {
		return
	}
	movie := vars["movie"]

	if b, err := ioutil.ReadFile(mov.shotMovies.getMovie(shot, movie)); err != nil {
		fmt.Printf("Error reading file: %s\n", err)
	} else {
		w.Write(b)
	}
}

