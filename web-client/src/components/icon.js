import React from 'react';
import PropTypes from 'prop-types';
import Icon from '@material-ui/core/Icon';

const IconWrapper = ({name, ...otherProps}) => (
	<Icon {...otherProps}>{name}</Icon>
);

IconWrapper.propTypes = {
	name: PropTypes.string.isRequired,
};

export default IconWrapper;
