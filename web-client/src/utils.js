function getUserName(user) {
	if (user) {
		return user.fullName || user.username || 'Anonymous';
	}

	return 'Anonymous';
}

export {
	getUserName,
};
