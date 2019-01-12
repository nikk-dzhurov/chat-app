import RestClient from './rest-client';

export default class UserClient extends RestClient {
	list() {
		return this.doRequest('GET', 'users', null, true);
	}

	listActiveUserIds() {
		return this.doRequest('GET', 'users/active', null, true)
			.catch(err => {
				console.log('Failed to list active users ids: ', err);

				return {};
			});
	}

	login(data) {
		return this.doRequest('POST', 'login', data);
	}

	get(userId) {
		return this.doRequest('GET', `user/${userId}`, null, true);
	}

	update(userId, data) {
		return this.doRequest('PUT', `user/${userId}`, data, true);
	}

	register(data) {
		return this.doRequest('POST', 'register', data);
	}

	logout() {
		return this.doRequest('POST', 'logout', null, true);
	}

	uploadAvatar(userId, blob) {
		return this.uploadBlob('POST', `user/${userId}/avatar`, blob, true);
	}

	getAvatar(userId) {
		return this.downloadBlob('GET', `user/${userId}/avatar`, true)
			.catch(err => {
				console.log('get-blob-error: ', err);

				return null;
			});
	}
}
