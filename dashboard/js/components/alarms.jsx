import React, { PropTypes, Component } from 'react'
import { Link } from 'react-router';
import classNames from 'classnames';

function compare(a,b) {
  if (a.name < b.name)
    return -1;
  else if (a.name > b.name)
    return 1;
  else
    return 0;
}

export default class Alarms extends Component {
  render() {
    const alarms = this.props.alarms
    alarms.sort(compare)
    var elems = []
    for (var i = 0; i < alarms.length; i++) {
      var alarm = alarms[i]
      var url = "/alarms/" + alarm.name
      var tileClasses = classNames({
        'alarm-tile': true,
        'ok': alarm.state == 'ok',
        'error': alarm.state == 'error'
      })
      var failingServices = null
      if (alarm.failed_services != null && alarm.failed_services.length > 0) {
        failingServices = (<p>
          <span className="icon-clock">{alarm.failed_services.length} services in error</span>
        </p>)
      }
      elems.push(
        <li key={i} className="alarm-li">
          <div className={tileClasses}>
            <h2 className="title">
              <Link to={url}>
                <span className="label-align">{alarm.name}</span>
              </Link>
            </h2>
          </div>

        </li>)
    }
    return <ul className="alarms">{elems}</ul>
  }
}

Alarms.propTypes = {
  alarms: PropTypes.array.isRequired
}
