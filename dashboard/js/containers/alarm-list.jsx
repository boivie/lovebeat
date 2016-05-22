import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { fetchAlarms } from '../actions'
import Alarms from '../components/alarms'

class AlarmList extends Component {
  componentWillMount() {
    const { dispatch } = this.props
    dispatch(fetchAlarms())
  }

  render() {
    const { alarms, isFetching, lastUpdated } = this.props
    const isEmpty = alarms.length === 0

    return (
      <div>
      {isEmpty
        ? (isFetching ? <h2>Loading...</h2> : <h2>Empty.</h2>)
        : <div style={{ opacity: isFetching ? 0.5 : 1 }}>
            <Alarms alarms={alarms}/>
          </div>
      }
      </div>
    )
  }
}

AlarmList.propTypes = {
  alarms: PropTypes.array.isRequired,
  isFetching: PropTypes.bool.isRequired,
  lastUpdated: PropTypes.number,
  dispatch: PropTypes.func.isRequired
}

function mapStateToProps(state) {
  const { listOfAlarms } = state
  const {
    isFetching,
    lastUpdated,
    items: alarms
  } = listOfAlarms || {
    isFetching: true,
    items: []
  }

  return {
    alarms,
    isFetching,
    lastUpdated
  }
}

export default connect(mapStateToProps)(AlarmList)
