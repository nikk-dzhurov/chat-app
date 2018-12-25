export default class RestClient {
	constructor(baseUrl, handleUnauthorized) {
		this.handleUnauthorized = handleUnauthorized;
		this.cfg = {
			baseUrl,
		};
	}

	getToken() {
		let userData = window.localStorage.getItem('user');

		let token = null;
		if (userData) {
			try {
				let user = JSON.parse(userData);
				token = user.accessToken;
			} catch (ex) {
				console.error(ex);
			}
		}

		return token;
	}

	async uploadBlob(method, path, blob, applyAuthorizationHeaders) {
		let url = `${this.cfg.baseUrl}${path}`;
		let req = {
			method,
			headers: {
				'Content-Type': blob.type,
			},
			body: blob,
		};

		return this._doR(url, req, applyAuthorizationHeaders);
	}

	async downloadBlob(method, path, applyAuthorizationHeaders) {
		let url = `${this.cfg.baseUrl}${path}`;
		let req = {
			method,
			headers: {},
		};

		if (applyAuthorizationHeaders) {
			let token = await this.getToken();
			if (!token) {
				return Promise.reject('Missing access token');
			}

			req.headers['Authorization'] = `Bearer ${token}`;
		}

		return fetch(url, req)
			.then(response => {
				if (response.ok) {
					return response.blob();
				}

				console.log('err status', response.status);

				if (response.status === 401) {
					return this.handleUnauthorized();
				}

				return response.json()
					.then(Promise.reject);
			})
			.catch(err => {
				if (err && err.message) {
					return Promise.reject(err);
				}

				return Promise.reject({message: err});
			});
	}

	doRequest(method, path, data, applyAuthorizationHeaders) {
		let url = `${this.cfg.baseUrl}${path}`;
		let req = {
			method,
			headers: {},
		};

		if (data) {
			req.body = JSON.stringify(data);
			req.headers['Content-Type'] = 'application/json';
			req.headers['Accept'] = 'application/json';
		}

		return this._doR(url, req, applyAuthorizationHeaders);
	}

	async _doR(url, request, applyAuthorizationHeaders) {
		if (applyAuthorizationHeaders) {
			let token = await this.getToken();
			if (!token) {
				return Promise.reject('Missing access token');
			}

			request.headers['Authorization'] = `Bearer ${token}`;
		}

		return fetch(url, request)
			.then(response => {
				if (response.ok) {
					return response.json();
				}

				console.log('err status', response.status);

				if (response.status === 401) {
					return this.handleUnauthorized();
				}

				return response.json()
					.then(Promise.reject);
			})
			.catch(err => {
				if (err && err.message) {
					return Promise.reject(err);
				}

				return Promise.reject({message: err});
			});
	}
}
