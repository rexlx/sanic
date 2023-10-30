package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var port int = 6666

type Application struct {
	Domain    string               `json:"domain"`
	Port      int                  `json:"port"`
	Log       *log.Logger          `json:"-"`
	Instances map[string]*Instance `json:"instances"`
	Server    *http.Server         `json:"-"`
	Visits    int                  `json:"visits"`
}

type Site struct {
	Name     string
	UI       UIConfig
	Handlers []Handler
}

type Handler struct {
	Name string
	Func func(http.ResponseWriter, *http.Request)
}

func main() {
	newLog := log.New(os.Stdout, "app: ", log.LstdFlags)

	app := &Application{
		Log:       newLog,
		Port:      8080,
		Domain:    "rxlx.us",
		Instances: make(map[string]*Instance),
	}

	for _, route := range routes {
		app.Log.Println("Adding route", route.Name)

		hostCfg := HostConfig{
			Domain:    app.Domain,
			IP:        "127.0.0.1",
			Port:      port,
			SubDomain: route.Name,
		}
		port++

		instance := NewInstance(hostCfg, route.UI)
		instance.ID = route.Name
		for _, handler := range route.Handlers {
			instance.AddHandler(handler.Name, handler.Func)
		}
		// TODO: need way to handle dynamic routes
		instance.Server.HandleFunc("/", instance.RootHandler)
		instance.Server.HandleFunc("/runtime", instance.GetRuntimeStats)

		app.AddInstance(route.Name, instance)
		go instance.Start()

	}

	app.Log.Println("Starting main server")
	app.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Port),
		Handler: app,
	}

	app.Server.ListenAndServe()
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

	url := fmt.Sprintf("http://localhost:%d/%v", svc.Port, path)

	a.Log.Println("forwarding request to", url)

	req, err := http.NewRequest(r.Method, url, r.Body)
	// req, err := http.NewRequest(r.Method, svc.URL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Header = r.Header
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	req.Header.Set("X-Forwarded-Host", r.Host)

	client := &http.Client{}
	a.Log.Println("making a client and performing request")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (a *Application) tidyDomain(domain []string) ([]string, error) {
	out := make([]string, 0)
	parts := strings.Split(a.Domain, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Domain is not valid")
	}
	for _, part := range domain {
		if strings.Contains(part, ":") || part == parts[0] {
			continue
		}
		out = append(out, part)
	}
	return out, nil
}
