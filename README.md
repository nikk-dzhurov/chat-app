# ChatApp

## Description

- This is simple chat application project. It is split to three main directories:
-- runtime - it contains project's configuration and basic setup
-- server - it contains the server side of the application
-- web-client - it contains the client side of the application
- Setup and run project
-- cd runtime
-- make up logs

## Server Tasks:

- [x] User handlers
-- [x] (POST) Login handler
-- [x] (POST) Sign up handler
-- [x] (PUT) Update user data
-- [x] (PUT) Update user avatar (this update is more like create/update)
-- [] (DELETE) Delete user - pretty optional case for this kind of project. It can be implemented if it's necessary

- [x] Message handlers
-- [x] (POST) Create message
-- [x] (GET) Get message
-- [x] (PUT) Update message
-- [x] (DELETE) Delete message
-- [x] (GET) List message

- [x] Chat handlers
-- [x] (POST) Create chat
-- [x] (GET) Get chat
-- [x] (PUT) Update chat
-- [] (DELETE) Delete chat - optional
-- [x] (GET) List chats

- [x] WebSocket handler
-- [x] Dispatch user changes
-- [x] Dispatch chat changes
-- [x] Dispatch message changes

## Client Tasks:
- [x] Side menu for easier navigation
- [x] Basic Application Bar
-- [x] Side menu button
-- [x] Simple logo
-- [x] Logged in user's name and avatar
- [x] Login form:
-- [x] Fields: Username/Password
-- [x] Simple client side validation
-- [x] Base error handling
- [x] Sign up form:
-- [x] Fields: Username/Password/Repeat Password
-- [x] Simple client side validation
-- [x] Base error handling
- [x] Profile page:
-- [x] Add/Update avatar
-- [x] Update user's data form
- [x] Chat page
-- [x] Create new direct message form
-- [x] Interactive chat list
-- [x] Interactive chat messages list
-- [x] Create new chat message
- [x] Basic REST clients for each entity
- [x] Basic WebSocket client/manager

## Additional Tasks:
- [] Add more message functionalities
-- [] Update message
-- [] Delete message
- [] Add group chat functionalities
-- [] Create group chat
-- [] Leave group chat
-- [] Add user to chat
-- [] Remove user from chat
- [] Order chats list by last change
- [] Test application with large lists for GUI/serverside bugs
