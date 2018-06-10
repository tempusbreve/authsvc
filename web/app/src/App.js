import React from 'react';
import { Router, Route } from 'react-router';
import createBrowserHistory from 'history/createBrowserHistory';
import axios from 'axios';
import Account from './Account';
import OAuthAsk from './OAuthAsk';
import Login from './Login';
import './App.css';

class App extends React.Component {
  constructor(props) {
    super(props);
    this.history = createBrowserHistory();
    this.state = {};
  }

  componentDidMount() {
    axios.get('/api/v4/user').then(res => {
      this.setState({ user: res.data });
    }).catch(() => {
      this.setState({ user: null });
    })
  }

  render() {
    const appName = process.env.REACT_APP_NAME || 'Authsvc';
    return (
      <Router history={this.history}>
        <div>
          <header style={{ width: '100%', background: '#333', color: '#eee', margin: 0, padding: 10 }}>
            <h1 style={{ margin: 0, padding: 0 }}>{appName}</h1>
          </header>
          <Route path="/" render={props => {return <Account user={this.state.user} {...props} />}} />
          <Route path="/oauth/ask" exact={true} component={OAuthAsk} />
          <Route path="/auth/login/" exact={true} render={() => {return <Login user={this.state.user} />}} />
        </div>
      </Router>
    );
  }
}

export default App;
