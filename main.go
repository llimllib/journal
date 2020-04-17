package main

// TODO: handle panics
// TODO: handle permalinks
// TODO: paging
// TODO: handle tags
//       * honestly not sure I have enough to even care
// TODO: reaosnable CSS

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

func logreq(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("|%s", r.URL.Path)

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
	ID           string
	Dt           string
	Typ          string
	QuoteBody    string
	QuoteSource  string
	PhotoCaption string
	PhotoLink    string
	PhotoURLs    []string
	TextTitle    string
	TextBody     string
	LinkURL      string
	LinkText     string
	LinkDesc     string
	VideoSource  string
	VideoCaption string
	AudioPlayer  string
	AudioCaption string
}

var szre = regexp.MustCompile(`_(\d+)[\w]*\.`)

func mustItoa(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

// BiggestImage returns the biggest image of the PhotoURLs array.
// From a brief scan of my posts, I guess you can only post one image per
// tumblr post
func (p *Post) BiggestImage() string {
	max := 0
	biggestImage := ""
	for _, im := range p.PhotoURLs {
		res := szre.FindSubmatch([]byte(im))
		sz := mustItoa(string(res[1]))
		if sz > max {
			max = sz
			biggestImage = im
		}
	}
	return biggestImage
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

func (s *JournalServer) handle(w http.ResponseWriter, r *http.Request) {
	rows, err := s.database.Query(`
		SELECT p.id, p.date, p.type, qp.body quote_body, qp.source
		  quote_source, pp.caption photo_caption, pp.link photo_link, tp.title
		  text_title, tp.body text_body, lp.url link_url, lp.text link_text,
		  lp.desc link_desc, vp.source video_source, vp.caption video_caption,
		  ap.player audio_player, ap.caption audio_caption, group_concat(pu.url)
		FROM posts p
		LEFT JOIN quote_posts qp ON p.id=qp.id
		LEFT JOIN photo_posts pp ON p.id=pp.id
		LEFT JOIN photo_urls pu on pu.id=p.id
		LEFT JOIN text_posts tp ON p.id=tp.id
		LEFT JOIN link_posts lp ON p.id=lp.id
		LEFT JOIN video_posts vp ON p.id=vp.id
		LEFT JOIN audio_posts ap ON p.id=ap.id
		GROUP BY p.id
		ORDER BY p.date desc
		LIMIT 5
	`)
	if err != nil {
		log.Printf("Failed to serve: %s", err.Error())
	}
	defer rows.Close()

	data := struct {
		Posts []Post
	}{}
	for rows.Next() {
		var id, dt, typ, quoteBody, quoteSource, photoCaption, photoLink, photoURLs, textTitle, textBody, linkURL, linkText, linkDesc, videoSource, videoCaption, audioPlayer, audioCaption sql.NullString

		err = rows.Scan(&id, &dt, &typ, &quoteBody, &quoteSource, &photoCaption, &photoLink, &textTitle, &textBody, &linkURL, &linkText, &linkDesc, &videoSource, &videoCaption, &audioPlayer, &audioCaption, &photoURLs)
		if err != nil {
			panic(err)
		}

		post := Post{
			ID:           id.String,
			Dt:           dt.String,
			Typ:          typ.String,
			QuoteBody:    quoteBody.String,
			QuoteSource:  quoteSource.String,
			PhotoCaption: photoCaption.String,
			PhotoLink:    photoLink.String,
			PhotoURLs:    strings.Split(photoURLs.String, ","),
			TextTitle:    textTitle.String,
			TextBody:     textBody.String,
			LinkURL:      linkURL.String,
			LinkText:     linkText.String,
			LinkDesc:     linkDesc.String,
			VideoSource:  videoSource.String,
			VideoCaption: videoCaption.String,
			AudioPlayer:  audioPlayer.String,
			AudioCaption: audioCaption.String,
		}
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
