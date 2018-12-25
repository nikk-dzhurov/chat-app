import React from 'react';
import PropTypes from 'prop-types';
import dateformat from 'dateformat';

import Button from '@material-ui/core/Button';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import Icon from '@material-ui/core/Icon';
import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import Avatar from '@material-ui/core/Avatar';
import TextField from '@material-ui/core/TextField';
import Tab from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import {withStyles} from '@material-ui/core/styles';

import container from '../container';
import LoadingIndication from '../components/loading-indication';
import UserAvatar from '../components/user-avatar';

export default class Chat extends React.Component {
	constructor(props) {
		super(props);

		this.userClient = container.get('userClient');
		this.chatClient = container.get('chatClient');

		this.state = {
			loading: true,
			users: [],
			chats: [],
			currentChatId: null,
			createChatDialogOpen: false,
		};
	}

	componentDidMount() {
		this.loadData();
	}

	async loadData() {
		const {currentUser} = this.context;
		Promise.all([this.userClient.list(), this.chatClient.list()])
			.then(data => {
				let users = data[0] || [];
				let chats = data[1] || [];
				users = users.filter(u => u.id !== currentUser.id);

				this.setState({
					loading: false,
					users,
					chats,
				});
			})
			.catch(err => {
				console.error(err);

				this.setState({
					loading: false,
					error: 'Failed to load initial data',
				});
			});
	}

	createOrOpenChat(userId) {
		if (!userId) {
			return;
		}

		const {currentUser} = this.context;
		let chat = this.state.chats.find(c => c.creatorId === userId || c.directUserId === userId);
		if (chat) {
			this.setState({
				currentChatId: chat.id,
				createChatDialogOpen: false,
			});
		} else {
			this.setState({
				createChatDialogOpen: false,
				creatingChat: true,
			});

			this.chatClient.create({
				creatorId: currentUser.id,
				directUserId: userId,
			})
				.then(chat => {
					let chats = this.addOrReplaceById(this.state.chats, chat);
					this.setState({
						chats,
						creatingChat: false,
						currentChatId: chat.id,
					});
				})
				.catch(err => {
					console.error(err);

					this.setState({
						creatingChat: false,
						currentChatId: null,
					});
				});
		}
	}

	addOrReplaceById(arr, entity) {
		let result = [...arr];
		let idx = result.findIndex(e => e.id === entity.id);
		if (idx !== -1) {
			result[idx] = entity;
		} else {
			result.push(entity);
		}

		return result;
	}

	renderUsersList() {
		const {users} = this.state;

		return (
			<List style={{minWidth: 360}}>
				{users.map(u => (
					<ListItem
						button
						key={u.id}
						alignItems="flex-start"
						onClick={() => this.createOrOpenChat(u.id)}
					>
						{false ?
							<Avatar alt="user-avatar" src="" />
							:
							<Avatar>
								<Icon>person</Icon>
							</Avatar>
						}
						<ListItemText
							primary={u.fullName ? u.fullName : u.username}
							secondary={`Joined ${dateformat(u.createdAt, 'dd.mm.yyyy')}`}
						/>
					</ListItem>
				))}
			</List>
		);
	}

	renderChatAvatar(chat) {
		let userId = chat.creatorId;
		if (this.context.currentUser.id === chat.creatorId) {
			userId = chat.directUserId;
		}

		return <UserAvatar userId={userId} />;
	}

	renderChatList() {
		const {chats, currentChatId} = this.state;

		return (
			<List style={{minWidth: 360}}>
				<ListItem
					button
					key={0}
					alignItems="flex-start"
					onClick={() => this.setState({createChatDialogOpen: true})}
				>
					<ListItemIcon>
						<Icon>add_circle</Icon>
					</ListItemIcon>
					<ListItemText primary='Create new chat' />
				</ListItem>
				{chats.map(c => (
					<ListItem
						button
						key={c.id}
						selected={c.id === currentChatId}
						alignItems="flex-start"
						onClick={() => this.setState({currentChatId: c.id})}
					>
						{this.renderChatAvatar(c)}
						<ListItemText
							primary={c.title ? c.title : c.creatorId}
							secondary={`Created at ${dateformat(c.createdAt, 'dd.mm.yyyy')}`}
						/>
					</ListItem>
				))}
			</List>
		);
	}

	render() {
		const {createChatDialogOpen, chats} = this.state;
		let messages = [];

		if (this.state.loading) {
			return <LoadingIndication />;
		}

		return (
			<div style={{display: 'flex', flexDirection: 'row'}}>
				<div style={{borderRight: '1px solid #7c7c7c', flex: 1}}>
					<h2>Chats</h2>
					{this.renderChatList()}
				</div>
				<div style={{flex: 1}}>
					<h2>Chat messages</h2>
					{messages.map((m, idx) => (
						<div key={idx}>
							{m.message}
						</div>
					))}
				</div>
				<Dialog
					open={createChatDialogOpen}
					onClose={() => this.setState({createChatDialogOpen: false})} aria-labelledby="simple-dialog-title"
				>
					<DialogTitle>Send Message To</DialogTitle>
					{this.renderUsersList()}
				</Dialog>
			</div>
		);
	}
}
Chat.contextTypes = {
	currentUser: PropTypes.object.isRequired,
	usersMap: PropTypes.object.isRequired,
};
