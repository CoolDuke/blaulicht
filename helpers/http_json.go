package helpers

import (
	"encoding/json"
	"net/http"
)

func HttpRespond(w http.ResponseWriter, httpStatus int, key string, message string) {
  data := map[string]interface{} {key : message}
	HttpSendResponse(w, httpStatus, data)
}

func HttpError(w http.ResponseWriter, httpStatus int, errorKey string, errorMessage string)  {
  data := map[string]interface{} {"error" : errorKey, "message" : errorMessage}
  HttpSendResponse(w, httpStatus, data)
}

func HttpRespondObject(w http.ResponseWriter, httpStatus int, message string, object interface{}) {
  data := map[string]interface{} {"message" : message, "object" :  object}
	HttpSendResponse(w, httpStatus, data)
}

func HttpSendResponse(w http.ResponseWriter, httpStatus int, data map[string]interface{}) {
  data["success"] = 0
	if httpStatus < 400 {
	  data["success"] = 1
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(data)
}
