/* -.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.-.

* File Name : gourl.go

* Purpose :

* Creation Date : 01-02-2014

* Last Modified : Mon 28 Jul 2014 11:25:10 PM UTC

* Created By : Kiyor

_._._._._._._._._._._._._._._._._._._._._.*/

package gourl

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type Req struct {
	http.Request
	Url      string
	Timeout  string
	MyHeader []*MyHeader
}

type Resp struct {
	http.Response
}

type MyHeader struct {
	Key   string
	Value string
}

var (
	timeout   time.Duration
	transport = http.Transport{
		Dial:               dialTimeout,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	client = http.Client{
		Transport: &transport,
	}
)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		// 		os.Exit(1)
	}
}

type tmpresp struct {
	resp *http.Response
	err  error
}

func (r *Req) getResp(method string, body io.Reader) (*http.Response, error) {
	var err error
	var req *http.Request
	req, err = http.NewRequest(method, r.Url, body)
	if err != nil {
		// 		fmt.Println("err here1", err.Error())
		return nil, err
	}

	for _, v := range r.MyHeader {
		req.Header.Add(v.Key, v.Value)
	}

	if r.Host != "" {
		req.Host = r.Host
	}

	timeout, err = time.ParseDuration(r.Timeout)
	if err != nil {
		timeout = 3 * time.Second
	}
	t := time.Tick(timeout)
	resp := make(chan tmpresp)

	go func(r chan tmpresp) {
		resp, err := client.Do(req)
		if err != nil {
			// 		fmt.Println("err here2", err.Error())
			r <- tmpresp{resp, err}
		}
		r <- tmpresp{resp, nil}
	}(resp)
	select {
	case r := <-resp:
		return r.resp, r.err
	case <-t:
		return nil, errors.New("Timeout")
	}
	// 	return resp, err
}

func (r *Req) GetFull() (Resp, error) {
	resp, err := r.getResp("GET", nil)
	if err != nil {
		return Resp{}, err
	}
	return Resp{*resp}, err
}

func (r *Req) PostStringGetFull(body string) (Resp, error) {
	b := strings.NewReader(body)
	resp, err := r.getResp("POST", b)
	if err != nil {
		return Resp{}, err
	}
	return Resp{*resp}, err
}

func (r *Req) PostStringGetString(body string) (string, error) {
	resp, err := r.PostStringGetFull(body)
	return resp.String(), err
}

func (r *Resp) String() string {
	if r.StatusCode != 200 {
		return ""
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		body = []byte("")
	}
	defer r.Body.Close()
	var s string
	if len(body) == 0 {
		s = ""
	} else {
		last := string(body)[len(string(body))-1 : len(string(body))]
		if last == "\n" || last == " " {
			s = string(body)[:len(string(body))-1]
		} else {
			s = string(body)
		}
	}
	return s
}

func (r *Resp) StringSlice() []string {
	if r.StatusCode != 200 {
		return []string{""}
	}
	s := r.String()
	slice := strings.Split(s, "\n")
	return slice[0 : len(slice)-1]
}

func (r *Req) GetString() (string, error) {
	resp, err := r.GetFull()
	return resp.String(), err
}

func (r *Req) GetStringSlice() ([]string, error) {
	s, err := r.GetString()
	if err != nil {
		return nil, err
	}
	slice := strings.Split(s, "\n")
	return slice[0 : len(slice)-1], nil
}

func (r *Req) GetHeader() (http.Header, error) {
	// 	resp, err := http.Head(r.Url)
	// 	checkErr(err)
	// 	return resp.Header
	resp, err := r.getResp("HEAD", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, err
}
