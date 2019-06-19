package main

import (
    "os"
    "net/http"


    "blaulicht/config"
    "blaulicht/api"

    "github.com/op/go-logging"
    "github.com/gorilla/mux"
    "github.com/gorilla/handlers"
    "github.com/tarm/serial"
}

var log = logging.MustGetLogger("blaulicht")
var format = logging.MustStringFormatter(
    `%{color}%{time:15:04:05.000} %{level:-8s} ▶ %{shortpkg:-10s} ▶%{color:reset} %{message}`,
)

var (
  conf config.Config
)

func main() {
    var err error

    logBackend := logging.NewLogBackend(os.Stderr, "", 0)
    logBackendFormatter := logging.NewBackendFormatter(logBackend, format)
    logBackendLeveled := logging.AddModuleLevel(logBackendFormatter)
    logBackendLeveled.SetLevel(logging.DEBUG, "")
    logging.SetBackend(logBackendLeveled)

    //load configuration
    conf, err = config.GetConfig(log, os.Getenv("CONFIG_PATH_PREFIX") + "config.yml")
    if err != nil {
      log.Error(err.Error())
      os.Exit(1)
    }
    api.Conf = conf

    //open serial port
    serialPort, err := serial.OpenPort(&serial.Config{Name: conf.SerialPortDevice, Baud: conf.SerialPortBaudRate})
    if err != nil {
      log.Errorf("Unable to open serial port: %v", err.Error())
      os.Exit(2)
    }
    defer serialPort.Close()

    //set up http router
    router := mux.NewRouter()
    router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {w.Write([]byte("Nothing here yet"))})

    router.HandleFunc("/api/v1/auth/login", api.AuthLogin)
    router.Handle("/api/v1/alert", api.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      api.Alert(w, r, serialPort)
    }))).Methods("POST", "OPTIONS")

    log.Info("Listening on", conf.ListenAddress)

    headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
    originsOk := handlers.AllowedOrigins([]string{"*"})
    methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

    log.Error(http.ListenAndServe(conf.ListenAddress, handlers.CORS(headersOk, originsOk, methodsOk)(router)))
}
