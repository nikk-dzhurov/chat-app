import UserClient from './core/user-client';
import ChatClient from './core/chat-client';
import MessageClient from './core/message-client';

const API_URL = 'http://localhost:3000/';
const container = {};

export default {
	init(handleUnauthorizedCallback) {
		container['userClient'] = new UserClient(API_URL, handleUnauthorizedCallback);
		container['chatClient'] = new ChatClient(API_URL, handleUnauthorizedCallback);
		container['messageClient'] = new MessageClient(API_URL, handleUnauthorizedCallback);
	},
	get(key) {
		return container[key];
	},
};
