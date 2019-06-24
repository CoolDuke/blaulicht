package api

import (
  "time"
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

func Alert(w http.ResponseWriter, r *http.Request, serialPort *serial.Port) {
  data := template.Data{}
  if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
    log.Error("Unable to parse incoming json")
    http.Error(w, "Unable to parse incoming json: " + err.Error(), http.StatusBadRequest)
    return
  }

  log.Debugf("Received alerts: GroupLabels=%v, CommonLabels=%v", data.GroupLabels, data.CommonLabels)
  for _, alert := range data.Alerts {
    if alert.Labels["severity"] == "critical" && alert.Status == "firing" {
      log.Debugf("Received critical alert: Status=%s,Labels=%v,Annotations=%v", alert.Status, alert.Labels, alert.Annotations)

      //TODO: put into submodule keeping track of the status
      serialPort.Write([]byte("A1\r\n")) //TODO: read response
      log.Info("Blaulicht enabled")

      time.AfterFunc(10 * time.Second, func() {
        serialPort.Write([]byte("A0\r\n")) //TODO: read response
        log.Info("Blaulicht disabled")
      })

      helpers.HttpRespondObject(w, http.StatusCreated, "Blaulicht started", data)
      return
    }
  }

  //return received alert
  helpers.HttpRespondObject(w, http.StatusOK, "No action taken", data)
}
