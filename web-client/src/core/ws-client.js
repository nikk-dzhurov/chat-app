export default class WsClient {
	constructor(apiHost) {
		this.cfg = {
			host: apiHost,
		};

		this.connection = null;

		this.changeListenersMap = {
			message: [],
			user: [],
			chat: [],
		};

		this.handleMessage = this.handleMessage.bind(this);
		this.handleClose = this.handleClose.bind(this);
	}

	addChangeListener(entityType, listener) {
		if (this.changeListenersMap[entityType] && this.changeListenersMap[entityType].indexOf(listener) === -1) {
			this.changeListenersMap[entityType].push(listener);
		} else {
			console.error('Invalid entityType/Existing listener');
		}
	}

	removeChangeListener(entityType, listener) {
		if (this.changeListenersMap[entityType]) {
			let idx = this.changeListenersMap[entityType].indexOf(listener);
			if (idx !== -1) {
				this.changeListenersMap[entityType].splice(idx, 1);
			}
		}
	}

	getToken() {
		let userData = window.localStorage.getItem('user');

		let token = null;
		if (userData) {
			try {
				let user = JSON.parse(userData);
				if (new Date() < new Date(user.accessTokenExpiresAt)) {
					token = user.accessToken;
				}
			} catch (ex) {
				console.error(ex);
			}
		}

		return token;
	}

	handleMessage(e) {
		console.log('received message:', e.data);

		let msg = null;
		try {
			msg = JSON.parse(e.data);
		} catch (ex) {
			console.log('Failed to parse WS message');
		}

		if (msg) {
			switch (msg.type) {
				case 'message_create':
				case 'message_update':
				case 'message_delete':
					for (let cb of this.changeListenersMap.message) {
						cb(msg);
					}

					break;
				case 'chat_create':
				case 'chat_update':
				case 'chat_delete':
					for (let cb of this.changeListenersMap.chat) {
						cb(msg);
					}

					break;
				case 'user_create':
				case 'user_update':
				case 'user_delete':
				case 'user_avatar_update':
				case 'user_status_change':
					for (let cb of this.changeListenersMap.user) {
						cb(msg);
					}

					break;
				default:
					console.log('Unrecognized message type: ' + msg.type);
			}
		}
	}

	handleClose(e) {
		this.connection = null;
		this.tryToReconnect();
		console.log('WS connection is closed', e);
	}

	tryToReconnect() {
		let accessToken = this.getToken();
		if (accessToken) {
			setTimeout(() => {
				console.log('retry to open WS connection');
				this.openConnection();
			}, 1000);
		}
	}

	openConnection() {
		if (this.connection) {
			// maintain only one open connection
			this.closeConnection();
		}

		let accessToken = this.getToken();
		if (!accessToken) {
			console.log('missing/expired accessToken');

			return;
		}

		this.connection = new WebSocket(`ws://${this.cfg.host}/ws`, ['access_token', accessToken]);

		this.connection.onclose = this.handleClose;
		this.connection.onmessage = this.handleMessage;
	}

	closeConnection() {
		this.connection.close();
	}
}
