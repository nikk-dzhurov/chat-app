import RestClient from './rest-client';

export default class ChatClient extends RestClient {
	list() {
		return this.doRequest('GET', 'chats', null, true);
	}

	create(data) {
		return this.doRequest('POST', 'chat', data, true);
	}

	update(chatId, data) {
		return this.doRequest('POST', `chat/${chatId}`, data, true);
	}

	delete(chatId) {
		return this.doRequest('POST', `chat/${chatId}`, null, true);
	}
}
