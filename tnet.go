/*
Tnet strings, see http://tnetstrings.org/

example      SIZE COLON VALUE TYPE
   2:42#     = len   :   value   type
   5:hello,  = len   :   value   type

   types:   
    ,     =  string (byte array)
    #     =  integer
    ^     =  float
    !     =  boolean of 'true' or 'false'
    ~     =  null always encoded as 0:~
    }     =  Dictionary which you recurse into to fill with key=value pairs inside the payload contents.
    ]     =  List which you recurse into to fill with values of any type.

*/
package m2go

import (
  "errors"
  "fmt"
  "strconv"
  "strings"
  //"reflect"
)


type Tnet struct {
  Raw string
  Payload string
  Datatype string
  Extra string
  Length int
  Value interface {}
}

// grabs the next value from the leftovers:  "extra"
func (t *Tnet) Next() (tnet Tnet, err error){
  return NewTnet(t.Extra)
}

// Parse string and return value, remaining value
func NewTnet(datain string) (tn Tnet, err error){
  if len(datain) < 1 {
    return 
  }

  tn = Tnet{Raw:datain}

  err = tn.parse()
  if tn.Length > 0 && err == nil {
    switch tn.Datatype {
      case "#":  
        tn.Value, err = strconv.Atoi64(tn.Payload)
      case  "}":
        err = parseDict(&tn)
      case  "]":
        err = parseList(&tn)
      case ",": 
        tn.Value = tn.Payload
      case "~": 
        tn.Value = nil
      case "!": 
        if tn.Payload == "true" {
          tn.Value = true
        } else if tn.Payload == "false" {
          tn.Value = false
        } else {
          err = errors.New("Unexpected true/false value " + tn.Payload)
        }
      case "^": 
        tn.Value, err = strconv.Atof64(tn.Payload)
      default:
        err = errors.New("Tnet unknown datatype = " + tn.Datatype)
    }
  }

  //fmt.Printf("tnetstrings 1 = %s\n 2=%s\n3=\n%s", payload, tnet.datatype, value)
  return 
}

//  tnetParse:   parses a single value by tnet rules (above)
func (tn *Tnet) parse() (err error){
  
  parts := strings.SplitN(tn.Raw,":",2)
  if len(parts) != 2 || len(parts[0]) < 1 {
    return errors.New("Invalid, didn't contain len:value+")
  }

  if tn.Length, er = strconv.Atoi(parts[0]); er != nil {
    err = errors.New("Error getting length part " + er.Error())
  }
  if len(parts[1]) < tn.Length + 1 {
    return errors.New("Invalid length, ")
  }
  
  tn.Payload = parts[1][:tn.Length]
  tn.Datatype = parts[1][tn.Length:tn.Length + 1]

  // validations
  if tn.Length != len(tn.Payload) {
    err = errors.New(fmt.Sprintf("TNet error, invalid length, was %d but expected %d", len(tn.Payload), tn.Length))
  }
  if len(tn.Raw) > tn.Length + 1 {
    tn.Extra = parts[1][tn.Length + 1:]
  }
  
  return
}
func parseList(tparent *Tnet) (err error){
  if len(tparent.Payload) == 0 {
    return 
  }

  var data string

  var tnetlist = make([]Tnet,0)
  data = tparent.Payload

  // tparent = 11:3:bob,2:55#]
  // data = tparent.Payload = 3:bob,2:55#
  // l1 = 3:bob,   extra = 2:55#
  // l2 = 2:55#    extra = nil
  for {
    tnl, er := NewTnet(data)
    if er != nil {
      err = er
      tparent.Value = tnetlist
      return
    }

    tnetlist = append(tnetlist,tnl)

    if len(tnl.Extra) == 0 {
      tparent.Value = tnetlist
      return 
    }

    data = tnl.Extra
  }
  
  return 
}
func parseDict(tparent *Tnet) (err error){
  
  if len(tparent.Payload) == 0 {
    return 
  }

  var data string
  var nvmap = make(map[string]Tnet)

  tparent.Value = nvmap
  data = tparent.Payload
  // tparent = 26:4:name,3:bob,3:age,2:55#}
  // data = tparent.Payload = 4:name,3:bob,3:age,2:55#
  // tn = 4:name,   extra = 3:bob,3:age,2:55#
  // tv = 3:bob,    extra = 3:age,2:55#
  // tn2 = 3:age,   extra = 2:55#
  // tv2 = 2:55     extra = nil
  for {
    tn, er := NewTnet(data)
    if er != nil {
      err = er
      return
    }
    if len(tn.Extra) < 1 {
      return errors.New("unbalanced dictionary name/value pair (missing value) ")
    }
    tv, er2 := NewTnet(tn.Extra)
    if er2 != nil {
      err = er2
      return
    }

    nvmap[tn.Payload] = tv 

    if len(tv.Extra) == 0 {
      return 
    }

    data = tv.Extra
    
  }
  
  return
}
