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

const IntServErr = "Internal Server Error"
const BadRequestErr = "Bad Request"
const NotFoundErr = "Not Found"

type apiController struct {
	store *dbcontroller.Store
}

type UserWithToken struct {
	model.PublicUser
	AccessToken string `json:"accessToken"`
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

func (c *apiController) register(w http.ResponseWriter, r *http.Request) {
	usernameRe := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	passwordRe := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	user := model.User{}
	err := c.readData(r.Body, &user)
	if err != nil {
		log.Println(err)
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Invalid Request"})
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
		c.writeResponse(w, http.StatusInternalServerError, ErrorMessage{IntServErr})
		return
	} else if exists {
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Already Registered"})
		return
	}

	err = c.store.UserRepo.Create(&user)
	if err != nil {
		c.writeResponse(w, http.StatusInternalServerError, ErrorMessage{IntServErr})
		return
	}

	token := model.AccessToken{UserID: user.ID}
	err = c.store.TokenRepo.Create(&token)
	if err != nil {
		c.writeResponse(w, http.StatusInternalServerError, ErrorMessage{IntServErr})
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
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{BadRequestErr})
		return
	}

	if user.Username == "" || user.Password == "" {
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{BadRequestErr})
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
		c.writeResponse(w, http.StatusInternalServerError, ErrorMessage{IntServErr})
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
		c.writeResponse(w, http.StatusUnauthorized, ErrorMessage{"Unauthorized"})
		return
	}

	chat := model.Chat{}
	err = c.readData(r.Body, &chat)
	if err != nil {
		log.Println(err)
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{BadRequestErr})
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
				c.writeResponse(w, http.StatusBadRequest, ErrorMessage{BadRequestErr})
				return
			}
		}
	}

	chat.CreatorID = currentUserID
	err = c.store.ChatRepo.Create(&chat)
	if err != nil {
		c.writeResponse(w, http.StatusInternalServerError, ErrorMessage{IntServErr})
		return
	}

	currChatUser := model.ChatUser{
		ChatID: chat.ID,
		UserID: currentUserID,
	}
	err = c.store.ChatUserRepo.Create(&currChatUser)
	if err != nil {
		c.writeResponse(w, http.StatusInternalServerError, ErrorMessage{IntServErr})
		return
	}

	if chat.DirectUserID != "" {
		directChatUser := model.ChatUser{
			ChatID: chat.ID,
			UserID: chat.DirectUserID,
		}
		err := c.store.ChatUserRepo.Create(&directChatUser)
		if err != nil {
			c.writeResponse(w, http.StatusInternalServerError, ErrorMessage{IntServErr})
			return
		}
	}

	c.writeResponse(w, http.StatusOK, chat)
}

func (c *apiController) getChat(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeResponse(w, http.StatusUnauthorized, ErrorMessage{"Unauthorized"})
		return
	}

	vars := mux.Vars(r)
	if vars["chatID"] == "" {
		c.writeResponse(w, http.StatusBadRequest, ErrorMessage{BadRequestErr})
		return
	}

	exists, err := c.store.ChatUserRepo.Exists(vars["chatID"], currentUserID)
	if err != nil {
		c.writeResponse(w, http.StatusUnauthorized, ErrorMessage{"Unauthorized"})
		return
	}

	if !exists {
		c.writeResponse(w, http.StatusNotFound, ErrorMessage{NotFoundErr})
		return
	}

	chat := model.Chat{}
	err = c.store.ChatRepo.Get(vars["chatID"], &chat)
	if err != nil {
		c.writeResponse(w, http.StatusInternalServerError, ErrorMessage{IntServErr})
		return
	}

	c.writeResponse(w, http.StatusOK, chat)
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
