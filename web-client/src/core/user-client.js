import RestClient from './rest-client';

export default class UserClient extends RestClient {
	list() {
		return this.doRequest('GET', 'users', null, true);
	}

	login(data) {
		return this.doRequest('POST', 'login', data);
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
