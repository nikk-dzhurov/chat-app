import RestClient from './rest-client';

export default class MessageClient extends RestClient {
	list(chatId) {
		return this.doRequest('GET', `chat/${chatId}/messages`, null, true);
	}

	create(chatId, data) {
		return this.doRequest('POST', `chat/${chatId}/message`, data, true);
	}

	get(chatId, messageId) {
		return this.doRequest('GET', `chat/${chatId}/message/${messageId}`, null, true);
	}

	update(chatId, messageId, data) {
		return this.doRequest('PUT', `chat/${chatId}/message/${messageId}`, data, true);
	}

	delete(chatId, messageId) {
		return this.doRequest('DELETE', `chat/${chatId}/message/${messageId}`, null, true);
	}
}
