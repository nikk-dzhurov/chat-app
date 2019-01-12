import React from 'react';
import PropTypes from 'prop-types';
import dateformat from 'dateformat';

import List from '@material-ui/core/List';
import Divider from '@material-ui/core/Divider';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import InputAdornment from '@material-ui/core/InputAdornment';
import TextField from '@material-ui/core/TextField';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import {withStyles} from '@material-ui/core/styles';

import container from '../container';
import LoadingIndication from '../components/loading-indication';
import UserAvatar from '../components/user-avatar';
import Icon from '../components/icon';
import Message from '../components/message';
import {getUserName} from '../utils';

const maxMessageDuration = 60 * 1000 * 5;
const styles = theme => ({
	container: {
		display: 'flex',
		flexDirection: 'row',
	},
	inputContainer: {
		borderTop: '1px solid ' + theme.palette.divider,
		borderBottom: '1px solid ' + theme.palette.divider,
		padding: 10,
		justifySelf: 'flex-end',
	},
	chatListContainer: {
		flex: 2,
		minWidth: 200,
		maxWidth: 320,
		borderRight: '1px solid ' + theme.palette.divider,
		flexDirection: 'column',
	},
	messageListContainer: {
		flex: 5,
		display: 'flex',
		flexDirection: 'column',
	},
	messageList: {
		display: 'flex',
		flexGrow: 1,
		flexDirection: 'column',
		paddingTop: 10,
		overflowY: 'scroll',
	},
	sectionHeader: {
		display: 'flex',
		flexDirection: 'column',
		justifyContent: 'center',
		minHeight: 60,
		alignItems: 'center',
		borderBottom: '1px solid rgba(0, 0, 0, 0.12)',
		padding: 8,
	},
	subtitleActive: {
		color: 'limegreen',
	},
	subtitle: {
		color: 'grey',
	},
});

class Chat extends React.Component {
	constructor(props) {
		super(props);

		this.userClient = container.get('userClient');
		this.chatClient = container.get('chatClient');
		this.messageClient = container.get('messageClient');
		this.wsClient = container.get('wsClient');

		this.state = {
			loading: true,
			users: [],
			chats: [],
			currentChatId: null,
			messagesLoading: false,
			messagesMap: {},
			createChatDialogOpen: false,
		};

		this.inputRef = React.createRef();
		this.messagesEnd = React.createRef();

		this.sendMessage = this.sendMessage.bind(this);
		this.handleKeyUp = this.handleKeyUp.bind(this);
		this.handleChatChange = this.handleChatChange.bind(this);
		this.handleMessageChange = this.handleMessageChange.bind(this);
	}

	componentDidMount() {
		this.wsClient.addChangeListener('chat', this.handleChatChange);
		this.wsClient.addChangeListener('message', this.handleMessageChange);
		this.loadInitialData();
	}

	componentDidUpdate(prevProps, prevState) {
		if (prevState.loading && !this.state.loading && this.inputRef.current) {
			this.inputRef.current.addEventListener('keyup', this.handleKeyUp);
		} else if (!prevState.loading && this.state.loading && this.inputRef.current) {
			this.inputRef.current.removeEventListener('keyup', this.handleKeyUp);
		}

		if (prevState.currentChatId !== this.state.currentChatId) {
			this.reloadMessages(this.state.currentChatId);
		}
	}

	componentWillUnmount() {
		if (this.inputRef.current) {
			this.inputRef.current.removeEventListener('keyup', this.handleKeyUp);
		}

		this.wsClient.removeChangeListener('chat', this.handleChatChange);
		this.wsClient.removeChangeListener('message', this.handleMessageChange);
	}

	handleKeyUp(e) {
		let keyCode = e.keyCode;
		if (keyCode === 13) {
			this.sendMessage();
		}
	}

	scrollToEnd(instant = false) {
		if (this.messagesEnd.current) {
			let opts = {block: 'end', behavior: 'smooth'};
			if (instant) {
				// opts.behavior = 'instant';
			}

			this.messagesEnd.current.scrollIntoView(opts);
		}
	}

	handleMessageChange(msg) {
		if (!msg || !msg.type || !msg.messageId || !msg.chatId) {
			return;
		}

		switch (msg.type) {
			case 'message_create':
			case 'message_update':

				this.messageClient.get(msg.chatId, msg.messageId)
					.then(message => {
						if (!message) {
							return;
						}

						let messagesMap = this.addNewMessageToMessageMap(message);
						this.setState({messagesMap});

						if (message.chatId === this.state.currentChatId && msg.type === 'message_create') {
							this.scrollToEnd();
						}
					})
					.catch(console.error);
				break;
			case 'message_delete':
				let messagesMap = this.removeMessage(msg.chatId, msg.messageId);
				if (messagesMap !== this.state.messagesMap) {
					this.setState({messagesMap});
				}

				break;
			default:
				console.log('Unrecognized message type:', msg.type);
		}
	}

