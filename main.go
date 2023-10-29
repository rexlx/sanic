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
	Log       *log.Logger          `json:"-"`
	Instances map[string]*Instance `json:"instances"`
	Server    *http.Server         `json:"-"`
	Visits    int                  `json:"visits"`
}

func main() {
	routes := []string{"", "about", "contact", "index", "blog"}
	newLog := log.New(os.Stdout, "app: ", log.LstdFlags)
	app := &Application{
		Log:       newLog,
		Domain:    "sanic",
		Instances: make(map[string]*Instance),
	}

	for _, route := range routes {
		app.Log.Println("Adding route", route)
		style := BasicStyle{
			BodyBG:   "#f5f5f5",
			BodyText: "#333",
			H1:       "#444",
			Btn:      "#becdc3",
			BtnText:  "#000",
		}
		templates := []Template{
			{
				Name: "index",
				Body: splashPage,
			},
		}
		uiCfg := UIConfig{
			Style:     style,
			Templates: templates,
		}
		hostCfg := HostConfig{
			Domain:    "sanic",
			IP:        "0.0.0.0",
			Port:      port,
			SubDomain: route,
		}
		port++
		instance := NewInstance(hostCfg, uiCfg)
		instance.ID = route
		instance.Domain = app.Domain
		instance.Server.HandleFunc("/", instance.RootHandler)
		instance.Server.HandleFunc("/runtime", instance.GetRuntimeStats)
		instance.Server.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "mycon")
		})
		app.AddInstance(route, instance)
		go instance.Start()

	}
	app.Log.Println("Starting main server")
	app.Server = &http.Server{
		Addr:    ":8080",
		Handler: app,
	}
	app.Server.ListenAndServe()
}

func (a *Application) AddInstance(path string, instance *Instance) {
	a.Instances[path] = instance
}

func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimLeft(r.URL.Path, "/")
	subDomain := a.tidyDomain(strings.Split(r.Host, "."))
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

func (a *Application) tidyDomain(domain []string) []string {
	out := make([]string, 0)
	for _, part := range domain {
		if strings.Contains(part, ":") || part == a.Domain {
			continue
		}
		out = append(out, part)
	}
	return out
}
