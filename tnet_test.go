/*
  See more about Tnet strings http://tnetstrings.org/

*/
package m2go_test

import (
  . "m2go"
  "testing"
  "reflect"
  //"fmt"
)
func assert(is bool, msg string) {
  
}

func TestTnet(t *testing.T) {

  
  tnet, err := NewTnet("5:hello,")
  if err != nil || tnet.Length != 5 || tnet.Datatype != "," || tnet.Value.(string) != "hello" {
    t.Errorf("Should be error free and = 'hello' but was: err=%s,  len=%d '%s'",err, tnet.Length, tnet.Value)
  }

  // make the length wrong (too long)
  tnet, err = NewTnet("7:hello,")
  if err == nil {
    t.Errorf("Should have error %s",err)
  }

  tnet, err = NewTnet("2:42#")
  if err != nil || tnet.Datatype != "#" || tnet.Value.(int64) != int64(42) {
    t.Errorf("Should be error free and = 42, but was;  err=%s, v=%v", err, tnet.Value)
  }

  tnet, err = NewTnet("6:42.356^")
  if err != nil || tnet.Datatype != "^" || tnet.Value.(float64) != float64(42.356) {
    t.Errorf("Should be error free and = 42, but was;  err=%s, v=%v", err,  tnet.Value)
  }

  tnet, err = NewTnet("4:true!")
  if err != nil || tnet.Datatype != "!" || tnet.Value.(bool) != true {
    t.Errorf("Should be error free and = true, but was;  err=%s, v=%v", err,  tnet.Value)
  }


  tnet, err = NewTnet("5:false!")
  if err != nil || tnet.Datatype != "!" || tnet.Value.(bool) != false {
    t.Errorf("Should be error free and = false, but was;  err=%s, v=%v", err,  tnet.Value)
  }

  tnet, err = NewTnet("0:~")
  if err != nil || tnet.Datatype != "~" || tnet.Value != nil || tnet.Length != 0{
    t.Errorf("Should be error free and = nil, but was;  err=%s, v=%v", err,  tnet.Value)
  }
  
  // dictionary
  tnet, err = NewTnet("24:4:name,3:bob,3:age,2:55#}")
  rv := reflect.ValueOf(tnet.Value)
  m := tnet.Value.(map[string]Tnet)

  if err != nil || tnet.Datatype != "}" || rv.Kind() != reflect.Map || len(m) != 2 || tnet.Length != 24 { 
    t.Errorf("Should be error free and = map len=2, but was;  err=%s, v=%v", err,  tnet.Value)
  }
  if m["name"].Value.(string) != "bob" || m["age"].Value.(int64) != int64(55) { 
    t.Errorf("should be bob, age =55 but was %v", m)
  }

  // list
  tnet, err = NewTnet("27:6:42.356^3:bob,2:55#4:true!]")
  rv = reflect.ValueOf(tnet.Value)
  
  l := tnet.Value.([]Tnet)

  if err != nil || tnet.Datatype != "]" || rv.Kind() != reflect.Slice || len(l) != 4 || tnet.Length != 27 { 
    t.Errorf("Should be error free and = map len=2, but was;  err=%s, v=%v", err,  tnet.Value)
  }
  if l[0].Value.(float64) != float64(42.356) || l[1].Value.(string) != "bob" || l[2].Value.(int64) != int64(55) { 
    t.Errorf("should be 42.356, bob, age =55, true but was %v", l)
  }

}


