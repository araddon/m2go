m2go, a `Mongrel2 http server <http://mongrel2.org>`_ for go.   

Includes adapter to run `Pat.go <github.com/bmizerany/pat>`_ 


Usage
===================

Example App::

	import (
		"github.com/araddon/m2go"
		"github.com/bmizerany/pat"
		"io"
		"log"
		"net/http"
	)

	func main() {
		log.SetFlags(log.Ltime | log.Lshortfile)
		m := pat.New()
		m.Get("/hello/:name", http.HandlerFunc(hello))
		m.Get("/stream", http.HandlerFunc(stream))
		m2go.ListenAndServe("tcp://127.0.0.1:9055|tcp://127.0.0.1:9056|d9eae9a0-6bad-11e1-9cc3-5254004a61b5", m)
	}

	func hello(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get(":name")
		io.WriteString(w, "Hello, "+name)
	}

	// this will serve as a Streaming API, continuing to push out updates
	// to connected client
	func stream(w http.ResponseWriter, r *http.Request) {
		// this line:  is the key to make it streaming.
		r.Header.Set("Transfer-Encoding", "chunked")
		r.Header.Set("Content-Type", "application/json")
		io.WriteString(w, "some content")
		
		// lets simulate a zeromq type connection that recieves messages periocically
		//  and pushes to client
		timer := time.NewTicker(time.Second * 1)

		go func() {
			for _ = range timer.C {
				io.WriteString(w, `{"msg":"still alive","status":200}`)
			}
		}()
	}