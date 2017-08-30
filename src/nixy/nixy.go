package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/peterbourgon/g2s"
)

// Task struct
type Task struct {
	Host         string
	Ports        []int64
	ServicePorts []int64
	StagedAt     string
	StartedAt    string
	Version      string
	AppID        string
	ID           string
}

// PortDefinitions struct
type PortDefinitions struct {
	Port     int64
	Protocol string
	Labels   map[string]string
}

// App struct
type App struct {
	NUM             string
	NAME            string
	Tasks           []Task
	Labels          map[string]string
	Env             map[string]string
	Hosts           []string
	PortDefinitions []PortDefinitions
}

// Config struct used by the template engine
type Config struct {
	sync.RWMutex
	Xproxy           string
	Realm            string
	Port             string   `json:"-"`
	Marathon         []string `json:"-"`
	User             string   `json:"-"`
	Pass             string   `json:"-"`
	NginxConfig      string   `json:"-" toml:"nginx_config"`
	NginxTemplate    string   `json:"-" toml:"nginx_template"`
	NginxCmd         string   `json:"-" toml:"nginx_cmd"`
	NginxIgnoreCheck bool     `json:"-" toml:"nginx_ignore_check"`
	LeftDelimiter    string   `json:"-" toml:"left_delimiter"`
	RightDelimiter   string   `json:"-" toml:"right_delimiter"`
	Statsd           StatsdConfig
	LastUpdates      Updates
	Apps             map[string]App
	Servers          map[string][]App
}

// Updates timings used for metrics
type Updates struct {
	LastSync           time.Time
	LastConfigRendered time.Time
	LastConfigValid    time.Time
	LastNginxReload    time.Time
}

// StatsdConfig statsd stuct
type StatsdConfig struct {
	Addr       string
	Namespace  string
	SampleRate int `toml:"sample_rate"`
}

// Status health status struct
type Status struct {
	Healthy bool
	Message string
}

// EndpointStatus health status struct
type EndpointStatus struct {
	Endpoint string
	Healthy  bool
	Message  string
}

// Health struct
type Health struct {
	Config    Status
	Template  Status
	Endpoints []EndpointStatus
}

// Global variables
var version = "master" //set by ldflags
var date string        //set by ldflags
var commit string      //set by ldflags
var config = Config{LeftDelimiter: "{{", RightDelimiter: "}}"}
var statsd g2s.Statter
var health Health
var lastConfig string
var logger = logrus.New()

// Eventqueue with buffer of two, because we dont really need more.
var eventqueue = make(chan bool, 2)

// Global http transport for connection reuse
var tr = &http.Transport{MaxIdleConnsPerHost: 10}

func newHealth() Health {
	var h Health
	for _, ep := range config.Marathon {
		var s EndpointStatus
		s.Endpoint = ep
		s.Healthy = true
		s.Message = "OK"
		h.Endpoints = append(h.Endpoints, s)
	}
	return h
}

func nixyReload(w http.ResponseWriter, r *http.Request) {
	logger.WithFields(logrus.Fields{
		"client": r.RemoteAddr,
	}).Info("marathon reload triggered")
	select {
	case eventqueue <- true: // Add reload to our queue channel, unless it is full of course.
		w.WriteHeader(202)
		fmt.Fprintln(w, "queued")
		return
	default:
		w.WriteHeader(202)
		fmt.Fprintln(w, "queue is full")
		return
	}
}

func nixyHealth(w http.ResponseWriter, r *http.Request) {
	err := checkTmpl()
	if err != nil {
		health.Template.Message = err.Error()
		health.Template.Healthy = false
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		health.Template.Message = "OK"
		health.Template.Healthy = true
	}
	err = checkConf(lastConfig)
	if err != nil {
		health.Config.Message = err.Error()
		health.Config.Healthy = false
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		health.Config.Message = "OK"
		health.Config.Healthy = true
	}
	allBackendsDown := true
	for _, endpoint := range health.Endpoints {
		if endpoint.Healthy {
			allBackendsDown = false
			break
		}
	}
	if allBackendsDown {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	b, _ := json.MarshalIndent(health, "", "  ")
	w.Write(b)
	return
}

func nixyConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	b, _ := json.MarshalIndent(&config, "", "  ")
	w.Write(b)
	return
}

func nixyVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "version: "+version)
	fmt.Fprintln(w, "commit: "+commit)
	fmt.Fprintln(w, "date: "+date)
	return
}

func main() {
	configtoml := flag.String("f", "/etc/nixy/nixy.toml", "Path to config. (default nixy.toml)")
	versionflag := flag.Bool("v", false, "prints current nixy version")
	nginxGroup := flag.String("g", "", "nginx dicover service group")
	marathon := flag.String("m", "", "marathon service url")
	port := flag.String("p", "", "nixy status port")
	flag.Parse()

	fmt.Printf("group is %s\n", *nginxGroup)

	if *versionflag {
		fmt.Printf("version: %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("date: %s\n", date)
		os.Exit(0)
	}

	file, err := ioutil.ReadFile(*configtoml)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatal("problem opening toml config")
	}
	err = toml.Unmarshal(file, &config)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatal("problem parsing config")
	}
	if nginxGroup != nil && *nginxGroup != "" {
		config.Realm = *nginxGroup
	}
	if marathon != nil && *marathon != "" {
		mhost := strings.Split(*marathon, ",")
		config.Marathon = mhost
	}
	if port != nil && *port != "" {
		config.Port = *port
	}

	// Lets default empty Xproxy to hostname.
	if config.Xproxy == "" {
		config.Xproxy, _ = os.Hostname()
	}
	statsd, _ = setupStatsd()
	mux := mux.NewRouter()
	mux.HandleFunc("/", nixyVersion)
	mux.HandleFunc("/v1/reload", nixyReload)
	mux.HandleFunc("/v1/config", nixyConfig)
	mux.HandleFunc("/v1/health", nixyHealth)
	s := &http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}
	health = newHealth()
	endpointHealth()
	eventStream()
	eventWorker()
	logger.Info("starting nixy on :" + config.Port)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
