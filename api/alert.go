package api

import (
  "net/http"
  "encoding/json"

  "blaulicht/config"
  "blaulicht/helpers"

  "github.com/op/go-logging"
  "github.com/tarm/serial"
  "github.com/prometheus/alertmanager/template"
)

var log = logging.MustGetLogger("blaulicht")
var Conf config.Config

func Alert(w http.ResponseWriter, r *http.Request, RegisterAdapter *serial.Port) {
  data := template.Data{}
  if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
    log.Error("Unable to parse incoming json")
    http.Error(w, "Unable to parse incoming json: " + err.Error(), http.StatusBadRequest)
    return
  }

  log.Debugf("Alerts: GroupLabels=%v, CommonLabels=%v", data.GroupLabels, data.CommonLabels)
  for _, alert := range data.Alerts {
    if alert.Labels["severity"] == "CRITICAL" {
      log.Debugf("Alert: status=%s,Labels=%v,Annotations=%v", alert.Status, alert.Labels, alert.Annotations)
    }
  }

  //return received alert
  helpers.HttpRespondObject(w, http.StatusCreated, "Alert received", data)
}
