import React from 'react';
import PropTypes from 'prop-types';
import Avatar from '@material-ui/core/Avatar';

import Icon from './icon';
import ArrowTooltip from './arrow-tooltip';
import {getUserName} from '../utils';

const UserAvatar = ({userId, className = undefined, size = 50, withTooltip = false}, {usersMap}) => {
	let url = null;
	let user = usersMap[userId];
	if (user && user.blobUrl) {
		url = user.blobUrl;
	}

	let avatar;
	if (url) {
		avatar = <Avatar className={className} style={{width: size, height: size}} alt='avatar' src={url} />;
	} else {
		let fontSize = 'default';
		if (size > 50) {
			fontSize = 'large';
		} else if (size < 50) {
			fontSize = 'small';
		}

		avatar = (
			<Avatar style={{width: size, height: size}} className={className} alt='avatar'>
				<Icon name='person' fontSize={fontSize} />
			</Avatar>
		);
	}

	if (withTooltip && user) {
		return (
			<ArrowTooltip tooltip={getUserName(user)}>
				{avatar}
			</ArrowTooltip>
		);
	}

	return avatar;
};

UserAvatar.contextTypes = {
	usersMap: PropTypes.object.isRequired,
};

export default UserAvatar;
