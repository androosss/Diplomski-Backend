package api

import (
	L "backend/internal/logging"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Timeout in seconds
const Timeout = 5

// Server struct contains server configuration
type Server struct {
	Repo *crud.Repo
	// Cache      repo.Cache
	SubServers map[DR.SubServer]*SubServer
	CorsEnable bool
}

type SubServer struct {
	HttpServer http.Server
	MuxRouter  *mux.Router
}

func newSubServer(port string) (ss *SubServer) {
	ss = &SubServer{}
	ss.MuxRouter = mux.NewRouter()
	ss.HttpServer.Handler = ss.MuxRouter
	ss.HttpServer.Addr = port
	return
}

// Init starts server
func (s *Server) Init(CLPort, BOPort, LOPort string, corsEnable bool, dbName, dbHost, dbPort, dbUser, dbPass string, auditEnable bool) {
	s.SubServers = make(map[DR.SubServer]*SubServer)
	s.SubServers[DR.SUB_LO] = newSubServer(LOPort)
	s.SubServers[DR.SUB_CL] = newSubServer(CLPort)
	s.SubServers[DR.SUB_BO] = newSubServer(BOPort)

	dbConnection := crud.DBConnection{
		DBName:   dbName,
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPass,
	}
	s.Repo = dbConnection.InitRepo()

	s.CorsEnable = corsEnable

	registerHandlers(s)

	L.L.Info("Server is set up...", L.Any("SubServers", s.SubServers))

	if auditEnable {
		s.Repo.AuditCrud.Start()
	}
}

// Run starts server
func (s *Server) Run() {
	wg := new(sync.WaitGroup)
	wg.Add(len(s.SubServers))
	for _, ser := range s.SubServers {
		// we need to save the range value in the loop since it changes faster than the go func() executes where racing condition occurs
		server := ser
		go func() {
			if err := server.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				L.L.Fatal("Sub Server shutdown with error", L.String("Addr", server.HttpServer.Addr), L.Error(err))
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// Stop server
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout*time.Second)
	defer cancel()

	for _, ser := range s.SubServers {
		L.L.Info("Stopping SubServer...", L.String("Addr", ser.HttpServer.Addr))
		if err := ser.HttpServer.Shutdown(ctx); err != nil {
			L.L.Fatal("Client API Server Shutdown", L.String("Addr", ser.HttpServer.Addr), L.Error(err))
		}
	}
}
