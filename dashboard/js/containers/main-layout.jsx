import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import {Link} from 'react-router';
import ViewList from '../containers/view-list'


class MainLayout extends Component {
  componentWillMount() {
    const { dispatch } = this.props
    if ("WebSocket" in window) {
      var loc = window.location, ws_uri;
      var path = "/" //loc.pathname.substr(0, loc.pathname.lastIndexOf('/')) + "/"
      if (loc.protocol === "https:") {
        ws_uri = "wss:";
      } else {
        ws_uri = "ws:";
      }
      ws_uri += "//" + loc.host + path + "ws";
      var ws = new WebSocket(ws_uri);

      ws.onopen = function() {
        console.log("websocket open")
      }

      ws.onmessage = function (evt) {
          const data = JSON.parse(evt.data)
          dispatch(data)
       };
    }
  }

  render() {
    return (
      <div className="application">
        <header id="top">
          <div className="brand">
            <svg className="icon icon-heartbeat"><use xlinkHref='#icon-heartbeat'/></svg> Lovebeat
          </div>
        </header>
        <div className="wrapper-main">
          <main id="main">
            {this.props.children}
          </main>
        </div>
        <aside id="left">
          <ViewList/>
        </aside>
      </div>
    );
  }
}


MainLayout.propTypes = {
  dispatch: PropTypes.func.isRequired,
}

function mapStateToProps(state) {
  return { }
}

export default connect(mapStateToProps)(MainLayout)
