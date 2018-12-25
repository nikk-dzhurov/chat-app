import RestClient from './rest-client';

export default class MessageClient extends RestClient {
	create(chatId, data) {
		return this.doRequest('POST', `chat/${chatId}/message`, data);
	}

	update(chatId, messageId, data) {
		return this.doRequest('POST', `chat/${chatId}/message/${messageId}`, data);
	}

	delete(chatId, messageId) {
		return this.doRequest('POST', `chat/${chatId}/message/${messageId}`);
	}
}
