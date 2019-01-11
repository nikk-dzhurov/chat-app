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
	"time"

	"./dbcontroller"
	"./model"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// Default error response messages
const (
	IntServErr      = "Internal Server Error"
	BadRequestErr   = "Bad Request"
	NotFoundErr     = "Not Found"
	UnauthorizedErr = "Unauthorized"
	ForbiddenErr    = "Forbidden"
)

// WebSocket messages types
const (
	WSTypeMessageCreate = "message_create"
	WSTypeMessageUpdate = "message_update"
	WSTypeMessageDelete = "message_delete"

	WSTypeChatCreate = "chat_create"
	WSTypeChatUpdate = "chat_update"
	WSTypeChatDelete = "chat_delete"

	WSTypeUserCreate = "user_create"
	WSTypeUserUpdate = "user_update"
	WSTypeUserDelete = "user_delete"
	WSTypeUserAvatarUpdate = "user_avatar_update"
	WSTypeUserStatusChange = "user_status_change"
)

const MAX_BLOB_SIZE = 1024 * 1024 * 15

var PERMITTED_AVATAR_CONTENT_TYPES = []string{"image/jpeg", "image/png"}

type apiController struct {
	store *dbcontroller.Store
	wsHub *WSHub
}

type UserWithToken struct {
	model.PublicUser
	AccessToken          string     `json:"accessToken"`
	AccessTokenExpiresAt *time.Time `json:"accessTokenExpiresAt"`
}

type ErrorMessage struct {
	Message string `json:"error"`
}

type WSMessageData struct {
	Type      string `json:"type"`
	ChatID    string `json:"chatId"`
	MessageID string `json:"messageId"`
}

type WSUserData struct {
	Type   string `json:"type"`
	UserID string `json:"userId"`
}

