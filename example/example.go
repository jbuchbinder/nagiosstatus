package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kiyor/gourl/lib"
	"github.com/kiyor/nagiosToJson"
	"io/ioutil"
	"log"
	"strconv"
)

var (
	statusf *string = flag.String("f", "/usr/local/nagios/var/status.dat", "status file")
	all     *bool   = flag.Bool("all", false, "get all info")
	mute    *bool   = flag.Bool("mute", false, "enable show mute info")
	ack     *bool   = flag.Bool("ack", false, " enable show ack")
	url     *string = flag.String("url", "", "get status file by url")
)

func init() {
	flag.Parse()
	if *url != "" {
		setStatByUrl(*url)
	} else {
		nagiosToJson.SetStatFile(*statusf)
	}
}

func setStatByUrl(url string) {
	var req gourl.Req
	req.Url = url
	res, err := req.GetString()
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = ioutil.WriteFile("/tmp/temp.dat", []byte(res), 0644)
	if err != nil {
		log.Fatalln(err.Error())
	}
	nagiosToJson.SetStatFile("/tmp/temp.dat")
}

func main() {
	var stat nagiosToJson.Mainstat
	a := nagiosToJson.GetStat()
	json.Unmarshal(a, &stat)
	if *all {
		j, _ := json.MarshalIndent(stat, "", "    ")
		fmt.Println(string(j))
	} else {
		for hostname, v := range stat.Hoststatus {
			if state(v) != 0 {
				fmt.Println(hostname, v.Plugin_output)
			}
			for servicename, v2 := range v.Servicestatus {
				if state(v2) != 0 {
					fmt.Println(hostname, servicename, v2.Plugin_output)
				}
			}
		}
	}
}

func state(v interface{}) int {
	var res int
	switch v := v.(type) {
	default:
		return 0
	case *nagiosToJson.Hoststatus:
		res, _ = strconv.Atoi(v.Current_state)
		return res
	case *nagiosToJson.Servicestatus:
		res, _ = strconv.Atoi(v.Current_state)
		return res
	}
}
func notifications(v interface{}) bool {
	var res int
	switch v := v.(type) {
	default:
		return false
	case *nagiosToJson.Hoststatus:
		res, _ = strconv.Atoi(v.Notifications_enabled)
		if res == 1 {
			return true
		}
	case *nagiosToJson.Servicestatus:
		res, _ = strconv.Atoi(v.Notifications_enabled)
		if res == 1 {
			return true
		}
	}
	return false
}
func acknowledged(v interface{}) bool {
	var res int
	switch v := v.(type) {
	default:
		return false
	case *nagiosToJson.Hoststatus:
		res, _ = strconv.Atoi(v.Problem_has_been_acknowledged)
		if res == 1 {
			return true
		}
	case *nagiosToJson.Servicestatus:
		res, _ = strconv.Atoi(v.Problem_has_been_acknowledged)
		if res == 1 {
			return true
		}
	}
	return false
}
