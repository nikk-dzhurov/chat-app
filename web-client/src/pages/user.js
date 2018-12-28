import React from 'react';
import PropTypes from 'prop-types';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import {withStyles} from '@material-ui/core/styles';

import container from '../container';
import UserAvatar from '../components/user-avatar';

const allowedImageFileTypes = ['image/jpeg', 'image/png'];

const styles = theme => ({
	container: {
		display: 'flex',
		flexDirection: 'row',
		justifyContent: 'center',
		alignItems: 'center',
		marginTop: 24,
	},
	innerContainer: {
		flexGrow: 1,
		maxWidth: 500,
		display: 'flex',
		alignSelf: 'center',
		alignItems: 'center',
		flexDirection: 'column',
	},
	buttonsContainer: {
		alignSelf: 'flex-end',
		marginTop: 16,
	},
	inputFields: {
		width: '100%',
		borderTop: '1px solid ' + theme.palette.divider,
		paddingTop: 16,
		marginTop: 16,
	},
	avatarForm: {
		display: 'flex',
		alignItems: 'center',
		flexDirection: 'column',
		paddingBotom: 24,
	},
	avatar: {
		marginBottom: 16,
	},
});

class User extends React.Component {
	constructor(props, context) {
		super(props);

		this.state = {
			fullName: context.currentUser.fullName || '',
			fullNameError: false,
		};

		this.userClient = container.get('userClient');

		this.handleSelectAvatar = this.handleSelectAvatar.bind(this);
		this.updateUserData = this.updateUserData.bind(this);
		this.handleKeyUp = this.handleKeyUp.bind(this);

		this.fullNameRef = React.createRef();
	}

	componentDidMount() {
		this.fullNameRef.current.addEventListener('keyup', this.handleKeyUp);
	}

	componentWillUnmount() {
		this.fullNameRef.current.removeEventListener('keyup', this.handleKeyUp);
	}

	handleKeyUp(e) {
		let keyCode = e.keyCode;
		if (keyCode === 13) {
			this.updateUserData();
		}
	}

	handleSelectAvatar(e) {
		const {currentUser} = this.context;
		const {files} = e.target;
		if (!files || files.length === 0) {
			return;
		}

		let imageFile = files[0];
		if (!imageFile) {
			return;
		}

		e.target.value = '';
		if (imageFile.size > 1024 * 1024 * 15) {
			console.error('File is bigger than 15MB');

			return;
		}

		if (allowedImageFileTypes.indexOf(imageFile.type) === -1) {
			console.error('File type is not allowed.');

			return;
		}

		this.userClient.uploadAvatar(currentUser.id, imageFile)
			.then(() => {
				this.props.updateCurrentUserData({id: currentUser.id});
			})
			.catch(console.error);
	}

	updateUserData() {
		let fullName = this.fullNameRef.current.value;
		fullName = fullName.trim();
		if (fullName.length > 255) {
			this.setState({
				fullNameError: true,
			});

			return;
		}

		if (fullName === this.context.currentUser.fullName) {
			return;
		}

		this.userClient.update(this.context.currentUser.id, {
			fullName,
			id: this.context.currentUser.id,
		})
			.then(user => {
				if (user) {
					this.props.updateCurrentUserData(user);
				}
			})
			.catch(console.error);
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
		const {currentUser} = this.context;

		return (
			<div className={classes.container}>
				<div className={classes.innerContainer}>
					<div className={classes.avatarForm}>
						<UserAvatar userId={currentUser.id} className={classes.avatar} size={100} />
						<input
							id="avatar-input"
							type='file'
							style={{display: 'none'}}
							onChange={this.handleSelectAvatar}
						/>
						<Button
							variant="contained"
							color="primary"
							onClick={() => {
								let el = document.getElementById('avatar-input');
								if (el) {
									el.click();
								}
							}}
							children='Change Avatar'
						/>
					</div>
					<div className={classes.inputFields}>
						<TextField
							fullWidth
							disabled
							label="Username"
							value={currentUser.username}
							margin="normal"
							variant="outlined"
						/>
						<TextField
							autoFocus
							fullWidth
							label="Full Name"
							error={this.state.fullNameError}
							inputRef={this.fullNameRef}
							defaultValue={this.state.fullName}
							onChange={this.onChange('fullNameError')}
							helperText={this.state.fullNameError ? 'Max symbols: 255' : null}
							margin="normal"
							variant="outlined"
						/>
					</div>
					<div className={classes.buttonsContainer}>
						<Button
							variant="contained"
							color="primary"
							onClick={this.updateUserData}
							children='Save'
						/>
					</div>
				</div>
			</div>
		);
	}
}
User.contextTypes = {
	usersMap: PropTypes.object.isRequired,
	currentUser: PropTypes.object.isRequired,
	logout: PropTypes.func.isRequired,
};

export default withStyles(styles)(User);
