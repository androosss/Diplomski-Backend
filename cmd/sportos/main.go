package main

import (
	L "backend/internal/logging"
	"backend/sportos/api"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

var CLAPIPort = flag.String("api.pub.port", ":8080", "Public API service port")
var BOAPIPort = flag.String("api.bo.port", ":8081", "Backoffice API service port")
var LOAPIPort = flag.String("api.lo.port", ":8082", "Login service port")
var corsEnable = flag.Bool("cors.enable", false, "Enable CORS headers")

var dbName = flag.String("db.name", "", "name of database")
var dbHost = flag.String("db.host", "", "host where db is located")
var dbPort = flag.String("db.port", "", "port on whitch database is listening")
var dbUser = flag.String("db.user", "", "db user")
var dbPass = flag.String("db.pass", "", "password for database")

var auditEnable = flag.Bool("audit.enable", false, "should audit table start logging when applications starts")

func main() {
	var s api.Server

	flag.Parse()
	// global logger is being initialized through Init()
	L.Init()
	defer L.L.Sync()

	L.L.Info("Server is starting...")
	L.L.Info("db params", L.String("db.name", *dbName), L.String("db.host", *dbHost), L.String("db.port", *dbPort), L.String("db.user", *dbUser), L.String("db.pass", *dbPass))
	L.L.Info("audit params", L.Bool("audit.enable", *auditEnable))

	s.Init(*CLAPIPort, *BOAPIPort, *LOAPIPort, *corsEnable, *dbName, *dbHost, *dbPort, *dbUser, *dbPass, *auditEnable)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	go s.Run()

	<-done
	//Graceful shutdown
	s.Stop()
}
