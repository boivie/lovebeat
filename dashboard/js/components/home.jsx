import React from 'react';
import { Link } from 'react-router';

const Home = React.createClass({
  render: function() {
    return (
      <div id="home">
        <h1>Welcome to Lovebeat</h1>
        <p>
          Have you ever had a nightly backup job fail, and it took you weeks 
          until you noticed it? Say hi to lovebeat, a zero-configuration 
          heartbeat monitor.
        </p>
        <p>
          We have a great chapter in our documentation 
          on <a target="_blank" href="http://lovebeat.readthedocs.io/en/latest/getting_started.html">getting started</a> with 
          Lovebeat.
        </p>
        <p>
          If you're impatient, you can try the following and see the results in the <Link to="/services">services page</Link>:
        </p>
        <pre>curl -d timeout=10 http://localhost:8080/api/services/example.service</pre>
        <p>
          As long as you keep running that command (within the 10 second timeout), the service will stay green. 
          But if you stop, the service will become red and say <b>ERROR</b>. When that happens, you can configure
          lovebeat to run a shell script, send an e-mail or post to Slack.
        </p>
      </div>
    );
  }
});

export default Home;
