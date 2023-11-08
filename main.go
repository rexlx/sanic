package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	port        = flag.Int("port", 6666, "port to start on")
	exposedPort = flag.Int("exposedPort", 8080, "port to expose")
	cert        = flag.String("cert", "cert.pem", "path to cert")
	key         = flag.String("key", "key.pem", "path to key")
)

type Application struct {
	TLSConfig *tls.Config          `json:"-"`
	Domain    string               `json:"domain"`
	Port      int                  `json:"port"`
	Log       *log.Logger          `json:"-"`
	Instances map[string]*Instance `json:"instances"`
	Server    *http.Server         `json:"-"`
	Visits    int                  `json:"visits"`
}

type Site struct {
	Name      string
	UI        UIConfig
	Handlers  []Handler
	ServePath string
}

type Handler struct {
	Name string
	Func func(http.ResponseWriter, *http.Request)
}

func main() {
	flag.Parse()
	newLog := log.New(os.Stdout, "app: ", log.LstdFlags)
	cert, err := tls.LoadX509KeyPair(*cert, *key)
	if err != nil {
		newLog.Fatalln("Error loading cert", err)
	}

	app := &Application{
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Log:       newLog,
		Port:      8080,
		Domain:    "rxlx.us",
		Instances: make(map[string]*Instance),
	}

	for _, route := range routes {
		app.Log.Println("Adding route", route.Name, route.ServePath)

		hostCfg := HostConfig{
			Domain:    app.Domain,
			IP:        "0.0.0.0",
			Port:      *port,
			SubDomain: route.Name,
		}
		*port++

		instance := NewInstance(hostCfg, route.UI)
		instance.ID = route.Name
		instance.Server.HandleFunc("/home", instance.RootHandler)
		instance.Server.HandleFunc("/runtime", instance.GetRuntimeStats)
		if route.ServePath != "" {
			err := checkPath(route.ServePath)
			if err != nil {
				app.Log.Println("Error checking path", err)
				continue
			}
			instance.ServePath = route.ServePath
			fs := http.FileServer(http.Dir(instance.ServePath))
			instance.Server.Handle("/", fs)
		}

		for _, handler := range route.Handlers {
			instance.AddHandler(handler.Name, handler.Func)
		}

		app.AddInstance(route.Name, instance)
		go instance.Start()

	}

	app.Log.Println("Starting main server")
	app.Server = &http.Server{
		TLSConfig: app.TLSConfig,
		Addr:      fmt.Sprintf(":%d", app.Port),
		Handler:   app,
	}

	// app.Server.ListenAndServe()
	app.Server.ListenAndServeTLS("", "")
}

func (a *Application) AddInstance(path string, instance *Instance) {
	a.Instances[path] = instance
}

func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimLeft(r.URL.Path, "/")
	subDomain, err := a.tidyDomain(strings.Split(r.Host, "."))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.Log.Println("got request from", r.RemoteAddr, "for", path)

	if len(subDomain) != 1 {
		http.Error(w, fmt.Sprintf("(%v)\tNot Found %v", len(subDomain), r.Host), http.StatusNotFound)
		return
	}

	svc, ok := a.Instances[subDomain[0]]
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	r.Host = "localhost"

	svc.Server.ServeHTTP(w, r)

}

func (a *Application) tidyDomain(domain []string) ([]string, error) {
	out := make([]string, 0)
	parts := strings.Split(a.Domain, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("domain is not valid")
	}
	for _, part := range domain {
		if strings.Contains(part, ":") || part == parts[0] {
			continue
		}
		out = append(out, part)
	}
	return out, nil
}

func checkPath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path %v does not exist", path)
	}
	return nil
}
