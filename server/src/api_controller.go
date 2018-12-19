package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"./dbcontroller"
	"./model"
	"github.com/gorilla/mux"
)

// Default error response messages
const IntServErr = "Internal Server Error"
const BadRequestErr = "Bad Request"
const NotFoundErr = "Not Found"
const UnauthorizedErr = "Unauthorized"
const ForbiddenErr = "Forbidden"

type apiController struct {
	store *dbcontroller.Store
}

type UserWithToken struct {
	model.PublicUser
	AccessToken string `json:"accessToken"`
}

type ErrorMessage struct {
	Message string `json:"errorMessage"`
}

func (c *apiController) readData(data io.Reader, result interface{}) error {
	return parseJSONData(data, result)
}

func (c *apiController) writeResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	writeJSONResponse(w, statusCode, data)
}

func (c *apiController) writeDefaultErrorResponse(w http.ResponseWriter, statusCode int) {
	var data interface{}

	switch statusCode {
	case http.StatusBadRequest:
		data = ErrorMessage{BadRequestErr}
		break
	case http.StatusUnauthorized:
		data = ErrorMessage{UnauthorizedErr}
		break
	case http.StatusForbidden:
		data = ErrorMessage{ForbiddenErr}
		break
	case http.StatusNotFound:
		data = ErrorMessage{NotFoundErr}
		break
	case http.StatusInternalServerError:
		data = ErrorMessage{IntServErr}
		break
	default:
		data = nil
	}

	writeJSONResponse(w, statusCode, data)
}

func (c *apiController) register(w http.ResponseWriter, r *http.Request) {
	usernameRe := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	passwordRe := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	user := model.User{}
	err := c.readData(r.Body, &user)
	if err != nil {
		log.Println(err)
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	// Validate user
	{
		nameLen := len(user.Username)
		if nameLen < 4 || nameLen >= 256 || !usernameRe.MatchString(user.Username) {
			c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Username is not valid"})
			return
		}

		passLen := len(user.Password)
		if passLen < 6 || passLen >= 256 || !passwordRe.MatchString(user.Password) {
			c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Password is not valid"})
			return
		}

		fullNameLen := len(user.FullName)
		if fullNameLen >= 256 {
			c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"FullName is not valid"})
			return
		}
	}

	// check if already registered
	exists, err := c.store.UserRepo.ExistsUsername(user.Username)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	} else if exists {
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Already Registered"})
		return
	}

	err = c.store.UserRepo.Create(&user)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	token := model.AccessToken{UserID: user.ID}
	err = c.store.TokenRepo.Create(&token)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	result := &UserWithToken{
		PublicUser:  user.PublicUser,
		AccessToken: token.Token,
	}

	c.writeResponse(w, http.StatusOK, result)
}

func (c *apiController) login(w http.ResponseWriter, r *http.Request) {
	user := model.User{}
	err := c.readData(r.Body, &user)
	if err != nil {
		log.Println(err)
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	dbUser, err := c.store.UserRepo.GetByUsername(user.Username)
	if err != nil {
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Invalid username or password"})
		return
	}

	if !c.store.UserRepo.VerifyPassword(dbUser.PasswordHash, user.Password) {
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Invalid username or password"})
		return
	}

	token := model.AccessToken{UserID: dbUser.ID}
	err = c.store.TokenRepo.Create(&token)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	result := &UserWithToken{
		PublicUser:  dbUser.PublicUser,
		AccessToken: token.Token,
	}

	c.writeResponse(w, http.StatusOK, result)
}

func (c *apiController) createChat(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	chat := model.Chat{}
	err = c.readData(r.Body, &chat)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	// Validate chat data
	{
		titleLen := len(chat.Title)
		if titleLen >= 256 {
			c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Title is not valid"})
			return
		}

		if chat.DirectUserID != "" {
			du := model.User{}
			err = c.store.UserRepo.Get(chat.DirectUserID, &du)
			if err != nil {
				log.Printf("User with id %s is not found\n", chat.DirectUserID)
				c.writeDefaultErrorResponse(w, http.StatusBadRequest)
				return
			}
		}
	}

	chat.CreatorID = currentUserID
	err = c.store.ChatRepo.Create(&chat)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	currChatUser := model.ChatUser{
		ChatID: chat.ID,
		UserID: currentUserID,
	}
	err = c.store.ChatUserRepo.Create(&currChatUser)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	if chat.DirectUserID != "" {
		directChatUser := model.ChatUser{
			ChatID: chat.ID,
			UserID: chat.DirectUserID,
		}
		err := c.store.ChatUserRepo.Create(&directChatUser)
		if err != nil {
			c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
			return
		}
	}

	c.writeResponse(w, http.StatusCreated, chat)
}

func (c *apiController) listChats(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	chats := []model.Chat{}
	err = c.store.ChatRepo.ListByUserID(currentUserID, &chats)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusOK, chats)
}

