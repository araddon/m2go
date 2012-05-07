package m2go

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var StatusText = map[int]string{
	http.StatusContinue:           "Continue",
	http.StatusSwitchingProtocols: "Switching Protocols",

	http.StatusOK:                   "OK",
	http.StatusCreated:              "Created",
	http.StatusAccepted:             "Accepted",
	http.StatusNonAuthoritativeInfo: "Non-Authoritative Information",
	http.StatusNoContent:            "No Content",
	http.StatusResetContent:         "Reset Content",
	http.StatusPartialContent:       "Partial Content",

	http.StatusMultipleChoices:   "Multiple Choices",
	http.StatusMovedPermanently:  "Moved Permanently",
	http.StatusFound:             "Found",
	http.StatusSeeOther:          "See Other",
	http.StatusNotModified:       "Not Modified",
	http.StatusUseProxy:          "Use Proxy",
	http.StatusTemporaryRedirect: "Temporary Redirect",

	http.StatusBadRequest:                   "Bad Request",
	http.StatusUnauthorized:                 "Unauthorized",
	http.StatusPaymentRequired:              "Payment Required",
	http.StatusForbidden:                    "Forbidden",
	http.StatusNotFound:                     "Not Found",
	http.StatusMethodNotAllowed:             "Method Not Allowed",
	http.StatusNotAcceptable:                "Not Acceptable",
	http.StatusProxyAuthRequired:            "Proxy Authentication Required",
	http.StatusRequestTimeout:               "Request Timeout",
	http.StatusConflict:                     "Conflict",
	http.StatusGone:                         "Gone",
	http.StatusLengthRequired:               "Length Required",
	http.StatusPreconditionFailed:           "Precondition Failed",
	http.StatusRequestEntityTooLarge:        "Request Entity Too Large",
	http.StatusRequestURITooLong:            "Request URI Too Long",
	http.StatusUnsupportedMediaType:         "Unsupported Media Type",
	http.StatusRequestedRangeNotSatisfiable: "Requested Range Not Satisfiable",
	http.StatusExpectationFailed:            "Expectation Failed",

	http.StatusInternalServerError:     "Internal Server Error",
	http.StatusNotImplemented:          "Not Implemented",
	http.StatusBadGateway:              "Bad Gateway",
	http.StatusServiceUnavailable:      "Service Unavailable",
	http.StatusGatewayTimeout:          "Gateway Timeout",
	http.StatusHTTPVersionNotSupported: "HTTP Version Not Supported",
}

/*  
  M2 Request and server handling
*/

type M2Request struct {
	uuid      string
	requestid string
	path      string
	rest      string
	body      string
	headers   string
	cid       uint //customer id
}

type Response struct {
	out    string // response text
	status string // OK etc
	code   string //  404, 500, etc
}

// new response
func NewResponse(req *http.Request) (resp *http.Response) {

	resp = new(http.Response)
	resp.Header = make(http.Header)
	resp.Request = req
	resp.Request.Method = strings.ToUpper(req.Method)

	resp.StatusCode = 200

	resp.Proto = req.Proto
	resp.ProtoMajor = 1
	resp.ProtoMinor = 1

	if ct, ok := req.Header["Content-Type"]; ok {
		resp.Header["Content-Type"] = ct
	} else {
		resp.Header["Content-Type"] = []string{"text/html; charset=utf-8"}
	}

	if clenStr := resp.Header.Get("Content-Length"); clenStr != "" {
		var err error
		if err != nil {
			log.Printf("http: invalid Content-Length of %q sent", clenStr)
			delete(resp.Header, "Content-Length")
		}
	}
	if te, ok := req.Header["Transfer-Encoding"]; ok {
		resp.Header["Transfer-Encoding"] = te
	}
	if _, ok := resp.Header["Date"]; !ok {
		resp.Header["Date"] = []string{time.Now().UTC().Format(http.TimeFormat)}
	}

	return resp
}

// ReadRequest creates http request from m2request
func ReadHttpRequest(m2req *M2Request) (req *http.Request, err error) {

	req = new(http.Request)
	headers := make(http.Header)

	var f map[string]interface{}

	err = json.Unmarshal([]byte(m2req.headers), &f)
	//log.Print("Headers ", m2req.headers, " \n\n")
	if err != nil {
		log.Printf("arg, error %s", err)
	}
	for k, v := range f {
		switch v.(type) {
		case []interface{}:
			// "accept\":[\"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\",\"*/*\"],
			log.Println(v)
			for _, v2 := range v.([]interface{}) {
				switch v2.(type) {
				case string:
					headers.Set(k, v2.(string))
				default:
					log.Println("not string? ", k, v2)
				}
			}
		case string:
			headers.Set(k, v.(string))
		}
	}
	body := strings.NewReader(m2req.body)
	//body := bytes.NewBuffer([]byte(m2req.body))
	//log.Print(m2req.body)

	req = NewHttpRequestFromM2(headers, body)

	req.Header.Set("m2body", m2req.body)
	req.Header.Set("m2header", m2req.headers)

	return
}

func NewHttpRequestFromM2(headers http.Header, body io.Reader) *http.Request {

	host := headers.Get("host")
	method := strings.ToUpper(headers.Get("method"))
	path := headers.Get("uri")
	proto := headers.Get("version")
	rawurl := "http://" + host + path
	url_, _ := url.Parse(rawurl)
	//useragent := headers.Get("user-agent")
	remoteAddr := headers.Get("x-forwarded-for")
	//remotePort, _ := strconv.Atoi(headers.Get("REMOTE_PORT"))

	if method == "POST" {
		if ctype, ok := headers["CONTENT_TYPE"]; ok {
			headers["Content-Type"] = ctype
		}

		if clength, ok := headers["CONTENT_LENGTH"]; ok {
			headers["Content-Length"] = clength
		}
	}
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	return &http.Request{
		Method:     method,
		URL:        url_,
		Proto:      proto,
		Host:       host,
		Body:       rc,
		Header:     headers,
		RemoteAddr: remoteAddr,
	}

}

// Parse an incoming request of format/type mongrel2
// and return an M2Request object
func M2Parse(reqs string) (r *M2Request, err error) {

	r = new(M2Request)
	log.Printf("raw\n\n %s \n\n", strconv.Quote(string(reqs)))

	parts := strings.SplitN(string(reqs), " ", 4)
	r.uuid = parts[0]
	r.requestid = parts[1]
	r.path = parts[2]
	r.rest = strings.ToLower(parts[3])

	tn, err := NewTnet(r.rest)
	if err != nil {
		err = errors.New("Error parsing request, no header? " + err.Error())
		return
	}
	if tn.Value != nil {
		r.headers = tn.Value.(string)
	}

	tnetb, err := tn.Next()
	if err != nil {
		err = errors.New("Error parsing request on body " + err.Error())
		return
	}
	if tn.Value != nil {
		r.body = tnetb.Value.(string)
	}

	return
}
