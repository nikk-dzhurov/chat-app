import React from 'react';
import PropTypes from 'prop-types';
import Avatar from '@material-ui/core/Avatar';

import Icon from './icon';
import ArrowTooltip from './arrow-tooltip';
import {getUserName} from '../utils';

const UserAvatar = ({userId, showActiveStatus = false, className = undefined, size = 50, withTooltip = false}, {usersMap}) => {
	let url = null;
	let user = usersMap[userId];
	if (user && user.blobUrl) {
		url = user.blobUrl;
	}

	let avatar;
	if (url) {
		avatar = <Avatar style={{width: size, height: size}} alt='avatar' src={url} />;
	} else {
		let fontSize = 'default';
		if (size > 50) {
			fontSize = 'large';
		} else if (size < 50) {
			fontSize = 'small';
		}

		avatar = (
			<Avatar style={{width: size, height: size}} alt='avatar'>
				<Icon name='person' fontSize={fontSize} />
			</Avatar>
		);
	}

	avatar = (
		<div className={className} style={{width: size, height: size}}>
			{avatar}
			{showActiveStatus && user && user.active &&
				<span style={{
					height: size / 5,
					width: size / 5,
					backgroundColor: 'limegreen',
					borderRadius: '50%',
					borderColor: '#AFAFAF',
					borderWidth: 1,
					position: 'absolute',
					top: size,
					left: size,
				}} />
			}
		</div>
	);

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
