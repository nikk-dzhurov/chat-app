import React from 'react';
import PropTypes from 'prop-types';
import { HashRouter as Router, Route } from 'react-router-dom';
import {MuiThemeProvider, createMuiTheme} from '@material-ui/core/styles';

import ChatPage from './pages/chat';
import UserPage from './pages/user';
import LoginPage from './pages/login';

import container from 'container';
import LoadingIndication from './components/loading-indication';

const themes = {
	light: createMuiTheme({
		palette: {
			type: 'light',
		},
		typography: {
			useNextVariants: true,
		},
	}),
	dark: createMuiTheme({
		palette: {
			type: 'dark',
		},
		typography: {
			useNextVariants: true,
		},
	}),
};

export default class App extends React.Component {
	constructor(props) {
		super(props);

		this.state = {
			loading: true,
			themeKey: 'light',
			currentUser: null,
			checked: false,
			usersMap: {},
		};

		this.logout = this.logout.bind(this);
		this.setCurrentUser = this.setCurrentUser.bind(this);
		this.clearCurrentUser = this.clearCurrentUser.bind(this);

		container.init(this.clearCurrentUser);
		this.userClient = container.get('userClient');
		this.chatClient = container.get('chatClient');

	}

	componentDidMount() {
		document.title = 'Chat App';
		let data = window.localStorage.getItem('user');
		let user = null;

		if (data) {
			try {
				user = JSON.parse(data);
			} catch (ex) {
				console.error('invalid user data');
			}
		}

		let state = {
			currentUser: user,
			loading: false,
		};

		if (user) {
			this.loadUserData();
			state.loading = true;
		}

		this.setState(state);
	}

	getChildContext() {
		return {
			currentUser: this.state.currentUser,
			usersMap: this.state.usersMap,
			logout: this.logout,
		};
	}

	clearCurrentUser() {
		this.setCurrentUser(null);
	}

	setCurrentUser(user) {
		if (user) {
			window.localStorage.setItem('user', JSON.stringify(user));
			this.setState({
				currentUser: user,
			});
		} else {
			window.localStorage.removeItem('user');
			this.setState({
				currentUser: null,
			});
		}

		this.loadUserData();
	}

	logout() {
		this.setState({loading: true});
		this.userClient.logout()
			.then(() => this.setCurrentUser(null))
			.catch(err => {
				console.error(err);

				this.setCurrentUser(null);
			});
	}

	async loadUserData() {
		let users = await this.userClient.list()
			.then(data => (data || []))
			.catch(err => {
				console.log(err);

				return [];
			});

		let usersMap = {};
		for (let u of users) {
			let blob = await this.userClient.getAvatar(u.id);
			let blobUrl = null;
			if (blob) {
				blobUrl = URL.createObjectURL(blob);
			}

			usersMap[u.id] = {...u, blob, blobUrl};
		}

		this.setState({
			usersMap,
			loading: false,
		});
	}

	render() {
		return (
			<MuiThemeProvider theme={themes[this.state.themeKey]}>
				{this.state.loading ?
					<LoadingIndication />
					:
					<Router basename='/'>
						{this.state.currentUser ?
							<React.Fragment>
								<Route exact path='/' component={ChatPage} />
								<Route path='/profile' render={(props) => (
									<UserPage
										{...props}
										logout={this.logout}
									/>
								)} />
							</React.Fragment>
							:
							<React.Fragment>
								<Route path='/' render={(props) => (
									<LoginPage
										{...props}
										setCurrentUser={this.setCurrentUser}
									/>
								)} />
							</React.Fragment>
						}
					</Router>
				}
			</MuiThemeProvider>
		);
	}
}
App.childContextTypes = {
	currentUser: PropTypes.object,
	usersMap: PropTypes.object,
	logout: PropTypes.func,
};

// const Bar = (props, {muiTheme}) => {
// 	return (
// 		<div style={{background: props.theme.palette.background}}>
// 			Choose theme:<br />
// 			Dark <Switch checked={props.checked} onChange={props.onChange} /> Light
// 		</div>
// 	);
// };
//
// const BarT = withTheme()(Bar);
