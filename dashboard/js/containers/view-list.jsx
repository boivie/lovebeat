import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { fetchViews } from '../actions'
import Views from '../components/views'

class ViewList extends Component {
  componentWillMount() {
    const { dispatch } = this.props
    dispatch(fetchViews())
  }

  render() {
    const { views, isFetching, lastUpdated } = this.props
    const isEmpty = views.length === 0

    return (
      <div>
        <div className="view-title">
          Views
        </div>
      {isEmpty
        ? (isFetching ? <h2>Loading...</h2> : <h2>Empty.</h2>)
        : <div style={{ opacity: isFetching ? 0.5 : 1 }}>
            <Views views={views}/>
          </div>
      }
      </div>
    )
  }
}

ViewList.propTypes = {
  views: PropTypes.array.isRequired,
  isFetching: PropTypes.bool.isRequired,
  lastUpdated: PropTypes.number,
  dispatch: PropTypes.func.isRequired
}

function mapStateToProps(state) {
  const { listOfViews } = state
  const {
    isFetching,
    lastUpdated,
    items: views
  } = listOfViews || {
    isFetching: true,
    items: []
  }

  return {
    views,
    isFetching,
    lastUpdated
  }
}

export default connect(mapStateToProps)(ViewList)
