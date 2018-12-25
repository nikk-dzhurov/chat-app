import React from 'react';
import PropTypes from 'prop-types';
import Icon from '@material-ui/core/Icon';
import Avatar from '@material-ui/core/Avatar';

const UserAvatar = ({userId, size = 50}, {usersMap}) => {
	let url = null;
	if (usersMap[userId] && usersMap[userId].blobUrl) {
		url = usersMap[userId].blobUrl;
	}

	if (url) {
		return <Avatar style={{width: size, height: size}} alt='avatar' src={url} />;
	}

	let fontSize = 'default';
	if (size > 50) {
		fontSize = 'large';
	} else if (size < 50) {
		fontSize = 'small';
	}

	return (
		<Avatar style={{width: size, height: size}} alt='avatar'>
			<Icon fontSize={fontSize}>person</Icon>
		</Avatar>
	);
};

UserAvatar.contextTypes = {
	usersMap: PropTypes.object.isRequired,
};

export default UserAvatar;