type WSChatData struct {
	Type   string `json:"type"`
	ChatID string `json:"chatId"`
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

func (c *apiController) wsHandler(w http.ResponseWriter, r *http.Request) {
	protocol := r.Header.Get("Sec-WebSocket-Protocol")
	arr := strings.Split(protocol, ", ")
	if len(arr) != 2 || arr[0] != "access_token" {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	token, err := c.validateAccessToken(arr[1])
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	conn, err := c.wsHub.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &WsClient{
		hub:         c.wsHub,
		conn:        conn,
		send:        make(chan []byte, 256),
		userID:      token.UserID,
		accessToken: token,
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	// go client.readPump()
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

	user.FullName = ""

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
		PublicUser:           user.PublicUser,
		AccessToken:          token.Token,
		AccessTokenExpiresAt: token.ExpiresAt,
	}

	c.broadcastUserChange(user.ID, WSTypeUserCreate)

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
		PublicUser:           dbUser.PublicUser,
		AccessToken:          token.Token,
		AccessTokenExpiresAt: token.ExpiresAt,
	}

	c.writeResponse(w, http.StatusOK, result)
}

func (c *apiController) logout(w http.ResponseWriter, r *http.Request) {

	token, err := c.authenticateWithToken(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	err = c.store.TokenRepo.Delete(token)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusNoContent, nil)
}

func (c *apiController) updateUser(w http.ResponseWriter, r *http.Request) {

	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	user := model.User{}
	err = c.readData(r.Body, &user)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	// validate userdata
	{
		if user.ID != currentUserID || len(user.FullName) > 255 {
			c.writeDefaultErrorResponse(w, http.StatusBadRequest)
			return
		}
	}

	err = c.store.UserRepo.Update(&user)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.broadcastUserChange(currentUserID, WSTypeUserUpdate)

	c.writeResponse(w, http.StatusOK, user.PublicUser)
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

			var exists bool
			exists, err = c.store.ChatRepo.DirectChatExists(currentUserID, chat.DirectUserID)
			if err != nil {
				log.Println(err)
				c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
				return
			}

			if exists {
				c.writeResponse(w, http.StatusBadRequest, ErrorMessage{"Direct chat already exists"})
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

	c.broadcastChatChange(chat.ID, WSTypeChatCreate)

	c.writeResponse(w, http.StatusCreated, chat)
}

func (c *apiController) listUsers(w http.ResponseWriter, r *http.Request) {
	_, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	users := []model.User{}
	err = c.store.UserRepo.List(&users)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	publicUsers := make([]model.PublicUser, len(users))
	for i := range users {
		publicUsers[i] = users[i].PublicUser
	}

	c.writeResponse(w, http.StatusOK, publicUsers)
}

func (c *apiController) getUser(w http.ResponseWriter, r *http.Request) {
	_, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	if vars["userID"] == "" {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	user := model.User{}
	err = c.store.UserRepo.Get(vars["userID"], &user)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.writeDefaultErrorResponse(w, http.StatusNotFound)
			return
		}

		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.writeResponse(w, http.StatusOK, user.PublicUser)
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

func (c *apiController) isContentTypePermitted(ct string) bool {
	for _, pct := range PERMITTED_AVATAR_CONTENT_TYPES {
		if ct == pct {
			return true
		}
	}

	return false
}

func (c *apiController) getAvatar(w http.ResponseWriter, r *http.Request) {
	_, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	if vars["userID"] == "" {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	exists, err := c.store.UserRepo.HasAvatar(vars["userID"])
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	if !exists {
		c.writeDefaultErrorResponse(w, http.StatusNotFound)
		return
	}

	avatar := model.UserAvatar{}
	err = c.store.UserRepo.GetAvatar(vars["userID"], &avatar)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", avatar.ContentType)
	w.Write(avatar.Blob)
}

func (c *apiController) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := c.authenticate(r)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	if vars["userID"] == "" || currentUserID != vars["userID"] {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	avatar, err := ioutil.ReadAll(r.Body)
	contentType := r.Header.Get("Content-Type")
	if err != nil || len(avatar) == 0 || len(avatar) > MAX_BLOB_SIZE || len(contentType) == 0 || !c.isContentTypePermitted(contentType) {
		c.writeDefaultErrorResponse(w, http.StatusBadRequest)
		return
	}

	exists, err := c.store.UserRepo.HasAvatar(currentUserID)
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	userAvatar := model.UserAvatar{
		UserID:      currentUserID,
		ContentType: contentType,
		Blob:        avatar,
	}

	if !exists {
		err = c.store.UserRepo.CreateAvatar(&userAvatar)
	} else {
		err = c.store.UserRepo.UpdateAvatar(&userAvatar)
	}
	if err != nil {
		c.writeDefaultErrorResponse(w, http.StatusInternalServerError)
		return
	}

	c.store.UserRepo.UpdateUpdatedAt(userAvatar.UserID, nil)

	c.broadcastUserChange(currentUserID, WSTypeUserAvatarUpdate)

	c.writeResponse(w, http.StatusNoContent, nil)
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

	c.broadcastChatChange(chat.ID, WSTypeChatUpdate)

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

	c.broadcastChatChange(vars["chatID"], WSTypeChatDelete)

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

	c.store.ChatRepo.UpdateUpdatedAt(msg.ChatID, msg.UpdatedAt)

	c.broadcastMessageChange(&msg, WSTypeMessageCreate)

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

	c.store.ChatRepo.UpdateUpdatedAt(msg.ChatID, msg.UpdatedAt)

	c.broadcastMessageChange(&msg, WSTypeMessageUpdate)

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

	c.store.ChatRepo.UpdateUpdatedAt(msg.ChatID, nil)


	c.broadcastMessageChange(&msg, WSTypeMessageDelete)

	c.writeResponse(w, http.StatusNoContent, nil)
}

func (c *apiController) authenticateWithToken(req *http.Request) (*model.AccessToken, error) {
	tokenString := ""
	bearerToken := req.Header.Get("Authorization")
	parts := strings.Split(bearerToken, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		tokenString = parts[1]
	}

	return c.validateAccessToken(tokenString)
}

func (c *apiController) validateAccessToken(tokenString string) (*model.AccessToken, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("Access token is missing")
	}

	token, err := c.store.TokenRepo.Get(tokenString)
	if err != nil {
		log.Println("Failed to get access token from store: ", err)
		return nil, fmt.Errorf("Access token is invalid")
	}

	if !token.IsValid() {
		c.store.TokenRepo.Delete(token)
		return nil, fmt.Errorf("Access token is expired")
	}

	return token, nil
}

func (c *apiController) authenticate(req *http.Request) (string, error) {
	token, err := c.authenticateWithToken(req)
	if err != nil {
		return "", err
	}

	return token.UserID, nil
}

func (c *apiController) broadcastMessageChange(msg *model.Message, messageType string) {
	chatUsers := []model.ChatUser{}
	err := c.store.ChatUserRepo.ListByChatID(msg.ChatID, &chatUsers)
	if err == nil && len(chatUsers) > 0 {
		userIDs := []string{}
		for i := range chatUsers {
			userIDs = append(userIDs, chatUsers[i].UserID)
		}

		c.wsHub.broadcastData(userIDs, &WSMessageData{
			Type:      messageType,
			MessageID: msg.ID,
			ChatID:    msg.ChatID,
		})
	}
}

func (c *apiController) broadcastChatChange(chatID string, messageType string) {
	chatUsers := []model.ChatUser{}
	err := c.store.ChatUserRepo.ListByChatID(chatID, &chatUsers)
	if err == nil && len(chatUsers) > 0 {
		userIDs := []string{}
		for i := range chatUsers {
			userIDs = append(userIDs, chatUsers[i].UserID)
		}

		c.wsHub.broadcastData(userIDs, &WSMessageData{
			Type:   messageType,
			ChatID: chatID,
		})
	}
}

func (c *apiController) broadcastUserChange(userID string, messageType string) {
	c.wsHub.broadcastDataToAll(&WSUserData{
		Type:   messageType,
		UserID: userID,
	})
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

	w.Header().Set("Content-Type", "application/json")
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
