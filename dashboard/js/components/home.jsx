import React from 'react';
import { Link } from 'react-router';

const Home = React.createClass({
  render: function() {
    return (
      <div id="home">
        <h1>Welcome to Lovebeat</h1>
        <p>
          For a flull-depth introduction on how to get started, head to <a target="_blank" href="http://lovebeat.readthedocs.io/en/latest/">our documentation</a>.
        </p>
        <p>Try any of these, and then go to the <Link to="/services">services page</Link> to see your heartbeat.</p>
        <h2>curl</h2>
        <pre>curl -d timeout=60 http://localhost:8080/api/services/example.service</pre>
        <h2>python</h2>
        <pre>import requests<br/>
        requests.post("http://localhost:8080/api/services/example.service", data=dict(timeout=60))</pre>
        <p>As long as you keep sending these commands at least once per minute, it will stay green. But when you wait more than one minute, it will turn into a red <b>ERROR</b> state.</p>
      </div>
    );
  }
});

export default Home;
