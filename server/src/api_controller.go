package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"./dbcontroller"
	"./model"
)

type apiController struct {
	store *dbcontroller.Store
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func (c *apiController) readData(data io.Reader, result interface{}) error {
	return parseJSONData(data, result)
}

func (c *apiController) writeResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	writeJSONResponse(w, statusCode, data)
}

func (c *apiController) helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Hello, World!</h1>"))
}

func (c *apiController) register(w http.ResponseWriter, r *http.Request) {
	user := model.User{}
	err := c.readData(r.Body, &user)
	if err != nil {
		log.Println(err)
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Invalid Request"})
		return
	}


	c.writeResponse(w, http.StatusOK, user)
}

func parseJSONData(data io.Reader, result interface{}) error {
	body, err := ioutil.ReadAll(data)
	if err != nil {
		return fmt.Errorf("Failed to read the data: %s", err.Error())
	}

	err = json.Unmarshal(body, result)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal the data: %s", err.Error())
	}

	return nil
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	if data == nil {
		return
	}

	jsonData, err := marshalJSONData(data)
	if err != nil {
		log.Println(err)
		return
	}

	w.Write(jsonData)
}

func marshalJSONData(data interface{}) ([]byte, error) {
	if data == nil {
		return nil, nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal the data: %s", err.Error())
	}

	return bytes, nil
}
