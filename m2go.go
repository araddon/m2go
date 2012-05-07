package m2go

import (
	"bytes"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var er error
var ReconnectOnZmq bool

/*
type m2Conn struct {
	headers      map[string][]string
	wroteHeaders bool
	out          bytes.Buffer
}

func (conn *m2Conn) Write(data []byte) (n int, err error) {
	var buf bytes.Buffer

	if !conn.wroteHeaders {
		conn.wroteHeaders = true
		for k, v := range conn.headers {
			for _, i := range v {
				buf.WriteString(k + ": " + i + "\r\n")
			}
		}

		buf.WriteString("\r\n")
		conn.out.Write(buf.Bytes())
	}
	return conn.out.Write(data)
}

func (conn *m2Conn) StartResponse(status int) {
	var buf bytes.Buffer
	text := web.StatusText[status]

	fmt.Fprintf(&buf, "HTTP/1.1 %d %s\r\n", status, text)
	conn.out.Write(buf.Bytes())
}

func (conn *m2Conn) SetHeader(hdr string, val string, unique bool) {
	if _, contains := conn.headers[hdr]; !contains {
		conn.headers[hdr] = []string{val}
		return
	}

	if unique {
		//just overwrite the first value
		conn.headers[hdr][0] = val
	} else {
		newHeaders := make([]string, len(conn.headers)+1)
		copy(newHeaders, conn.headers[hdr])
		newHeaders[len(newHeaders)-1] = val
		conn.headers[hdr] = newHeaders
	}
}

func (conn *m2Conn) Close() {

}
*/
// implementation of http.ResponseWriter interface
type M2Writer struct {
	req          *http.Request
	m2req        *M2Request
	wroteHeaders bool
	out          bytes.Buffer
	hdr          bytes.Buffer
	Stream       func([]byte)
	resp         *http.Response
	chunked      bool
}

func (w *M2Writer) Flush() {

	if !w.wroteHeaders {
		w.WriteHeaderResponse(nil)

		msg := fmt.Sprintf("%s %d:%s, %s%s", w.m2req.uuid, len(w.m2req.requestid), w.m2req.requestid,
			string(w.hdr.Bytes()), string(w.out.Bytes()))

		if len(msg) > 200 {
			log.Print("--------------------RESPONSE---------------------\n" + msg[:200] + "\n")
		} else {
			log.Print("--------------------RESPONSE---------------------\n" + msg + "\n")
		}

		w.Stream([]byte(msg))

		w.out = bytes.Buffer{}
		w.hdr = bytes.Buffer{}
	} else {
		// for streaming! not the \r\n terminating
		msg := fmt.Sprintf("%s %d:%s, %s\r\n", w.m2req.uuid, len(w.m2req.requestid), w.m2req.requestid, string(w.out.Bytes()))
		w.Stream([]byte(msg))
		w.out = bytes.Buffer{}
	}
}
func (w *M2Writer) Write(data []byte) (n int, err error) {
	w.out.Write(data)
	if w.wroteHeaders {
		w.Flush()
	}
	return len(data), nil
}

func (w *M2Writer) Header() http.Header {
	return w.resp.Header
}
func (w *M2Writer) WriteHeaderResponse(data []byte) {

	w.wroteHeaders = true
	n := w.out.Len()
	w.resp.Header["Content-Length"] = []string{strconv.FormatInt(int64(n), 10)}

	// RequestMethod should be upper-case
	if w.req != nil {
		w.req.Method = strings.ToUpper(w.req.Method)
	}

	// Status line
	statusText, ok := StatusText[w.resp.StatusCode]
	if !ok {
		statusText = "status code " + strconv.Itoa(w.resp.StatusCode)
	}
	w.hdr.WriteString("HTTP/" + strconv.Itoa(w.resp.ProtoMajor) + ".")
	w.hdr.WriteString(strconv.Itoa(w.resp.ProtoMinor) + " ")
	w.hdr.WriteString(strconv.Itoa(w.resp.StatusCode) + " " + statusText + "\r\n")

	te := w.resp.Header.Get("Transfer-Encoding")
	if te == "chunked" {
		//resp.Header.Del("Content-Length")
		w.chunked = true
		delete(w.resp.Header, "Content-Length")
	}

	// if we have set a new cookie, make sure we add in others that haven't changed
	/*
		if s := w.resp.Header.Get("Set-Cookie"); len(s) > 0 {
			for _, ckie := range w.req.Cookies() {
				if strings.Index(s, ckie.Name+"=") == -1 {
					//w.resp.Header.Add("Set-Cookie", ckie.String())
				}
			}
		}
	*/

	for k, v := range w.resp.Header {
		// this isn't right
		for _, i := range v {
			w.hdr.WriteString(k + ": " + i + "\r\n")
		}
	}

	// End-of-header
	w.hdr.WriteString("\r\n")

}
func (w *M2Writer) WriteHeader(code int) {
	w.resp.StatusCode = code
}

func HandleM2Request(datain []byte, response func(b []byte), handler http.Handler) {

	m2datas := string(datain)
	if strings.Contains(m2datas, `{"METHOD":"JSON"},21:{"type":"disconnect"}`) {
		//should we return something?  notify handler?
		return
	}

	m2req, tnerr := M2Parse(m2datas)
	if tnerr != nil && (m2req != nil && len(m2req.uuid) != 0) {
		log.Print("ERROR", tnerr.Error())
		errmsg := ", HTTP/1.1 500 Internal Server Error\nContent-Type: text/html; charset=utf-8\nContent-Length: 0"
		errmsg = fmt.Sprintf("%s %d:%s, %s", m2req.uuid, len(m2req.requestid), m2req.requestid, errmsg)
		response([]byte(errmsg))
		return
	} else if tnerr != nil {
		log.Print("ERROR", tnerr.Error())
		return
	}

	req, err := ReadHttpRequest(m2req)
	//log.Println(req)

	if err != nil {
		// TODO: tap into web.go 500 mechanism
		log.Print("error", err.Error())
		return
	}
	w := &M2Writer{req, m2req, false, bytes.Buffer{}, bytes.Buffer{}, response, NewResponse(req), false}

	//m.ServeHTTP(w, req)
	handler.ServeHTTP(w, req)

	w.Flush()

}

// the  pattern muxer for M2go, is a light adapter 
// over pat
type M2Mux struct {
	*pat.PatternServeMux
}

func (m *M2Mux) Disconnect(pat string, h http.Handler) {
	m.Add("POST", pat, h)
}
func New() *M2Mux {
	return &M2Mux{pat.New()}
}

// the listen and server for mongrel, expects an address like this
// @addr = string config parameter like this:   
//    m2go.ListenAndServe("tcp://127.0.0.1:9555|tcp://127.0.0.1:9556|54c6755b-9628-40a4-9a2d-cc82a816345e", handler)
func ListenAndServe(addr string, handler http.Handler) {
	var Context zmq.Context
	var SocketIn zmq.Socket
	var SocketOut zmq.Socket
	var hasExited bool
	var err error

	m2addr := strings.Split(addr, "|") //  

	log.Printf("m2go serving  %s\n", addr)

	/*
	  Connection to ZMQ setup
	*/
	connect := func() {
		if Context, err = zmq.NewContext(); err != nil {
			panic("No ZMQ Context?")
		}

		// listen for incoming requests
		if SocketIn, err = Context.NewSocket(zmq.PULL); err != nil {
			panic("No ZMQ Socket?")
		}
		SocketIn.Connect(m2addr[0])

		if SocketOut, err = Context.NewSocket(zmq.PUB); err != nil {
			panic("No ZMQ Socket Outbound??")
		}
		// outbound response on a different channel
		SocketOut.SetSockOptString(zmq.IDENTITY, m2addr[2])
		//socket.SetSockOptString(zmq.SUBSCRIBE, filter)
		SocketOut.Connect(m2addr[1])
	}

	connect()

	handleResponse := func(response []byte) {
		SocketOut.Send(response, 0)
	}
	stopper := func() {
		if !hasExited {
			hasExited = true
			SocketOut.Close()
			SocketIn.Close()
			Context.Close()
		}
	}
	defer stopper()

	for {
		// each inbound request
		m2data, err := SocketIn.Recv(0)
		//log.Println(string(m2data))
		if err != nil {
			log.Println("ZMQ Socket Input accept error ", err.Error())
		} else {
			go HandleM2Request(m2data, handleResponse, handler)
		}
	}
	log.Print("after close of runner")
}
