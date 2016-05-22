import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { fetchAlarm } from '../actions'
import Services from './services'
import ServicesToolbar from './services-toolbar'

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

  render() {
    const { checked, services, isFetching, lastUpdated } = this.props
    const isEmpty = services.length === 0
    const pathSepStyles = {margin: "0 0.25em"}
    return (
      <div>
      <h1><span>alarms</span><span style={pathSepStyles}>/</span><span>{this.props.alarmId}</span></h1>
      <ServicesToolbar checked={checked}/>
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
  fetchAlarm
})(AlarmDetails)
