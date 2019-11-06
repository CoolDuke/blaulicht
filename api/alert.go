package api

import (
  "time"
  "net/http"
  "encoding/json"

  "github.com/freenetgigital/blaulicht/config"
  "github.com/freenetgigital/blaulicht/helpers"

  "github.com/op/go-logging"
  "github.com/prometheus/alertmanager/template"
)

var log = logging.MustGetLogger("blaulicht")
var Conf config.Config
var lastAlert = time.Now()

func Alert(w http.ResponseWriter, r *http.Request, serialPort *helpers.SerialPort) {
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

      if time.Now().After(lastAlert.Add(time.Second * Conf.AlertSilence)) {
        log.Info("Alert triggerd")
        lastAlert = time.Now()

        err := serialPort.SendCommand("A1")
        if err != nil {
          log.Error("Error while enabling Blaulicht: " + err.Error())
          helpers.HttpError(w, http.StatusInternalServerError, "alert", err.Error())
          return
        }
        log.Info("Blaulicht enabled")

        //schedule turning it off again
        time.AfterFunc(Conf.AlertDuration * time.Second, func() {
          err := serialPort.SendCommand("A0")
          if err != nil {
            log.Error("Error while disabling Blaulicht :" + err.Error())
            return
          }
          log.Info("Blaulicht disabled")
        })
        helpers.HttpRespondObject(w, http.StatusCreated, "Blaulicht started", data)
      }else{
        log.Info("Alert silenced")

        lastAlert = time.Now()
        helpers.HttpRespondObject(w, http.StatusCreated, "Blaulicht silenced", data)
      }
      return
    }
  }

  //return received alert
  helpers.HttpRespondObject(w, http.StatusOK, "No action taken", data)
}
