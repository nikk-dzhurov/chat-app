import React from 'react';
import PropTypes from 'prop-types';
import dateformat from 'dateformat';

import Typography from '@material-ui/core/Typography';
import {withStyles} from '@material-ui/core/styles';

import UserAvatar from '../atoms/user-avatar';

const styles = theme => ({
	container: {
		display: 'flex',
		flexDirection: 'row',
		marginBottom: 10,
		alignSelf: 'flex-start',
	},
	containerReversed: {
		display: 'flex',
		flexDirection: 'row-reverse',
		marginBottom: 10,
		alignSelf: 'flex-end',
	},
	avatar: {
		marginLeft: 10,
		marginRight: 10,
		alignSelf: 'flex-end',
	},
	text: {
		padding: '5px 10px',
		borderRadius: 10,
		maxWidth: '70%',
		wordBreak: 'break-word',
		backgroundColor: theme.palette.divider,
	},
	textCurrentUser: {
		backgroundColor: theme.palette.primary.main,
		color: theme.palette.common.white,
		// marginRight: 10,
	},
	time: {
		padding: '5px',
		alignSelf: 'center',
		fontSize: 12,
		color: theme.palette.grey[600],
	},
	separator: {
		textAlign: 'center',
		backgroundColor: theme.palette.grey[100],
		borderBottom: '1px solid ' + theme.palette.divider,
		marginBottom: 10,
	},
	noAvatar: {
		width: 50,
	},
});

class Message extends React.Component {
	render() {
		const {classes, message, hasAvatar, isCurrentUser, hasDateSeparator} = this.props;
		const {message: text, userId, createdAt} = message;

		let time = dateformat(createdAt, 'HH:MM');
		let daySeparatorText = dateformat(createdAt, 'mmmm dS, yyyy');

		return (
			<div>
				{hasDateSeparator &&
					<Typography variant='overline' className={classes.separator}>{daySeparatorText}</Typography>
				}
				<div className={isCurrentUser ? classes.containerReversed : classes.container}>
					{hasAvatar ?
						<UserAvatar
							size={30}
							withTooltip
							userId={userId}
							className={classes.avatar}
						/>
						:
						<span className={classes.noAvatar} />
					}
					<Typography variant='body2' className={`${classes.text} ${isCurrentUser ? classes.textCurrentUser : ''}`}>{text}</Typography>
					<span className={classes.time}>{time}</span>
				</div>
			</div>
		);
	}
}
Message.propTypes = {
	message: PropTypes.object.isRequired,
	hasAvatar: PropTypes.bool.isRequired,
	isCurrentUser: PropTypes.bool.isRequired,
};

export default withStyles(styles)(Message);
