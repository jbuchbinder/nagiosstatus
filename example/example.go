package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kiyor/gourl/lib"
	nagiosstatus "../"
	"io/ioutil"
	"log"
	"strconv"
)

var (
	statusf *string = flag.String("f", "/var/log/nagios/status.dat", "status file")
	all     *bool   = flag.Bool("all", false, "get all info")
	mute    *bool   = flag.Bool("mute", false, "enable show mute info")
	ack     *bool   = flag.Bool("ack", false, " enable show ack")
	url     *string = flag.String("url", "", "get status file by url")
	n nagiosstatus.NagiosStatusParser
)

func init() {
	flag.Parse()
	if *url != "" {
		setStatByUrl(*url)
	} else {
		n.SetStatusFile(*statusf)
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
	n.SetStatusFile("/tmp/temp.dat")
}

func main() {
	var stat nagiosstatus.NagiosStatus
	a := n.ToJson(n.Parse())
	json.Unmarshal(a, &stat)
	if *all {
		j, _ := json.MarshalIndent(stat, "", "    ")
		fmt.Println(string(j))
	} else {
		for hostname, v := range stat.HostStatus {
			if state(v) != 0 {
				fmt.Println(hostname, v.Plugin_output)
			}
			for servicename, v2 := range v.ServiceStatus {
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
	case *nagiosstatus.HostStatus:
		res, _ = strconv.Atoi(v.Current_state)
		return res
	case *nagiosstatus.ServiceStatus:
		res, _ = strconv.Atoi(v.Current_state)
		return res
	}
}
func notifications(v interface{}) bool {
	var res int
	switch v := v.(type) {
	default:
		return false
	case *nagiosstatus.HostStatus:
		res, _ = strconv.Atoi(v.Notifications_enabled)
		if res == 1 {
			return true
		}
	case *nagiosstatus.ServiceStatus:
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
	case *nagiosstatus.HostStatus:
		res, _ = strconv.Atoi(v.Problem_has_been_acknowledged)
		if res == 1 {
			return true
		}
	case *nagiosstatus.ServiceStatus:
		res, _ = strconv.Atoi(v.Problem_has_been_acknowledged)
		if res == 1 {
			return true
		}
	}
	return false
}
