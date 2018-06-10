import React from 'react';
import { parse } from 'qs';

class OAuthAsk extends React.Component {
  render() {
    const qp = parse(this.props.location.search, { ignoreQueryPrefix: true });
    return qp.id ?
      <div>
        <h3>Allow {qp.app || 'application'} access?</h3>
        <form action="/oauth/approve" method="post">
          <input type="hidden" name="corr" value={qp.id} />
          <input type="submit" name="approve" value="Approve" />
          <input type="submit" name="deny" value="Deny" />
        </form>
      </div> : <span className="nothing-here" />
  }
}

export default OAuthAsk;
