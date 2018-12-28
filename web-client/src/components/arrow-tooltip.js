import React from 'react';
import Tooltip from '@material-ui/core/Tooltip';
import {withStyles} from '@material-ui/core/styles';

function arrowGenerator(color) {
	return {
		'&[x-placement*="bottom"] $arrow': {
			top: 0,
			left: 0,
			marginTop: '-0.9em',
			width: '3em',
			height: '1em',
			'&::before': {
				borderWidth: '0 1em 1em 1em',
				borderColor: `transparent transparent ${color} transparent`,
			},
		},
		'&[x-placement*="top"] $arrow': {
			bottom: 0,
			left: 0,
			marginBottom: '-0.9em',
			width: '3em',
			height: '1em',
			'&::before': {
				borderWidth: '1em 1em 0 1em',
				borderColor: `${color} transparent transparent transparent`,
			},
		},
		'&[x-placement*="right"] $arrow': {
			left: 0,
			marginLeft: '-0.9em',
			height: '3em',
			width: '1em',
			'&::before': {
				borderWidth: '1em 1em 1em 0',
				borderColor: `transparent ${color} transparent transparent`,
			},
		},
		'&[x-placement*="left"] $arrow': {
			right: 0,
			marginRight: '-0.9em',
			height: '3em',
			width: '1em',
			'&::before': {
				borderWidth: '1em 0 1em 1em',
				borderColor: `transparent transparent transparent ${color}`,
			},
		},
	};
}

const styles = theme => ({
	arrowPopper: arrowGenerator(theme.palette.grey[700]),
	arrow: {
		position: 'absolute',
		fontSize: 7,
		width: '3em',
		height: '3em',
		'&::before': {
			content: '""',
			margin: 'auto',
			display: 'block',
			width: 0,
			height: 0,
			borderStyle: 'solid',
		},
	},
});

class ArrowTooltip extends React.Component {
	constructor(props) {
		super(props);

		this.state = {
			arrowRef: null,
		};

		this.handleArrowRef = this.handleArrowRef.bind(this);
	}

	handleArrowRef(node) {
		this.setState({
			arrowRef: node,
		});
	};

	render() {
		const {classes, tooltip, children} = this.props;

		return (
			<Tooltip
				title={(
					<React.Fragment>
						{tooltip}
						<span className={classes.arrow} ref={this.handleArrowRef} />
					</React.Fragment>
				)}
				classes={{popper: classes.arrowPopper}}
				PopperProps={{
					popperOptions: {
						modifiers: {
							arrow: {
								enabled: Boolean(this.state.arrowRef),
								element: this.state.arrowRef,
							},
						},
					},
				}}
			>
				{children}
			</Tooltip>
		);
	}
};

export default withStyles(styles)(ArrowTooltip);
