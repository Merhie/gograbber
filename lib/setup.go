package lib

import (
	"fmt"
	"math"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
)

// Initialise sets up the program's state
func Initialise(s *State, ports string, wordlist string, statusCodesIgn string, protocols string) (errors *multierror.Error) {
	s.URLProvided = false
	if wordlist != "" {
		pathData, err := GetDataFromFile(wordlist)
		if err != nil {
			panic(err)
		}
		s.Paths = StringSet{Set: map[string]bool{}}
		for _, path := range pathData {
			s.Paths.Add(path)
		}
	}
	if s.URLFile != "" || s.SingleURL != "" { // A url and/or file full of urls was supplied - treat them as gospel
		s.URLProvided = true
		if s.URLFile != "" {
			s.Scan = false
			inputData, err := GetDataFromFile(s.URLFile)
			if err != nil {
				panic(err)
			}

			for _, item := range inputData {
				// s.URLComponents
				h, err := ParseURLToHost(item)
				if err != nil {
					continue
				}
				s.URLComponents = append(s.URLComponents, h)
			}
		} else if s.SingleURL != "" {
			h, err := ParseURLToHost(s.SingleURL)
			if err == nil {
				s.URLComponents = append(s.URLComponents, h)
			}
		}
		s.Scan = false
		return
	}
	if ports != "" {
		for _, port := range StrArrToInt(strings.Split(ports, ",")) {
			if v := int(math.Pow(2, 16.0)); 0 > port || port >= v {
				fmt.Printf("Port: (%v) is invalid!\n", port)
				continue
			}
			s.Ports.Add(port)
		}
		// for p, _ := range s.Ports.Set {
		// 	fmt.Printf("%v\n", p)
		// }
	}
	if s.InputFile != "" {
		inputData, err := GetDataFromFile(s.InputFile)
		if err != nil {
			panic(err)
		}
		c := make(chan StringSet)
		go ExpandHosts(inputData, c)
		targetList := <-c
		if s.Debug {
			for target := range targetList.Set {
				fmt.Printf("Target: %v\n", target)
			}
		}
		s.Hosts = targetList
	}
	s.StatusCodesIgn = IntSet{map[int]bool{}}
	for _, code := range StrArrToInt(strings.Split(statusCodesIgn, ",")) {
		s.StatusCodesIgn.Add(code)
	}

	c2 := make(chan []Host)
	s.Protocols = StringSet{map[string]bool{}}
	for _, p := range strings.Split(protocols, ",") {
		s.Protocols.Add(p)
	}
	go GenerateURLs(s.Hosts, s.Ports, &s.Paths, c2)
	s.URLComponents = <-c2
	return
}

// Start does the thing
func Start(s State) {

	if s.Scan {
		fmt.Printf("Starting Port Scanner\n")
		if s.Debug {
			fmt.Printf("Testing %v host:port combinations\n", len(s.URLComponents))
		}
		fmt.Printf(LineSep())
		s.URLComponents = ScanHosts(&s)

		fmt.Printf(LineSep())
	}
	if s.Dirbust {
		fmt.Printf("Starting Dirbuster\n")
		if s.Debug {
			var numURLs int
			if len(s.Paths.Set) != 0 {
				numURLs = len(s.URLComponents) * len(s.Paths.Set)
			} else {
				numURLs = len(s.URLComponents)
			}
			fmt.Printf("Testing %v URLs\n", numURLs)
		}
		fmt.Printf(LineSep())

		s.URLComponents = DirbustHosts(&s)
		if s.Debug {
			fmt.Println(s.URLComponents)
		}
		fmt.Printf(LineSep())
	}
	if s.Screenshot {
		fmt.Printf("Starting Screenshotter\n")
		if s.Debug {
			fmt.Printf("Testing %v URLs\n", len(s.URLComponents)*len(s.Paths.Set))
		}
		s.URLComponents = Screenshot(&s)
		fmt.Printf(LineSep())
	}
}
