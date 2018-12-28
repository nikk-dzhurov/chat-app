import UserClient from './core/user-client';
import ChatClient from './core/chat-client';
import MessageClient from './core/message-client';
import WsClient from './core/ws-client';

// const API_HOST = 'localhost:3000';
const API_HOST = `${window.location.hostname}:3000`;
const API_URL = `http://${API_HOST}/`;
const container = {};

export default {
	init(handleUnauthorizedCallback) {
		container['userClient'] = new UserClient(API_URL, handleUnauthorizedCallback);
		container['chatClient'] = new ChatClient(API_URL, handleUnauthorizedCallback);
		container['messageClient'] = new MessageClient(API_URL, handleUnauthorizedCallback);
		container['wsClient'] = new WsClient(API_HOST);
	},
	get(key) {
		return container[key];
	},
};
