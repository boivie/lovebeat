import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { fetchAllServices } from '../actions'
import Services from './services'
import ServicesToolbar from './services-toolbar'


class AllServices extends Component {
  componentWillMount() {
    this.props.fetchAllServices()
  }

  render() {
    const { checked, services, isFetching, lastUpdated } = this.props
    const isEmpty = services.length === 0

    return (
      <div>
      <h1>All Services</h1>
      <ServicesToolbar checked={checked}/>
      {isEmpty
        ? (isFetching ? <h2>Loading...</h2> : <p>Sorry - we couldn't find any services. Start sending heartbeats and they'll turn up here.</p>)
        : <div>
            <Services key="all-services" services={services} checked={checked}/>
          </div>
      }
      </div>
    )
  }
}

AllServices.propTypes = {
  services: PropTypes.array.isRequired,
  isFetching: PropTypes.bool.isRequired,
  lastUpdated: PropTypes.number,
  fetchAllServices: PropTypes.func.isRequired,
  checked: PropTypes.object.isRequired
}

function mapStateToProps(state, ownProps) {
  const { servicesByAlarm } = state

  const {
    isFetching,
    lastUpdated,
    items: services,
    checked
  } = servicesByAlarm[""] || {
    isFetching: true,
    items: [],
    checked: {}
  }

  return {
    services,
    isFetching,
    lastUpdated,
    checked
  }
}

export default connect(mapStateToProps, {
  fetchAllServices
})(AllServices)
