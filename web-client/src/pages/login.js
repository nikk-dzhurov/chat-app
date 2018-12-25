import React from 'react';
import PropTypes from 'prop-types';

import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Tab from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import {withStyles} from '@material-ui/core/styles';

import container from '../container';
import LoadingIndication from '../components/loading-indication';

const styles = theme => ({
	container: {
		display: 'flex',
		flexDirection: 'column',
		alignItems: 'center',
		justifyContent: 'center',
	},
	textField: {
		marginLeft: theme.spacing.unit,
		marginRight: theme.spacing.unit,
	},
	button: {
		margin: theme.spacing.unit,
	},
});

const BASE_INPUT_REGEX = new RegExp('^[a-zA-Z0-9_-]+$');
const BASE_HELPER_TEXT = 'You can use only a-z, A-Z, 0-9, \'-\' and \'_\' characters. Characters length: ';

class Login extends React.Component {
	constructor(props) {
		super(props);
		this.defaultErrors = {
			loading: false,
			usernameError: false,
			passwordError: false,
			passwordRepeatError: false,
		};

		this.state = {
			clearInputsState: false,
			tabIndex: 0,
			...this.defaultErrors,
		};

		this.userClient = container.get('userClient');

		this.usernameRef = React.createRef();
		this.passwordRef = React.createRef();
		this.passwordRepeatRef = React.createRef();

		this.reset = this.reset.bind(this);
		this.submit = this.submit.bind(this);
	}

	reset() {
		this.setState({
			clearInputsState: !this.state.clearInputsState,
			...this.defaultErrors,
		});
	}

	submit() {
		let errorState = {};

		if (!this.usernameRef.current || !this.passwordRef.current || (this.state.tabIndex === 1 && !this.passwordRepeatRef.current)) {
			console.error('Invalid input reference!');

			return;
		}

		let username = this.usernameRef.current.value;
		username = username.trim();
		if (!this.isTextFieldValid(username, 4, 255)) {
			errorState.usernameError = true;
		}

		let password = this.passwordRef.current.value;
		if (!this.isTextFieldValid(password, 4, 255)) {
			errorState.passwordError = true;
		}

		if (this.state.tabIndex === 1) {
			let passwordRe = this.passwordRepeatRef.current.value;
			if (passwordRe !== password) {
				errorState.passwordRepeatError = true;
			}
		}

		if (Object.keys(errorState).length > 0) {
			return this.setState({...this.defaultErrors, ...errorState});
		}

		this.setState({loading: true});

		let request;
		if (this.state.tabIndex === 0) {
			request = this.userClient.login({
				username,
				password,
			});
		} else {
			request = this.userClient.register({
				username,
				password,
			});
		}

		request
			.then(this.props.setCurrentUser)
			.catch(console.error);
	}

	isTextFieldValid(message, minLength, maxLength, regex = BASE_INPUT_REGEX) {
		if (minLength > message.length || maxLength < message.length) {
			return false;
		}

		if (!regex.test(message)) {
			return false;
		}

		return true;
	}

	getKey(fieldName) {
		return `${fieldName}-${this.state.tabIndex}-${this.state.clearInputsState}`;
	}

	onChange(errKey) {
		return () => {
			if (this.state[errKey]) {
				this.setState({[errKey]: false});
			}
		};
	}

	render() {
		const {classes} = this.props;
		const {tabIndex, loading, usernameError, passwordError, passwordRepeatError} = this.state;

		if (loading) {
			return <LoadingIndication />;
		}

		return (
			<div className={classes.container}>
				<Tabs
					fullWidth
					indicatorColor="primary"
					textColor="primary"
					value={tabIndex}
					onChange={(_, idx) => this.setState({tabIndex: idx})}
				>
					<Tab label="Login" />
					<Tab label="Sign Up" />
				</Tabs>
				<div style={{width: 500, display: 'flex', flexDirection: 'column', alignItems: 'center', padding: 20}}>
					<TextField
						autoFocus
						fullWidth
						key={this.getKey('username')}
						error={usernameError}
						inputRef={this.usernameRef}
						label="Username*"
						className={classes.textField}
						margin="normal"
						variant="outlined"
						onChange={this.onChange('usernameError')}
						helperText={usernameError ? BASE_HELPER_TEXT + '4-255' : null}
					/>
					<TextField
						fullWidth
						key={this.getKey('password')}
						error={passwordError}
						inputRef={this.passwordRef}
						label="Password*"
						type='password'
						className={classes.textField}
						margin="normal"
						variant="outlined"
						onChange={this.onChange('passwordError')}
						helperText={passwordError ? BASE_HELPER_TEXT + '6-255' : null}
					/>
					{tabIndex === 1 &&
						<TextField
							fullWidth
							key={this.getKey('passwordRe')}
							error={passwordRepeatError}
							inputRef={this.passwordRepeatRef}
							label="Repeat Password*"
							type='password'
							className={classes.textField}
							margin="normal"
							variant="outlined"
							onChange={this.onChange('passwordRepeatError')}
							helperText={passwordRepeatError ? 'Passwords mismatched' : null}
						/>
					}
					<div style={{alignSelf: 'flex-end', marginTop: 20}}>
						<Button
							variant="contained"
							color="default"
							className={classes.button}
							onClick={this.reset}
							children='Reset'
						/>
						<Button
							variant="contained"
							color="primary"
							className={classes.button}
							onClick={this.submit}
							children={tabIndex === 1 ? 'Sign Up' : 'Login'}
						/>
					</div>
				</div>
			</div>
		);
	}
}

export default withStyles(styles)(Login);
