import React from 'react';
import PropTypes from 'prop-types';
import Button from '@material-ui/core/Button';
import Avatar from '@material-ui/core/Avatar';
import {withStyles} from '@material-ui/core/styles';

import container from '../container';
import UserAvatar from '../components/user-avatar';

const allowedImageFileTypes = ['image/jpeg', 'image/png'];

const styles = theme => ({
	bigAvatar: {
		width: 100,
		height: 100,
	},
});

class User extends React.Component {
	constructor(props) {
		super(props);

		this.userClient = container.get('userClient');

		this.handleSelectAvatar = this.handleSelectAvatar.bind(this);
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
			.catch(console.error)
			.then(console.log);
	}

	render() {
		const {classes, logout} = this.props;
		const {currentUser, usersMap} = this.context;

		// let currUserData = usersMap[currentUser.id];
		// let avatarUrl = null;
		// if (currUserData && currUserData.blobUrl) {
		// 	avatarUrl = currUserData.blobUrl;
		// }

		return (
			<div>
				<UserAvatar userId={currentUser.id} size={100} />
				<h1>{currentUser.username}</h1>
				<h2>{currentUser.fullName}</h2>
				<Button
					variant="contained"
					color="primary"
					onClick={logout}
					children='Logout'
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
				<input
					id="avatar-input"
					type='file'
					style={{display: 'none'}}
					onChange={this.handleSelectAvatar}
				/>
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
