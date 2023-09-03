// SPORTOS is a service that provides payment routing of depositing or withdrawing actions.
// It provides communication with different payment providers and with PAM (Player Account Manager) service.
//
// Usage
//
// Server parameters
//
//  -db.name string
//        sportos database name
//  -db.host string
//        host where sportos database is located
//  -db.port string
//        sportos databaseb port
//  -db.user string
//        sportos database user
//  -db.pass string
//        database password
//  -api.pub.port string
//        public API service port. default is 8080
//  -api.bo.port string
//        backoffice API service port. default is 8081
//  -api.pp.port string
//        service port for API used by payment providers. default is 8082
//  -api.mgmt.port string
//        port for metrics and log services: /loglevel /log /health default is 8880
//  -llev
//        loglevel (debug, info, warn, error, dpanic, panic, fatal)
//  -cors.enable boolean
//        enable CORS headers in HTTP requests
//  -scheduler.enable boolean
//        should scheduler start when applications starts. default is false
//  -scheduler.interval int
//        scheduler interval (in miliseconds)
// 	-business.webhookNotificationsEndpoint
//        Sportos endpoint for receiving webhook notifications
//	-audit.enable boolean
//		  should audit table be filled when application start. default is false
//  Example: .\sportos.exe -'db.name' sportos -'db.host' localhost -'db.port' 5432 -'db.user' postgres -'db.pass' secret -'scheduler.enable' true -'scheduler.interval' 1000 -'audit.enable' true -'business.webhookNotificationsEndpoint' https://sportos-notifications.fincoreltd.rs
package main
