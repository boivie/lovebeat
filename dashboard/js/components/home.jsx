import React from 'react';

const Home = React.createClass({
  render: function() {
    return (
      <div className="home-page">
        <h1>Welcome to Lovebeat</h1>
        <p>
          There will soon be more instructions here on how to get started.
        </p>
        <p>
          For now, head to <a target="_blank" href="http://lovebeat.readthedocs.io/en/latest/">our documentation</a> and give it a spin!
        </p>
      </div>
    );
  }
});

export default Home;
