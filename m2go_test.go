package m2go_test

import (
  //"m2go"
  "web"
  "testing"
)
type M2Data struct {
  inbound string
  uuid string
  path string
}
var testData  = [...]M2Data{M2Data{"54c6755b-9628-40a4-9a2d-cc82a816345e 143 /c/18597 652:{\"PATH\":\"/c/18597\",\"x-forwarded-for\":\"127.0.0.1\",\"cache-control\":\"max-age=0\",\"origin\":\"http://localhost:6767\",\"content-type\":\"application/x-www-form-urlencoded\",\"accept-language\":\"en-US,en;q=0.8\",\"accept-encoding\":\"gzip,deflate,sdch\",\"connection\":\"keep-alive\",\"content-length\":\"58\",\"accept-charset\":\"ISO-8859-1,utf-8;q=0.7,*;q=0.3\",\"accept\":\"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\",\"user-agent\":\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_2) AppleWebKit/535.7 (KHTML, like Gecko) Chrome/16.0.912.41 Safari/535.7\",\"host\":\"localhost:6767\",\"cookie\":\"cookies\",\"METHOD\":\"POST\",\"VERSION\":\"HTTP/1.1\",\"URI\":\"/c/18597\",\"PATTERN\":\"/c/\"},58:source=myid%3D1234%26category%3Dbooks%26ts%3D1322416191407,","54c6755b-9628-40a4-9a2d-cc82a816345e","/c/18597" }}



func hello(ctx *web.Context, val, val2 string) string { 
    for k,v := range ctx.Params {
        println(k, v)
    }
    return "hello " + val + " " + val2
}


func TestMain(t *testing.T) {

  web.Post("/(.*)/(.*)", hello)

  //m2go.Run("tcp://127.0.0.1:9555|tcp://127.0.0.1:9556|54c6755b-9628-40a4-9a2d-cc82a816345e")

}

func TestM2FormatParse(t *testing.T) {
  
  //54c6755b-9628-40a4-9a2d-cc82a816345e 134 /c/18597 652:{"PATH":"/c/18597","x-forwarded-for":"127.0.0.1","cache-control":"max-age=0","origin":"http://localhost:6767","content-type":"application/x-www-form-urlencoded","accept-language":"en-US,en;q=0.8","accept-encoding":"gzip,deflate,sdch","connection":"keep-alive","content-length":"58","accept-charset":"ISO-8859-1,utf-8;q=0.7,*;q=0.3","accept":"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8","user-agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_2) AppleWebKit/535.7 (KHTML, like Gecko) Chrome/16.0.912.41 Safari/535.7","host":"localhost:6767","cookie":"cookies","METHOD":"POST","VERSION":"HTTP/1.1","URI":"/c/18597","PATTERN":"/c/"},58:source=myid%3D1234%26category%3Dbooks%26ts%3D1322415601141, 

}


