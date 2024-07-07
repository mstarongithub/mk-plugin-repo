package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

type RegisterData struct {
	ProcessId string `json:"process_id"`
	Value     string `json:"value"`
}

func authRegisterStartHandler(w http.ResponseWriter, r *http.Request) {
	authLayer := AuthFromRequestContext(w, r)
	if authLayer == nil {
		return
	}
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}

	data := RegisterData{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	log.WithField("body", string(body)).Debugln("Received body")
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.WithError(err).WithField("body", string(body)).Debugln("Failed to decode body")
		http.Error(w, "body not json data", http.StatusBadRequest)
		return
	}

	processId := authLayer.RegisterStart(data.Value)
	returnData := RegisterData{
		ProcessId: processId,
	}
	returnJson, err := json.Marshal(&returnData)
	if err != nil {
		log.WithError(err).WithField("processId", processId).Warningln("Failed to encode response")
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(returnJson))
}

func authRegisterAddMailHandler(w http.ResponseWriter, r *http.Request) {
	authLayer := AuthFromRequestContext(w, r)
	if authLayer == nil {
		return
	}
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	reqData := RegisterData{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		http.Error(w, "failed to decode body into json", http.StatusBadRequest)
		return
	}

	err = authLayer.RegisterContinueMail(reqData.ProcessId, reqData.Value)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"process-id": reqData.ProcessId,
			"mail":       reqData.Value,
		}).Warningln("Failed to add mail to active registration process")
		http.Error(w, "failed to update process", http.StatusInternalServerError)
		return
	}
}

func authRegisterAddPasswordHandler(w http.ResponseWriter, r *http.Request) {
	authLayer := AuthFromRequestContext(w, r)
	if authLayer == nil {
		return
	}
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	reqData := RegisterData{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		http.Error(w, "failed to decode body into json", http.StatusBadRequest)
		return
	}

	err = authLayer.RegisterContinuePassword(reqData.ProcessId, reqData.Value)
	if err != nil {
		log.WithError(err).
			WithField("process-id", reqData.ProcessId).
			Warning("Failed to update password in registration process")
		http.Error(w, "failed to update password", http.StatusInternalServerError)
		return
	}
}

func authRegisterAddDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	authLayer := AuthFromRequestContext(w, r)
	if authLayer == nil {
		return
	}
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	reqData := RegisterData{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		http.Error(w, "failed to decode body into json", http.StatusBadRequest)
		return
	}

	err = authLayer.RegisterContinueDescription(reqData.ProcessId, reqData.Value)
	if err != nil {
		log.WithError(err).
			WithField("process-id", reqData.ProcessId).
			WithField("description", reqData.Value).
			Warning("Failed to update password in registration process")
		http.Error(w, "failed to update password", http.StatusInternalServerError)
		return
	}
}

func authRegisterFinaliseHandler(w http.ResponseWriter, r *http.Request) {
	authLayer := AuthFromRequestContext(w, r)
	if authLayer == nil {
		return
	}
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	reqData := RegisterData{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		http.Error(w, "failed to decode body into json", http.StatusBadRequest)
		return
	}

	err = authLayer.RegisterFinalise(reqData.ProcessId)
	if err != nil {
		log.WithError(err).
			WithField("process-id", reqData.ProcessId).
			Warning("Failed to finalise registration")
		http.Error(w, "failed to finalise registration", http.StatusInternalServerError)
		return
	}
}

func authRegisterCancelHandler(w http.ResponseWriter, r *http.Request) {
	authLayer := AuthFromRequestContext(w, r)
	if authLayer == nil {
		return
	}
	//log, ok := LogFromRequestContext(w, r)
	// if !ok {
	// return
	// }
	reqData := RegisterData{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		http.Error(w, "failed to decode body into json", http.StatusBadRequest)
		return
	}

	authLayer.RegisterCancel(reqData.ProcessId)
}
