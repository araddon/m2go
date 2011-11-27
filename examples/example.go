//#!~/gocode/bin/gorun
package main

import (
  "m2go"
  //"fmt"
  "web"
)

func hello(val string) string { 
  return "hello " + val 
}

func main() {
  web.Get("/(.*)", hello)
  web.Post("/(.*)", hello)
  
  m2go.Run( "tcp://127.0.0.1:9555|tcp://127.0.0.1:9556|54c6755b-9628-40a4-9a2d-cc82a816345e")
}
