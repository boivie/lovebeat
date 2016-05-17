import React, { PropTypes, Component } from 'react'
import { Link } from 'react-router';
import classNames from 'classnames';
import juration from 'juration';
import moment from 'moment';
import EditTimeout from './edit-timeout'
import ReactCSSTransitionGroup from 'react-addons-css-transition-group'

function humanDateTime(d) {
  const m = moment(d)
  if (m.isSame(moment(), 'day')) {
    return m.format("LTS")
  } else {
    return m.format("lll")
  }
}

function tohuman(d) {
  return juration.stringify(d, { format: 'short' })
}

export default class Service extends Component {
  constructor(props, context) {
    super(props, context)
    this.state = {
      editTimeout: false
    }
  }

  handleDoubleClick() {
    this.setState({ editTimeout: true })
  }

  handleSetTimeout(id, text) {
    let duration = juration.parse(text)
    this.props.setServiceTimeout(id, duration)
    this.setState({ editTimeout: false })
  }

  render() {
    const service = this.props.service
    var tileClasses = classNames({
      'service-tile': true,
      'ok': service.state == 'ok',
      'error': service.state == 'error',
      'muted': service.state == 'muted'
    })
    let icon
    let subtitle
    if (service.state == 'ok') {
      icon = (<svg className="icon icon-checkmark"><use xlinkHref='#icon-checkmark'/></svg>)
      subtitle = null
    } else if (service.state == 'error') {
      icon = (<svg className="icon icon-cross"><use xlinkHref='#icon-cross'/></svg>)
      if (service.last_beat > 0) {
        subtitle = "No heartbeats since " + humanDateTime(service.last_beat)
      } else {
        subtitle = "No heartbeats have ever been seen"
      }
    } else if (service.state == 'muted') {
      icon = (<svg className="icon icon-mute"><use xlinkHref='#icon-mute'/></svg>)
      subtitle = "Muted since " + humanDateTime(service.muted_since)
    } else {
      subtitle = "Unknown state!"
    }
    const subtitleComponent = subtitle == null ? "" : subtitle
    let timeout
    if (service.timeout == 0) {
      timeout = "always error"
    } else if (service.timeout == -1) {
      timeout = "always ok"
    } else if (service.timeout < 0) {
      timeout = "unknown"
    } else {
      timeout = juration.stringify(service.timeout / 1000, { format: 'short' });
    }

    let timeoutComponent
    if (this.state.editTimeout) {
      timeoutComponent = (<div>
        <svg className="icon icon-clock"><use xlinkHref='#icon-clock'/></svg>
        <EditTimeout text={timeout} onSave={(text) => this.handleSetTimeout(service.name, text)}/>
        </div>
      )
    } else {
      timeoutComponent = (<div title="Double-click to edit" onDoubleClick={this.handleDoubleClick.bind(this)}>
        <svg className="icon icon-clock"><use xlinkHref='#icon-clock'/></svg>
        <span className="label-align">{timeout}</span>
      </div>)
    }
    const checked = this.props.checked ? "✔" : ""
    let beatAnalysis
    if (service.analysis) {
      if (service.analysis.unstable) {
        beatAnalysis = (<span className="unstable">unstable ({tohuman(service.analysis.lower)} &ndash; {tohuman(service.analysis.upper)})</span>)
      } else if (service.analysis.upper - service.analysis.lower <= 10) {
        beatAnalysis = (<span>{tohuman(service.analysis.upper)}</span>)
      } else {
        beatAnalysis = (<span>{tohuman(service.analysis.lower)} &ndash; {tohuman(service.analysis.upper)}</span>)
      }
    } else {
      beatAnalysis = (<span>-</span>)
    }

    return (<li className="service-li">
        <div className={tileClasses}>
          <div className="section section1">
            <div className="section-row">
              <h2 className="title">
                <div className="checkbox" onClick={this.props.toggleChecked}>{checked}</div>
                {icon}
                <span className="label-align">{service.name}</span>
              </h2>
            </div>
            <div className="section-row">
              <div className="subtitle">{subtitleComponent}</div>
            </div>
          </div>
          <div className="section section2">
            {timeoutComponent}
          </div>
          <div className="section section2">
            <div className="section-row">
              <svg className="icon icon-calendar"><use xlinkHref='#icon-calendar'/></svg>
              <ReactCSSTransitionGroup transitionName="lastbeat" transitionEnterTimeout={500} transitionLeave={false}>
                <span key={service.last_beat} className="label-align beat-analysis">{humanDateTime(service.last_beat)}</span>
              </ReactCSSTransitionGroup>
            </div>
            <div className="section-row">
              <svg className="icon icon-stopwatch"><use xlinkHref='#icon-stopwatch'/></svg>
              <span className="label-align beat-analysis">{beatAnalysis}</span>
            </div>
          </div>
        </div>
      </li>)
  }
}

Service.propTypes = {
  service: PropTypes.object.isRequired,
  checked: PropTypes.bool.isRequired,
  setServiceTimeout: PropTypes.func.isRequired,
  toggleChecked: PropTypes.func.isRequired
}
