import React from 'react';
import { parse } from 'qs';
import LoginControl from './LoginControl';
import LogoutControl from './LogoutControl';


class Account extends React.Component {
  render() {
    const q = this.props.location ? parse(this.props.location.search, { ignoreQueryPrefix: true }) : {};
    return (
      <div style={{ width: '100%', background: '#ffe', padding: 3, display: 'flex', flexDirection: 'row-reverse' }}>
        <div style={{ order: 1, flex: 1 }}>{q.msg}</div>
        <div style={{ order: 0, paddingRight: 10 }}>
          {this.props.user
            ? (<LogoutControl user={this.props.user} />)
            : (<LoginControl redir={q.redirect_uri} />)}
        </div>
      </div>);
  }
}

export default Account;
