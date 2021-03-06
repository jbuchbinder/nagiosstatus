package nagiosstatus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	reS               = regexp.MustCompile(`(\w+) {`)
	reE               = regexp.MustCompile(`\t}`)
	reHost            = regexp.MustCompile(`host_name=(.*)`)
	reIg1             = regexp.MustCompile(`^#`)
	reIg2             = regexp.MustCompile(`^\n`)
	defaultStatusFile = "/usr/local/nagios/var/status.dat"
)

// NagiosStatusParser is the parent object of this package, and serves to
// expose methods allowing a Nagios status.dat file to be parsed into an
// internal object representation.
type NagiosStatusParser struct {
	statusFile string
}

// SetStatusFile sets the name of the Nagios status.dat file which is
// read by the package. This defaults to "/usr/local/nagios/var/status.dat".
func (self *NagiosStatusParser) SetStatusFile(file string) {
	self.statusFile = file
}

// Parse parses the specified Nagios status.dat file and creates an internal
// object representation.
func (self *NagiosStatusParser) Parse() *NagiosStatus {
	if self.statusFile == "" {
		self.statusFile = defaultStatusFile
	}
	raws := parseToBlock(readFileToString(self.statusFile))
	s := new(NagiosStatus)
	var wg sync.WaitGroup
	for _, v := range raws {
		wg.Add(1)
		go func(r rawstatus) {
			r.blockToStruct(s, &wg)
		}(v)
	}
	wg.Wait()
	return s
}

// ToJson returns a []byte object containing the JSON representation of
// a parsed Nagios status.dat file. It is a convenience method to pretty
// print the data.
func (self *NagiosStatusParser) ToJson(s *NagiosStatus) []byte {
	j, err := json.MarshalIndent(s, "", "    ")
	checkErr(err)
	return j
}

type rawstatus struct {
	string
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func readFileToString(filename string) string {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return string(b)
}

func parseToBlock(input string) []rawstatus {
	/* each block output should like
	`hoststatus {
		...
		...
	}
	servicestatus {
		...
		...
	}`
	*/
	strs := strings.Split(input, "\n")
	var raws []rawstatus
	var raw rawstatus
	load := false
	// not able to use go func, this process must be step by step process
	for k, v := range strs {
		if (!reIg1.MatchString(v) && !reIg2.MatchString(v)) || k > 7 {
			if reS.MatchString(v) {
				load = true
			} else if reE.MatchString(v) {
				load = false
			}
			raw.string = raw.string + "\n" + v
			if !load {
				if (len(raw.string)) > 1 {
					raw.string = raw.string[1:len(raw.string)]
					raws = append(raws, raw)
				}
				raw.string = ""
			}
		}
	}
	return raws
}

func (raw *rawstatus) blockToStruct(mystat *NagiosStatus, wg *sync.WaitGroup) {
	strs := strings.Split(raw.string, " ")
	ident := strs[0]
	strs = strings.Split(raw.string, "\n")
	strs = strs[1 : len(strs)-1]

	s, host, desc := formatStr(strs)
	stat := make(map[string]interface{})
	for k, v := range s {
		stat[k] = interface{}(v)
	}
	if ident == "hoststatus" {
		if len(mystat.HostStatus) == 0 {
			mystat.HostStatus = make(map[string]*HostStatus)
		}
		var tempstat HostStatus
		a, err := json.Marshal(stat)
		checkErr(err)
		err = json.Unmarshal(a, &tempstat)
		checkErr(err)
		mystat.HostStatus[host] = &tempstat
	} else if ident == "servicestatus" && mystat.HostStatus[host] != nil {
		if len(mystat.HostStatus[host].ServiceStatus) == 0 {
			mystat.HostStatus[host].ServiceStatus = make(map[string]*ServiceStatus)
		}
		var tempstat ServiceStatus
		a, err := json.Marshal(stat)
		checkErr(err)
		err = json.Unmarshal(a, &tempstat)
		checkErr(err)
		mystat.HostStatus[host].ServiceStatus[desc] = &tempstat
	} else if ident == "info" {
		var tempstat Info
		a, err := json.Marshal(stat)
		checkErr(err)
		err = json.Unmarshal(a, &tempstat)
		checkErr(err)
		mystat.Info = tempstat
	} else if ident == "programstatus" {
		var tempstat ProgramStatus
		a, err := json.Marshal(stat)
		checkErr(err)
		err = json.Unmarshal(a, &tempstat)
		checkErr(err)
		mystat.ProgramStatus = tempstat
	}
	wg.Done()
}

// input raw data return format json looks format and hostname as string
func formatStr(strs []string) (map[string]string, string, string) {
	/* input like
	host_name=xxx
	aaa=bbb
	output :
	{"host_name": "xxx", "aaa": bbb} as map format
	*/
	res := make(map[string]string)
	var host, desc string
	for _, s := range strs {
		s = strings.Trim(s, " \t")
		if strings.Contains(s, "host_name") {
			host = strings.Split(s, "=")[1]
		} else if strings.Contains(s, "service_description") {
			desc = strings.Split(s, "=")[1]
		} else {
			a := strings.Split(s, "=")
			key := a[0]
			var value string
			a = a[1:len(a)]
			for _, v := range a {
				value = value + "=" + v
			}
			if len(value) > 1 {
				value = value[1:len(value)] //remove first '='
			}
			if len(a[0]) == 0 { // if value is empty, overwrite
				value = ""
			}
			res[key] = value
		}
	}
	// 	fmt.Println(host, res)
	return res, host, desc
}
