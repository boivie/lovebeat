import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { fetchAlarms } from '../actions'
import {Link} from 'react-router';
import AlarmList from '../containers/alarm-list'
import ReconnectingWebSocket from 'reconnectingwebsocket'


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
      var ws = new ReconnectingWebSocket(ws_uri);

      ws.onopen = function() {
        console.log("websocket OPEN")
        dispatch(fetchAlarms())
      }

      ws.onclose = function() {
        console.log("websocket CLOSED")
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
          <AlarmList/>
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
