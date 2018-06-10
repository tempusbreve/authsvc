import React from 'react';

export class LogoutControl extends React.Component {
  render() {
    return (
      <form action="/auth/logout/" method="POST">
        <span style={{ fontWeight: 'bold', marginRight: 10 }}>{this.props.user.name}</span>
        <input type="submit" name="submit" value="Logout" />
      </form>
    )
  }
}

export default LogoutControl;
