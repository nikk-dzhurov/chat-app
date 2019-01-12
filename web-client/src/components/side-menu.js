import React from 'react';
import PropTypes from 'prop-types';
import {withRouter} from 'react-router-dom';
import Divider from '@material-ui/core/Divider';
import Drawer from '@material-ui/core/Drawer';
import {withStyles} from '@material-ui/core/styles';
import ListItem from '@material-ui/core/ListItem';
import Typography from '@material-ui/core/Typography';
import List from '@material-ui/core/List';
import ListItemText from '@material-ui/core/ListItemText';
import ListItemIcon from '@material-ui/core/ListItemIcon';

import UserAvatar from '../atoms/user-avatar';
import Icon from '../atoms/icon';
import {getUserName} from '../utils';

const styles = theme => ({
	container: {
		minWidth: 250,
	},
	userContainer: {
		padding: 24,
		backgroundColor: theme.palette.primary.main,
		color: theme.palette.primary.contrastText,
	},
});

class SideMenu extends React.Component {

	navigateTo(route) {
		const {history} = this.props;
		if (history.location.pathname !== route) {
			history.push(route);
		}

		this.props.toggleDrawer();
	}

	render() {
		const {classes, isOpen, toggleDrawer} = this.props;
		const {currentUser, logout} = this.context;

		return (
			<Drawer open={isOpen} onClose={toggleDrawer}>
				<div className={classes.container}>
					<div className={classes.userContainer}>
						<UserAvatar userId={currentUser.id} size={120} />
						<Typography variant="h6" color="inherit" style={{marginTop: 10}}>
							{getUserName(currentUser)}
						</Typography>
					</div>
					<Divider />
					<List>
						<ListItem button onClick={() => this.navigateTo('/')}>
							<ListItemIcon>
								<Icon name='chat' />
							</ListItemIcon>
							<ListItemText primary='Chats' />
						</ListItem>
						<ListItem button onClick={() => this.navigateTo('/profile')}>
							<ListItemIcon>
								<Icon name='person' />
							</ListItemIcon>
							<ListItemText primary='My Profile' />
						</ListItem>
					</List>
					<Divider />
					<List>
						<ListItem button onClick={logout}>
							<ListItemIcon>
								<Icon name='logout' />
							</ListItemIcon>
							<ListItemText primary='Logout' />
						</ListItem>
					</List>
				</div>
			</Drawer>
		);
	}
}
SideMenu.propTypes = {
	isOpen: PropTypes.bool.isRequired,
	toggleDrawer: PropTypes.func.isRequired,
};

SideMenu.contextTypes = {
	currentUser: PropTypes.object.isRequired,
	logout: PropTypes.func.isRequired,
};

const SideMenuWithRouter = withRouter(SideMenu);
export default withStyles(styles)(SideMenuWithRouter);
