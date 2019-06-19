package api

import (
//  "regexp"
  "net/http"
//  "net/url"
//  "encoding/json"

  "blaulicht/types"
  "blaulicht/config"
  "blaulicht/helpers"

  "github.com/op/go-logging"
//  "github.com/gorilla/mux"
  "github.com/tarm/serial"

//  "github.com/kr/pretty"
)

var log = logging.MustGetLogger("blaulicht")
var Conf config.Config

func Alert(w http.ResponseWriter, r *http.Request, RegisterAdapter *serial.Port) {
/*
  params := mux.Vars(r)

  err := json.NewDecoder(r.Body).Decode(&types.Alert)
  if err != nil {
    log.Error("Unable to parse incoming json")
    http.Error(w, "Unable to parse incoming json: " + err.Error(), http.StatusBadRequest)
    return
  }

  //validate input
  regex, _ := regexp.Compile("^[a-z0-9-]+$")
  if !regex.MatchString(alert.Name) {
    errStr := "Invalid alert name given"
    log.Error(errStr)
    helpers.HttpError(w, http.StatusBadRequest, "input", errStr)
    return
  }
*/

  alert := new(types.Alert);

  //return sent alert
  helpers.HttpRespondObject(w, http.StatusCreated, "Alert sent", alert)
}
