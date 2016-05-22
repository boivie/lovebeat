import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { fetchAlarm, triggerBeat, muteService, unmuteService, deleteService } from '../actions'
import Services from './services'

function loadData(props) {
  const { alarmId } = props
  props.fetchAlarm(alarmId)
}

class AlarmDetails extends Component {
  componentWillMount() {
    loadData(this.props)
  }

  componentWillReceiveProps(nextProps) {
    if (nextProps.alarmId !== this.props.alarmId) {
      loadData(nextProps)
    }
  }

  forEachChecked(fn) {
    const { checked } = this.props
    Object.keys(checked).forEach(key => checked[key] && fn(key))
  }

  trigger() {
    this.forEachChecked(this.props.triggerBeat)
  }

  mute() {
    this.forEachChecked(this.props.muteService)
  }

  unmute() {
    this.forEachChecked(this.props.unmuteService)
  }

  deleteService() {
    this.forEachChecked(this.props.deleteService)
  }

  render() {
    const { checked, services, isFetching, lastUpdated } = this.props
    const isEmpty = services.length === 0
    var enabled = false
    Object.keys(checked).forEach(key => enabled |= checked[key])

    return (
      <div>
      <h1>{this.props.alarmId}</h1>
      <div className="toolbar">
        <button onClick={this.trigger.bind(this)} disabled={!enabled} title="Trigger" className="tool-btn"><svg className="btn-icon"><use xlinkHref='#icon-heartbeat'/></svg></button>
        <button onClick={this.mute.bind(this)} disabled={!enabled} title="Mute" className="tool-btn"><svg className="btn-icon"><use xlinkHref='#icon-mute'/></svg></button>
        <button onClick={this.unmute.bind(this)} disabled={!enabled} title="Unmute" className="tool-btn"><svg className="btn-icon"><use xlinkHref='#icon-unmute'/></svg></button>
        <button onClick={this.deleteService.bind(this)} disabled={!enabled} title="Delete" className="tool-btn"><svg className="btn-icon"><use xlinkHref='#icon-delete'/></svg></button>
      </div>
      {isEmpty
        ? (isFetching ? <h2>Loading...</h2> : <h2>Empty.</h2>)
        : <div>
            <Services key={this.props.alarmId} alarmId={this.props.alarmId} services={services} checked={checked}/>
          </div>
      }
      </div>
    )
  }
}

AlarmDetails.propTypes = {
  alarmId: PropTypes.string.isRequired,
  services: PropTypes.array.isRequired,
  isFetching: PropTypes.bool.isRequired,
  lastUpdated: PropTypes.number,
  fetchAlarm: PropTypes.func.isRequired,
  checked: PropTypes.object.isRequired
}

function mapStateToProps(state, ownProps) {
  const { servicesByAlarm } = state
  const alarmId = ownProps.params.alarmId

  const {
    isFetching,
    lastUpdated,
    items: services,
    checked
  } = servicesByAlarm[alarmId] || {
    isFetching: true,
    items: [],
    checked: {}
  }

  return {
    alarmId,
    services,
    isFetching,
    lastUpdated,
    checked
  }
}

export default connect(mapStateToProps, {
  fetchAlarm, triggerBeat, muteService, unmuteService, deleteService
})(AlarmDetails)
