import React from 'react';
import PropTypes from 'prop-types';
import {withRouter} from 'react-router-dom';

import Button from '@material-ui/core/Button';
import {withStyles} from '@material-ui/core/styles';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import IconButton from '@material-ui/core/IconButton';
import Typography from '@material-ui/core/Typography';

import Icon from '../atoms/icon';
import UserAvatar from '../atoms/user-avatar';
import {getUserName} from '../utils';

const styles = theme => ({
	appBarRoot: {
		height: 64,
	},
	toolbar: {
		height: 64,
		paddingLeft: 16,
		paddingRight: 16,
	},
	logoButton: {
		paddingLeft: 16,
		paddingRight: 16,
	},
	userData: {
		flex: 1,
		display: 'flex',
		alignItems: 'center',
		justifyContent: 'flex-end',
		flexDirection: 'row',
	},
	userName: {
		textTransform: 'none',
		marginRight: 16,
	},
});

class Navbar extends React.Component {

	navigateTo(route) {
		const {history} = this.props;
		if (history.location.pathname !== route) {
			history.push(route);
		}
	}

	render() {
		const {classes} = this.props;
		const {currentUser} = this.context;

		return (
			<AppBar position="static" classes={{root: classes.appBarRoot}}>
				<Toolbar classes={{root: classes.toolbar}}>
					{currentUser &&
						<IconButton color="inherit" onClick={this.props.toggleDrawer}>
							<Icon name='menu' />
						</IconButton>
					}
					<Button color="inherit" className={classes.logoButton} onClick={() => this.navigateTo('/')}>
						<Typography variant="h6" color="inherit" noWrap>
							ChatApp
						</Typography>
					</Button>
					{currentUser &&
						<div className={classes.userData}>
							<Button color="inherit" onClick={() => this.navigateTo('/profile')}>
								<Typography variant="h6" color="inherit" className={classes.userName} noWrap>
									{getUserName(currentUser)}
								</Typography>
								<UserAvatar userId={currentUser.id} size={40} />
							</Button>
						</div>
					}
				</Toolbar>
			</AppBar>
		);
	}
}
Navbar.propTypes = {
	toggleDrawer: PropTypes.func,
};
Navbar.contextTypes = {
	currentUser: PropTypes.object,
};

const NavbarWithRouter = withRouter(Navbar);
export default withStyles(styles)(NavbarWithRouter);
