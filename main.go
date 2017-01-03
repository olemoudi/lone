package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	//"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strings"
)

const banner = "lone"

var (
	infoLog   *log.Logger
	debugLog  *log.Logger
	debugMode bool
)

func main() {
	fmt.Println(banner)
	//	flag.StringVar(&starturl, "u", "", "Base URL")
	flag.BoolVar(&debugMode, "debug", false, "log additional debug traces")
	//	flag.BoolVar(&serverMode, "server", false, "launch testing server")

	flag.Parse()

	LogInit(debugMode)

	info("launching server at :8000")
	// global handler = polite and dupe tests
	r := mux.NewRouter()
	//r.HandleFunc("/", globalHandler(rootHandler))
	//r.HandleFunc("/", globalHandler(http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/log/{App}/{Sink}", globalHandler(LogHandler))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/tmp"))))
	http.Handle("/", r)
	http.ListenAndServe(":8000", nil)
	os.Exit(0)
}

func globalHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Server", "Lone 0.1")
		/*
			fail, diff := politenessTest(r)
			if fail {
				fmt.Println("Politeness Test failed by", strconv.Itoa(int(diff)))
			}
			if dupe(r) {
				fmt.Println("URL visited twice:", r.URL.String())
			}
		*/
		fn(w, r)
	}
}

func LogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Redirect(w, r, "/", 301)
	}
	vars := mux.Vars(r)
	app, okapp := vars["App"]
	sink, oksink := vars["Sink"]
	if !okapp || !oksink {
		http.Redirect(w, r, "/", 301)
	}
	err := ioutil.WriteFile(fmt.Sprintf("/tmp/%s.%s.csv", app, sink), []byte(fmt.Sprintf("%s\n", r.Method)), 0666)
	sinkFile, err := os.OpenFile(fmt.Sprintf("/tmp/%s.%s.csv", app, sink), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		info("Error")
		return
	}
	defer sinkFile.Close()
	_, err = io.Copy(sinkFile, strings.NewReader(r.Method+"\n"))
	//_, err = sinkFile.WriteString("\n")
}

/*
/////////////////////////
/////////////////////////
UTLIITY FUNCTIONS
/////////////////////////
/////////////////////////

*/

func LogInit(debug_flag bool) {
	logfile, err := os.OpenFile("/tmp/lone.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file")
	}
	infowriter := io.MultiWriter(logfile, os.Stdout)

	if debug_flag {
		debuglogfile, err := os.OpenFile("/tmp/lone.debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Error opening debug log file")
		}

		infowriter = io.MultiWriter(logfile, os.Stdout, debuglogfile)

		debugwriter := io.MultiWriter(debuglogfile, os.Stdout)
		debugLog = log.New(debugwriter, "[DEBUG] ", log.Ldate|log.Ltime)

	} else {
		debugLog = log.New(ioutil.Discard, "", 0)
	}

	infoLog = log.New(infowriter, "", log.Ldate|log.Ltime)

}

func info(msg ...string) {
	s := make([]interface{}, len(msg))
	for i, v := range msg {
		s[i] = v
	}
	infoLog.Println(s...)
}

func debug(msg ...string) {
	s := make([]interface{}, len(msg))
	for i, v := range msg {
		s[i] = v
	}
	debugLog.Println(s...)
}
