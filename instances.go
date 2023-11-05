package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Instance struct {
	Log         *log.Logger    `json:"-"`
	ID          string         `json:"id"`
	ServePath   string         `json:"serve_path"`
	Style       BasicStyle     `json:"style"`
	Templates   []Template     `json:"templates"`
	SubDomain   string         `json:"subdomain"`
	Domain      string         `json:"domain"`
	IP          string         `json:"ip"`
	Port        int            `json:"port"`
	URL         string         `json:"url"`
	Stats       RuntimeStats   `json:"stats"`
	Server      *http.ServeMux `json:"-"`
	KillChan    chan bool      `json:"-"`
	MessageChan chan SmallTalk `json:"-"`
}

type HostConfig struct {
	Domain    string `json:"domain"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	SubDomain string `json:"subdomain"`
}

type UIConfig struct {
	Style     BasicStyle `json:"style"`
	Templates []Template `json:"templates"`
}

type SmallTalk struct {
	InstanceID string    `json:"instance_id"`
	Time       time.Time `json:"time"`
	Errors     []string  `json:"errors"`
	Visits     int       `json:"visits"`
}

type RuntimeStats struct {
	MX     *sync.RWMutex `json:"-"`
	Start  time.Time     `json:"start"`
	Visits int           `json:"visits"`
	Errors []string      `json:"errors"`
}

type Template struct {
	Name string
	Body string
}

type BasicStyle struct {
	BodyBG   string
	BodyText string
	H1       string
	Btn      string
	BtnText  string
}

func (i *Instance) AddError(err string) {
	i.Stats.MX.Lock()
	defer i.Stats.MX.Unlock()
	i.Stats.Errors = append(i.Stats.Errors, err)
}

func (i *Instance) AddVisit() {
	i.Stats.MX.Lock()
	defer i.Stats.MX.Unlock()
	i.Stats.Visits++
}

func (i *Instance) GetStats() RuntimeStats {
	i.Stats.MX.RLock()
	defer i.Stats.MX.RUnlock()
	return i.Stats
}

func (i *Instance) RootHandler(w http.ResponseWriter, r *http.Request) {
	i.Log.Println(i.ID, "Root handler called")
	fmt.Println("Root handler called")
	name := "index"
	url := fmt.Sprintf("http://%v.%v:%d", i.SubDomain, i.Domain, 8080)
	var tmpl string
	for _, t := range i.Templates {
		if t.Name == name {
			tmpl = t.Body
		}
	}
	i.AddVisit()
	// fmt.Println(i.ID, fmt.Sprintf(tmpl, url, addMinimalStyling(i.Style)))
	fmt.Fprintf(w, fmt.Sprintf(tmpl, url, addMinimalStyling(i.Style)))
}

func (i *Instance) GetRuntimeStats(w http.ResponseWriter, r *http.Request) {
	// i.Log.Println(i.ID, "Getting runtime stats", i.GetStats())
	res := i.GetStats()
	out := fmt.Sprintf("<small>%v visits; running for %v <br>", res.Visits, time.Since(res.Start))
	fmt.Fprintf(w, fmt.Sprintf("%v", out))
}

func (i *Instance) Start() {
	newHTTP := &http.Server{
		Addr:    fmt.Sprintf(":%v", i.Port),
		Handler: i.Server,
	}

	defer i.Stop()

	go func(newHTTP *http.Server) {
		if err := newHTTP.ListenAndServe(); err != nil {
			i.Log.Printf("Error starting server: %s", err)
			i.AddError(err.Error())
			return
		}
	}(newHTTP)
	for {
		select {
		case <-i.KillChan:
			i.Log.Printf("Stopping server on %s", i.URL)
			if err := newHTTP.Close(); err != nil {
				i.Log.Printf("Error stopping server: %s", err)
				i.AddError(err.Error())
			}
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func (i *Instance) Stop() {
	i.KillChan <- true
}

func (i *Instance) AddHandler(path string, handler http.HandlerFunc) {
	i.Server.HandleFunc(path, handler)
}

func NewInstance(hostCfg HostConfig, uiCfg UIConfig) *Instance {
	newLog := log.New(os.Stdout, fmt.Sprintf("%s.%s -> ", hostCfg.SubDomain, hostCfg.Domain), log.LstdFlags)
	urlString := fmt.Sprintf("http://%s.%s:%d", hostCfg.SubDomain, hostCfg.Domain, hostCfg.Port)
	errs := make([]string, 0)
	mx := &sync.RWMutex{}
	killChan := make(chan bool)
	messageChan := make(chan SmallTalk)
	svr := http.NewServeMux()

	return &Instance{
		Log:         newLog,
		SubDomain:   hostCfg.SubDomain,
		Domain:      hostCfg.Domain,
		IP:          hostCfg.IP,
		URL:         urlString,
		Port:        hostCfg.Port,
		Style:       uiCfg.Style,
		Templates:   uiCfg.Templates,
		KillChan:    killChan,
		MessageChan: messageChan,
		Server:      svr,
		Stats: RuntimeStats{
			Start:  time.Now(),
			MX:     mx,
			Errors: errs,
		},
	}
}

func addMinimalStyling(bs BasicStyle) string {
	styleString := `
	<style>
	  body{font-family:Arial,Helvetica,sans-serif;font-size:16px;line-height:1.5;margin:0;padding:0;background-color:%v;color:%v;}
	  h1{font-size:2rem;margin-bottom:1rem;color:%v;}
	  label{margin-bottom:0.5rem;}input{padding:0.5rem;margin-bottom:1rem;border-radius:0.25rem;border:1px solid #ccc;}
	  table{border-collapse:collapse;}
	  th,td{padding:0.5rem;}
	  tr{border-bottom: 1px solid #ddd;}
	  tr:nth-child(even){background-color: #D6EEEE;}
	  button{padding:0.5rem 1rem;background-color:%v;color:%v;border:none;border-radius:0.25rem;cursor:pointer;}
	</style>`
	return fmt.Sprintf(styleString, bs.BodyBG, bs.BodyText, bs.H1, bs.Btn, bs.BtnText)
}

var splashPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>w e l c o m e</title>
  <script src="https://unpkg.com/htmx.org@1.9.6" integrity="sha384-FhXw7b6AlE/jyjlZH5iHa/tTe9EpJ1Y55RjcgPbjeWMskSxZt1v9qkxLJWNJaGni" crossorigin="anonymous"></script>
</head>
<body>
  <h1>thanks for visiting!</h1>
  <div id="runtime" hx-trigger="every 2s" hx-get="%v/runtime">runtime stats</div>
  <div class="target" id="target"></div>
  <div id="content"><hr><br /><h2>this is it</h2></div>
	<div id="guests"></div>
	%v
	<script>

    </script>
</body>
</html>`
