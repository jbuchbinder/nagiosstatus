/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : nginx.go

* Purpose :

* Creation Date : 01-10-2014

* Last Modified : Thu 01 May 2014 09:30:51 PM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package gourl

import (
	// 	"fmt"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	re1 = regexp.MustCompile(`(\d+)`)
	re2 = regexp.MustCompile(`Reading:\s(\d+)\sWriting:\s(\d+)\sWaiting:\s(\d+)\s`)
	re3 = regexp.MustCompile(`(\d+) (\d+) (\d+)`)
	uri = "/server-status"
)

type NginxStatus struct {
	Host            string
	Active          int
	Reading         int
	Writing         int
	Waiting         int
	LastUpdate      int64
	TTL             int64
	SinceLastUpdate int64
	StatusUri       string
	Server          NginxServer
}

type NginxServer struct {
	Accepts  int
	Handled  int
	Requests int
}

func (s *NginxServer) Init() {
	s.Accepts = -1
	s.Handled = -1
	s.Requests = -1
}

func (stat *NginxStatus) Init() {
	stat.Active = -1
	stat.Reading = -1
	stat.Writing = -1
	stat.Waiting = -1
	stat.LastUpdate = -1
	stat.TTL = 60
	stat.SinceLastUpdate = -1
	if stat.StatusUri == "" {
		stat.StatusUri = uri
	}
	var s NginxServer
	s.Init()
	stat.Server = s
}

func (stat *NginxStatus) UpdateSinceLastUpdate() {
	stat.SinceLastUpdate = time.Now().Unix() - stat.LastUpdate
}

func (stat *NginxStatus) Update() error {
	var err error
	url := "http://" + stat.Host + stat.StatusUri
	var r Req
	r.Url = url
	r.Timeout = "2s"
	fullbody, err := r.GetString()
	if err != nil {
		return err
	}
	body := strings.Split(fullbody, "\n")
	if len(body) == 0 {
		return errors.New("not able to get correct resp")
	}
	activeLine := body[0]

	serverLine := body[2]

	mistLine := body[len(body)-1]
	active := re1.FindStringSubmatch(activeLine)[0]
	stat.Active, err = strconv.Atoi(active)
	if err != nil {
		return err
	}
	mist := re2.FindStringSubmatch(mistLine)

	server := re3.FindStringSubmatch(serverLine)

	stat.Reading, err = strconv.Atoi(mist[1])
	if err != nil {
		return err
	}
	stat.Writing, err = strconv.Atoi(mist[2])
	if err != nil {
		return err
	}
	stat.Waiting, err = strconv.Atoi(mist[3])
	if err != nil {
		return err
	}
	stat.LastUpdate = time.Now().Unix()

	var sv NginxServer

	sv.Accepts, err = strconv.Atoi(server[1])
	if err != nil {
		return err
	}
	sv.Handled, err = strconv.Atoi(server[2])
	if err != nil {
		return err
	}
	sv.Requests, err = strconv.Atoi(server[3])
	if err != nil {
		return err
	}

	stat.Server = sv
	return nil
}