func (c *apiController) getChat(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	if vars["chatID"] == "" {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	exists, err := c.store.ChatUserRepo.Exists(vars["chatID"], currentUserID)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusForbidden)
		return
	}

	if !exists {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	chat := model.Chat{}
	err = c.store.ChatRepo.Get(vars["chatID"], &chat)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusOK, chat)
}

func (c *apiController) updateChat(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	chat := model.Chat{}
	err = c.readData(r.Body, &chat)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	if vars["chatID"] == "" || vars["chatID"] != chat.ID {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	// validate chat data
	{
		titleLen := len(chat.Title)
		if titleLen >= 256 {
			c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Title is not valid"})
			return
		}
	}

	exists, err := c.store.ChatUserRepo.Exists(vars["chatID"], currentUserID)
	if err != nil || !exists {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	err = c.store.ChatRepo.Update(&chat)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusOK, chat)
}

func (c *apiController) deleteChat(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	if vars["chatID"] == "" {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	exists, err := c.store.ChatUserRepo.Exists(vars["chatID"], currentUserID)
	if err != nil || !exists {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	chat := model.Chat{}
	err = c.store.ChatRepo.Get(vars["chatID"], &chat)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	if chat.CreatorID != currentUserID {
		c.writeDefaultErrorResponse(w, http.StatusForbidden)
		return
	}

	err = c.store.ChatUserRepo.DeleteByChatID(vars["chatID"])
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	err = c.store.ChatRepo.Delete(vars["chatID"])
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	err = c.store.MessageRepo.DeleteByChatID(vars["chatID"])
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusNoContent, nil)
}

func (c *apiController) createMessage(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	msg := model.Message{}
	err = c.readData(r.Body, &msg)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	// Validate message data
	{
		if msg.ChatID != vars["chatID"] || msg.UserID != currentUserID {
			c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Invalid message data"})
			return
		}
	}

	exists, err := c.store.ChatUserRepo.Exists(vars["chatID"], currentUserID)
	if err != nil || !exists {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	err = c.store.MessageRepo.Create(&msg)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusCreated, msg)
}

func (c *apiController) listMessages(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	if vars["chatID"] == "" {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	exists, err := c.store.ChatUserRepo.Exists(vars["chatID"], currentUserID)
	if err != nil || !exists {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	messages := []model.Message{}
	err = c.store.MessageRepo.ListByChatID(vars["chatID"], &messages)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusOK, messages)
}

func (c *apiController) getMessage(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	if vars["chatID"] == "" || vars["messageID"] == "" {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	exists, err := c.store.ChatUserRepo.Exists(vars["chatID"], currentUserID)
	if err != nil || !exists {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	msg := model.Message{}
	err = c.store.MessageRepo.Get(vars["messageID"], &msg)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	if msg.ChatID != vars["chatID"] {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	c.writeResponse(w, http.StatusOK, msg)
}

func (c *apiController) updateMessage(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	msg := model.Message{}
	err = c.readData(r.Body, &msg)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	if vars["chatID"] == "" || vars["messageID"] == "" {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	exists, err := c.store.ChatUserRepo.Exists(vars["chatID"], currentUserID)
	if err != nil || !exists {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	old := model.Message{}
	err = c.store.MessageRepo.Get(vars["messageID"], &old)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	if old.ChatID != vars["chatID"] {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	if old.UserID != currentUserID {
		c.writeDefaultErrorResponse(w, http.StatusForbidden)
		return
	}

	err = c.store.MessageRepo.Update(&msg)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusOK, msg)
}

func (c *apiController) deleteMessage(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	if vars["chatID"] == "" || vars["messageID"] == "" {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	exists, err := c.store.ChatUserRepo.Exists(vars["chatID"], currentUserID)
	if err != nil || !exists {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	msg := model.Message{}
	err = c.store.MessageRepo.Get(vars["messageID"], &msg)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	if msg.ChatID != vars["chatID"] {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	if msg.UserID != currentUserID {
		c.writeDefaultErrorResponse(w, http.StatusForbidden)
		return
	}

	err = c.store.MessageRepo.Delete(msg.ID)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusNoContent, nil)
}

func (c *apiController) authenticate(req *http.Request) (string, error) {
	tokenString := ""
	bearerToken := req.Header.Get("Authorization")
	parts := strings.Split(bearerToken, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		tokenString = parts[1]
	}

	if tokenString == "" {
		return "", fmt.Errorf("Access token is missing")
	}

	token, err := c.store.TokenRepo.Get(tokenString)
	if err != nil {
		log.Println("Failed to get access token from store: ", err)
		return "", fmt.Errorf("Access token is invalid")
	}

	if !token.IsValid() {
		c.store.TokenRepo.Delete(token)
		return "", fmt.Errorf("Access token is expired")
	}

	return token.UserID, nil
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
	w.Header().Set("Content-Type", "application/json")
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
