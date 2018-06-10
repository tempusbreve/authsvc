import React from 'react';

class Login extends React.Component {
  render() {
    return this.props.user
      ? <span className="nothing-here" />
      : <div>
        <h3>LOGIN PLEASE</h3>
      </div>
  }
}
export default Login;
