import React from 'react';
import CircularProgress from '@material-ui/core/CircularProgress';

export default () => (
	<div style={{display: 'flex', flex: 1, flexDirection: 'column', justifyContent: 'center', alignItems: 'center'}}>
		<CircularProgress color='primary' />
	</div>
);
