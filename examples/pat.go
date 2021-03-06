package main

import (
	"github.com/araddon/m2go"
	"github.com/bmizerany/pat"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	m := pat.New()
	m.Get("/hello/:name", http.HandlerFunc(hello))
	m.Get("/cookie/:name/:value", http.HandlerFunc(cookie))
	m.Get("/favicon.ico", http.HandlerFunc(empty))
	m.Get("/stream", http.HandlerFunc(stream))
	m2go.ListenAndServe("tcp://127.0.0.1:9055|tcp://127.0.0.1:9056|d9eae9a0-6bad-11e1-9cc3-5254004a61b5", m)
}

func hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	io.WriteString(w, "Hello, "+name)
}

func cookie(w http.ResponseWriter, r *http.Request) {
	c := http.Cookie{Name: r.URL.Query().Get(":name"), Value: r.URL.Query().Get(":value"), Path: "/"}
	http.SetCookie(w, &c)
	log.Println(len(r.Cookies()))
	//w.Header().Add("Set-Cookie", cookie.String())
	for _, ckie := range r.Cookies() {
		io.WriteString(w, "Cookie:  "+ckie.String()+"<br />")
	}
	if len(r.Cookies()) == 0 {
		io.WriteString(w, "No cookies")
	}
}

func empty(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "empty ")
}

func stream(w http.ResponseWriter, r *http.Request) {
	// Path variable names are in the URL.Query() and start with ':'.
	log.Println("about to set transfer encoding")
	r.Header.Set("Transfer-Encoding", "chunked")
	r.Header.Set("Content-Type", "application/json")
	io.WriteString(w, "morestuff")
	log.Println("called write in stream")
	// lets set a timer to create fake data
	timer := time.NewTicker(time.Second * 1)

	go func() {
		for _ = range timer.C {
			io.WriteString(w, `{"msg":"still alive","status":200}`)
		}
	}()
}
