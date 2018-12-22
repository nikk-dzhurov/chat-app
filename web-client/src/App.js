import React, { useReducer } from 'react';
import PropTypes from 'prop-types';

import Switch from '@material-ui/core/Switch';
import {MuiThemeProvider, createMuiTheme, withTheme} from '@material-ui/core/styles';

const themes = {
	light: createMuiTheme({
		palette: {
			type: 'light',
		},
	}),
	dark: createMuiTheme({
		palette: {
			type: 'dark',
		},
	}),
};

export default class App extends React.Component {
	constructor(props) {
		super(props);

		this.state = {
			currentUser: null,
			checked: false,
		};
	}

	componentDidMount() {
		document.title = 'Chat App';
	}

	render() {
		return (
			<MuiThemeProvider theme={themes[this.state.themeKey]}>
				<BarT checked={this.state.checked} onChange={(e, checked) => this.setState({themeKey: checked ? 'light' : 'dark'})} />
			</MuiThemeProvider>
		);
	}
}
App.childContextTypes = {
	currentUser: PropTypes.object,
};

const Bar = (props, {muiTheme}) => {
	return (
		<div style={{background: props.theme.palette.background}}>
			Choose theme:<br />
			Dark <Switch checked={props.checked} onChange={props.onChange} /> Light
		</div>
	)
}

const BarT = withTheme()(Bar);