	handleChatChange(msg) {
		if (!msg || !msg.type || !msg.chatId) {
			return;
		}

		switch (msg.type) {
			case 'chat_create':
			case 'chat_update':
				this.chatClient.get(msg.chatId)
					.then(chat => {
						if (!chat) {
							return;
						}

						let chats = this.addOrReplaceById(this.state.chats, chat);
						this.setState({chats});
					})
					.catch(console.error);

				break;
			case 'chat_delete':
				let idx = this.state.chats.findIndex(c => c.id === msg.chatId);
				if (idx !== -1) {
					let chats = [...this.state.chats];
					chats.splice(idx, 1);
					let messagesMap = this.state.messagesMap;
					if (messagesMap[msg.chatId]) {
						messagesMap = {...messagesMap};
						delete messagesMap[msg.chatId];
					}

					this.setState({
						chats,
						messagesMap,
					});
				}

				break;
			default:
				console.log('Unrecognized message type:', msg.type);
		}
	}

	removeMessageFromMap(chatId, messageId) {
		let chatMessages = this.state.messagesMap[chatId] || [];
		let idx = chatMessages.findIndex(m => m.id === messageId);
		if (idx === -1) {
			return this.state.messagesMap;
		}

		chatMessages = [...chatMessages];
		chatMessages.splice(idx, 1);

		let newMap = {...this.state.messagesMap};
		newMap[chatId] = chatMessages;

		return newMap;
	}

