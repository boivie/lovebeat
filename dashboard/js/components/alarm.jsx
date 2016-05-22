import React, { PropTypes, Component } from 'react'
import { Link } from 'react-router';
import classNames from 'classnames';

export default class Alarm extends Component {
  render() {
    const { alarm } = this.props
    const url = "/alarms/" + alarm.name
    let tileClasses = classNames({
      'alarm-tile': true,
      'ok': alarm.state == 'ok',
      'error': alarm.state == 'error'
    })
    let failingServices = null
    if (alarm.failed_services != null && alarm.failed_services.length > 0) {
      failingServices = (<p>
        <span className="icon-clock">{alarm.failed_services.length} services in error</span>
      </p>)
    }
    return (
      <li className="alarm-li">
        <div className={tileClasses}>
          <h2 className="title">
            <Link to={url}>
              <span className="label-align">{alarm.name}</span>
            </Link>
          </h2>
        </div>
      </li>)
  }
}

Alarm.propTypes = {
  alarm: PropTypes.object.isRequired
}
