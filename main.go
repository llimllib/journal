package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

func logreq(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s", r.URL.Path)

		f(w, r)
	})
}

func readTemplate(name string) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func mustTemplate(name string) *template.Template {
	f, err := readTemplate(name)
	if err != nil {
		panic(err)
	}

	t, err := template.New("index").Parse(f)
	if err != nil {
		panic(err)
	}

	return t
}

var (
	homeTemplate = mustTemplate("index.html")
)

type Post struct {
	ID   string
	Dt   string
	Typ  string
	HTML string
}

type JournalServer struct {
	host     string
	port     string
	database *sql.DB
}

func (s *JournalServer) start() {
	defer s.database.Close()

	http.Handle("/", logreq(s.handle))
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	log.Printf("Serving on %s:%s", s.host, s.port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (s *JournalServer) loadQuote(id string) string {
	stmt, err := s.database.Prepare("SELECT body, source FROM quote_posts WHERE id=?")
	if err != nil {
		log.Printf("Failed to serve: %s", err.Error())
		return ""
	}
	defer stmt.Close()

	var body, source string
	err = stmt.QueryRow(id).Scan(&body, &source)
	if err != nil {
		log.Printf("Failed to query: %s", err.Error())
		return ""
	}

	return "<blockquote>" + body + "</blockquote>\n" + source

}

func (s *JournalServer) loadPost(id, typ string) string {
	switch typ {
	case "quote":
		return s.loadQuote(id)
	case "default":
		return ""
	}

	return ""
}

func (s *JournalServer) handle(w http.ResponseWriter, r *http.Request) {
	// TODO: handle permalinks
	rows, err := s.database.Query("SELECT id, date, type FROM posts LIMIT 5")
	if err != nil {
		log.Printf("Failed to serve: %s", err.Error())
	}
	defer rows.Close()

	data := struct {
		Posts []Post
	}{}
	for rows.Next() {
		var id string
		var dt string
		var typ string

		err = rows.Scan(&id, &dt, &typ)
		if err != nil {
			panic(err)
		}

		post := Post{
			ID:  id,
			Dt:  dt,
			Typ: typ,
		}
		post.HTML = s.loadPost(id, typ)
		data.Posts = append(data.Posts, post)
	}

	err = homeTemplate.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func main() {
	var host string
	var port string
	var database string

	flag.StringVar(&host, "host", "0.0.0.0", "host for journal to listen on")
	flag.StringVar(&port, "port", "11111", "port for journal to listen on")
	flag.StringVar(&database, "database", "posts.sqlite3", "database to serve from")
	flag.Parse()

	db, err := sql.Open("sqlite3", database)
	if err != nil {
		panic(err)
	}
	server := JournalServer{host, port, db}
	server.start()
}
