/*

  Mongrel2 Handler, Parser, and Web.go Adapter

*/
package m2go

import (
  "bytes"
  "encoding/json"
  "errors"
  "flag"
  "fmt"
  "io"
  "log"
  "net/http"
  "net/url"
  "os"
  "strings"
  //"strconv"
  "web"
  zmq "github.com/alecthomas/gozmq"
)

var Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

var er error

type Config struct {
  logfile string
  configFile string
  useconsole bool
}

var M2Config = &Config{"log/dev.log","prod.conf", true}


/*  
  M2 Request and server handling
*/

type M2Request struct {
  uuid string
  requestid string 
  path string
  rest string
  body string
  headers string
  cid uint //customer id
}

type Response struct {
  out string // response text
  status string // OK etc
  code string //  404, 500, etc
}

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

func NewHttpRequestFromM2(headers http.Header, body io.Reader) *web.Request {

  host := headers.Get("host")
  method := strings.ToUpper(headers.Get("method"))
  path := headers.Get("uri")
  proto := headers.Get("version")
  rawurl := "http://" + host  + path
  url_, _ := url.Parse(rawurl)
  useragent := headers.Get("user-agent")
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

  //read the cookies
  cookies := web.ReadCookies(headers)

  return &web.Request{
      Method:     method,
      RawURL:     rawurl,
      URL:        url_,
      Proto:      proto,
      Host:       host,
      UserAgent:  useragent,
      Body:       body,
      Headers:    headers,
      RemoteAddr: remoteAddr,
      //RemotePort: remotePort,
      Cookie:     cookies,
  }

}

// Parse an incoming request of format/type mongrel2
// and return an M2Request object
func M2Parse(reqs string) (r *M2Request, err error){
    
  r = new(M2Request)
  //Logger.Printf("raw\n\n %s \n\n", strconv.Quote(string(reqs)))

  parts := strings.SplitN(string(reqs)," ",4)
  r.uuid = parts[0]
  r.requestid = parts[1]
  r.path = parts[2]
  r.rest = strings.ToLower(parts[3])

  tn, err := NewTnet(r.rest)
  if err != nil {
    err = errors.New("Error parsing request " + err.Error())
    return
  }
  r.headers = tn.Value.(string)
  tnetb, err := tn.Next()
  if err != nil {
    err = errors.New("Error parsing request " + err.Error())
    return
  }
  r.body = tnetb.Value.(string)
  
  return
}

func MakeWebGoRequest(m2req *M2Request) (req *web.Request, err error) {
  headers := make(http.Header)

  var f map[string]interface{}

  err = json.Unmarshal([]byte(m2req.headers), &f)
  
  if err != nil {
    Logger.Printf("arg, error %s", err)
  }
  for k, v := range f {
    headers.Set(k, v.(string))
  }
  body := bytes.NewBuffer([]byte(m2req.body))
  
  req = NewHttpRequestFromM2(headers, body)

  return req, nil
}

func HandleM2Request(s *web.Server, datain []byte, response func(response string)()) {
  
  m2req, tnerr := M2Parse(string(datain))
  if tnerr != nil {
    // TODO: tap into web.go 500 mechanism
    Logger.Println("ERROR", tnerr.Error())
  }

  wreq, err := MakeWebGoRequest(m2req)
  if err != nil {
    // TODO: tap into web.go 500 mechanism
    Logger.Println("error", err.Error())
  }

  conn := m2Conn{ make(map[string][]string), false, bytes.Buffer{}}
  s.RouteHandler(wreq, &conn)
  
  msg := fmt.Sprintf("%s %d:%s, %s", m2req.uuid, len(m2req.requestid), m2req.requestid, string(conn.out.Bytes()))
  Logger.Print("--------------------RESPONSE---------------------\n" + msg + "\n")
  response(msg)
}

// default package init function
func init() {

  RegisterSignalHandler(os.SIGINT, func() { 
    fmt.Print("in signal handler SIGINT")
    Exit(0) 
  })
  RegisterSignalHandler(os.SIGTERM, func() { 
    fmt.Print("in signal handler SIGTERM")
    Exit(0) 
  })
  RegisterSignalHandler(os.SIGUSR1, func() { 
    fmt.Print("in signal handler USR1")
  })
  
  go handleSignals()

  loadConfig()
  
}


// server Method to initialize with config
func loadConfig(){
  
  flag.StringVar(&M2Config.configFile, "config", "none", "Config File to use")
  flag.BoolVar(&M2Config.useconsole, "useconsole", true, "log to console?")
  flag.Parse()

  if M2Config.configFile != "none" {
    //c, _ := config.ReadDefault(M2Config.configFile)
    //  read log file etc

  } else {

    if M2Config.useconsole == true {
      fmt.Println("Logging to console")
    }
  }
  
}

//Runs the web application and serves scgi requests
func Run(addr string) {
  web.Runner(addr, M2Runner)
}

// method for server that Runs the web application, sets up m2 connections
// and serves http requests
// @addr = string config parameter like this:   
//    "tcp://127.0.0.1:9555|tcp://127.0.0.1:9556|54c6755b-9628-40a4-9a2d-cc82a816345e"
func M2Runner(s *web.Server, addr string) {

  var Context zmq.Context
  var SocketIn zmq.Socket
  var SocketOut zmq.Socket

  m2addr := strings.Split(addr,"|")//  
  
  // turn off static web serving from web.go, it is handled by mongrel2
  s.Config.StaticDir = "NONE"

  if s.Logger == nil {
    s.Logger = Logger
  }
  s.Logger.Printf("web.go serving m2 %s\n", addr)
  
  /*
  Connection to ZMQ setup
  */
  var err error
  if Context, err = zmq.NewContext(); err != nil {
    panic("No ZMQ Context?")
  }
  defer Context.Close()

  // listen for incoming requests
  if SocketIn, err = Context.NewSocket(zmq.PULL); err != nil {
    panic("No ZMQ Socket?")
  }
  defer SocketIn.Close()
  SocketIn.Connect(m2addr[0])


  if SocketOut, err = Context.NewSocket(zmq.PUB); err != nil {
    panic("No ZMQ Socket Outbound??")
  }
  // outbound response on a different channel
  SocketOut.SetSockOptString(zmq.IDENTITY, m2addr[2])
  //socket.SetSockOptString(zmq.SUBSCRIBE, filter)
  defer SocketOut.Close()
  SocketOut.Connect(m2addr[1])
  
  handleResponse := func(response string) {
    SocketOut.Send([]byte(response), 0)
  }
  
  for {
    // each inbound request
    datapt, err := SocketIn.Recv(0)
    if err != nil {
      Logger.Println("ZMQ Socket Input accept error", err.Error())
    }
    go HandleM2Request(s,datapt,handleResponse)
  }
    
}
