import React from 'react';

export class LoginControl extends React.Component {
  render() {
    return (
      <div>
        <form action="/auth/login/" method="POST">
          <input type="hidden" name="redirect_uri" value={this.props.redir} />
          <input type="text" placeholder="username" name="username" />
          <input type="password" placeholder="password" name="password" />
          <input type="submit" name="submit" value="Login" />
        </form>
      </div>
    )
  }
}
export default LoginControl;