	async loadInitialData() {
		this.chatClient.list()
			.then(chats => {
				chats = chats || [];

				let currentChatId = null;
				if (chats.length > 0) {
					currentChatId = chats[0].id;
				}

				this.setState({
					loading: false,
					chats,
					currentChatId,
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

	sendMessage() {
		if (!this.inputRef.current) {
			return;
		}

		const {currentChatId} = this.state;
		if (!currentChatId) {
			return;
		}

		let message = this.inputRef.current.value;
		message = message.trim();
		if (message.length === 0) {
			return;
		}

		this.inputRef.current.value = '';
		this.setState({
			creatingMessage: true,
		});

		this.messageClient.create(currentChatId, {
			message,
			chatId: currentChatId,
			userId: this.context.currentUser.id,
		})
			.then(message => {
				if (!message) {
					return this.setState({
						creatingMessage: false,
					});
				}

				let newMap = this.addNewMessageToMessageMap(message);
				this.setState({
					creatingMessage: false,
					messagesMap: newMap,
				});

				this.scrollToEnd();
			})
			.catch(err => {
				console.error(err);

				this.setState({
					creatingMessage: false,
				});
			});
	}

	addNewMessageToMessageMap(message) {
		let chatMessages = this.state.messagesMap[message.chatId] || [];
		let idx = chatMessages.findIndex(m => m.id === message.id);
		if (idx !== -1) {
			chatMessages = [...chatMessages];
			chatMessages[idx] = message;
		} else {
			chatMessages = [...chatMessages, message];
		}

		chatMessages = chatMessages.sort(this.sortByCreatedAt());
		let newMap = {...this.state.messagesMap};
		newMap[message.chatId] = chatMessages;

		return newMap;
	}

	reloadMessages(chatId) {
		this.setState({messagesLoading: true});

		this.messageClient.list(chatId)
			.then(list => {
				list = list || [];
				list = list.sort(this.sortByCreatedAt());
				let newMap = {...this.state.messagesMap};
				newMap[chatId] = list;

				this.setState({
					messagesLoading: false,
					messagesMap: newMap,
				});

				this.scrollToEnd(true);
			})
			.catch(err => {
				console.log(err);

				// let newMap = {...this.state.messagesMap};
				// newMap[chatId] = [];
				// this.setState({
				// 	messagesLoading: false,
				// 	messagesMap: newMap,
				// });
				//
				// this.scrollToEnd(true);
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

	sortByCreatedAt(isAsc = true) {
		let diffFn = (a, b) => b - a;
		if (isAsc) {
			diffFn = (a, b) => a - b;
		}

		return (a, b) => {
			let d1 = new Date(a.createdAt);
			let d2 = new Date(b.createdAt);

			let diff = diffFn(d1, d2);
			if (diff > 0) {
				return 1;
			} else if (diff < 0) {
				return -1;
			}

			return 0;
		};
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
		const {usersMap, currentUser} = this.context;

		let users = [];
		for (let id in usersMap) {
			if (id !== currentUser.id) {
				users.push(usersMap[id]);
			}
		}

		users = users.sort((u1, u2) => {
			let name1 = getUserName(u1);
			let name2 = getUserName(u2);
			if (name1 < name2) {
				return -1;
			}

			if (name1 > name2) {
				return 1;
			}

			return 0;
		});

		return (
			<List style={{minWidth: 360}}>
				{users.map(u => (
					<ListItem
						button
						key={u.id}
						alignItems="flex-start"
						onClick={() => this.createOrOpenChat(u.id)}
					>
						<UserAvatar showActiveStatus userId={u.id} />
						<ListItemText
							primary={getUserName(u)}
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

		return <UserAvatar showActiveStatus userId={userId} />;
	}

	renderChatList() {
		const {chats, currentChatId} = this.state;
		const {classes} = this.props;

		return (
			<div className={classes.chatListContainer}>
				<div className={classes.sectionHeader}>
					<Typography align='center' variant="h5" noWrap>
						Chat List
					</Typography>
				</div>
				<List>
					<ListItem
						button
						key={0}
						alignItems="flex-start"
						onClick={() => this.setState({createChatDialogOpen: true})}
					>
						<ListItemIcon>
							<Icon name='add_circle' />
						</ListItemIcon>
						<ListItemText primary='Send message' />
					</ListItem>
					<Divider />
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
								primary={this.getChatTitle(c)}
								secondary={`Created at ${dateformat(c.createdAt, 'dd.mm.yyyy')}`}
							/>
						</ListItem>
					))}
				</List>
			</div>
		);
	}

	getChatTitle(chat) {
		if (chat.title) {
			return chat.title;
		}

		let user = this.getChatUser(chat);

		return getUserName(user);
	}

	getChatUser(chat) {
		const {usersMap, currentUser} = this.context;
		let userId = chat.creatorId;
		if (userId === currentUser.id) {
			userId = chat.directUserId;
		}

		if (!userId) {
			return null;
		}

		return usersMap[userId] || null;
	}

	shouldAddDateSeparator(prev, curr) {
		if (!prev) {
			return true;
		}

		let prevDate = new Date(prev.createdAt);
		let currDate = new Date(curr.createdAt);

		if (
			prevDate.getDate() !== currDate.getDate() ||
			prevDate.getMonth() !== currDate.getMonth() ||
			prevDate.getFullYear() !== currDate.getFullYear()
		) {
			return true;
		}

		return false;
	}

	shouldAddAvatar(curr, next) {
		if (!next || next.userId !== curr.userId || (new Date(next.createdAt) - new Date(curr.createdAt)) > maxMessageDuration) {
			return true;
		}

		return false;
	}

	renderMessageList() {
		const {classes} = this.props;
		const {currentUser} = this.context;
		const {currentChatId, messagesMap} = this.state;
		let messages = [];
		let currentChat = null;
		if (currentChatId) {
			messages = messagesMap[currentChatId] || [];
			currentChat = this.state.chats.find(({id}) => id === currentChatId);
		}

		let chatUser = null;
		if (currentChat) {
			chatUser = this.getChatUser(currentChat);
		}

		return (
			<div id='messages' className={classes.messageListContainer}>
				<div className={classes.sectionHeader}>
					<Typography align='center' variant="h5" noWrap>
						{currentChat ? this.getChatTitle(currentChat) : 'Messages'}
					</Typography>
					{!!chatUser &&
						<Typography align='center' variant="body2" className={chatUser.active ? classes.subtitleActive : classes.subtitle} noWrap>
							{chatUser.active ? 'Online' : 'Offline'}
						</Typography>
					}
				</div>
				{this.state.messagesLoading && messages.length === 0 ?
					<LoadingIndication />
					:
					<div className={classes.messageList}>
						{messages.map((m, idx) => (
							<Message
								key={idx}
								hasAvatar={this.shouldAddAvatar(m, messages[idx + 1])}
								hasDateSeparator={this.shouldAddDateSeparator(messages[idx - 1], m)}
								message={m}
								isCurrentUser={currentUser.id === m.userId}
							/>
						))}
						{this.state.messagesLoading &&
							<div style={{minHeight: 80, display: 'flex'}}>
								<LoadingIndication />
							</div>
						}
						<div ref={this.messagesEnd} />
					</div>
				}
				<div className={classes.inputContainer}>
					<TextField
						fullWidth
						id='message-input'
						autoComplete='off'
						inputRef={this.inputRef}
						placeholder={currentChatId ? 'Write a message' : 'Select or create chat first'}
						InputProps={{
							disabled: !currentChatId,
							disableUnderline: true,
							endAdornment: (
								<InputAdornment position="end">
									<IconButton disabled={!currentChatId} onClick={currentChatId ? this.sendMessage : undefined}>
										<Icon name='send' color={currentChatId ? 'primary' : 'disabled'} />
									</IconButton>
								</InputAdornment>
							),
						}}
					/>
				</div>
			</div>
		);
	}

	render() {
		const {classes} = this.props;
		const {createChatDialogOpen} = this.state;

		if (this.state.loading) {
			return <LoadingIndication />;
		}

		return (
			<div className={`page-content ${classes.container}`}>
				{this.renderChatList()}
				{this.renderMessageList()}
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

export default withStyles(styles)(Chat);
