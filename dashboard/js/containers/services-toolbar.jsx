import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { triggerBeat, muteService, unmuteService, deleteService } from '../actions'

class ServicesToolbar extends Component {
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
    const { checked } = this.props
    var enabled = false
    Object.keys(checked).forEach(key => enabled |= checked[key])

    return (
      <div className="toolbar">
        <button onClick={this.trigger.bind(this)} disabled={!enabled} title="Trigger" className="tool-btn"><svg className="btn-icon"><use xlinkHref='#icon-heartbeat'/></svg></button>
        <button onClick={this.mute.bind(this)} disabled={!enabled} title="Mute" className="tool-btn"><svg className="btn-icon"><use xlinkHref='#icon-mute'/></svg></button>
        <button onClick={this.unmute.bind(this)} disabled={!enabled} title="Unmute" className="tool-btn"><svg className="btn-icon"><use xlinkHref='#icon-unmute'/></svg></button>
        <button onClick={this.deleteService.bind(this)} disabled={!enabled} title="Delete" className="tool-btn"><svg className="btn-icon"><use xlinkHref='#icon-delete'/></svg></button>
      </div>
    )
  }
}

ServicesToolbar.propTypes = {
  checked: PropTypes.object.isRequired
}

function mapStateToProps(state, ownProps) {
  return {}
}

export default connect(mapStateToProps, {
  triggerBeat, muteService, unmuteService, deleteService
})(ServicesToolbar)
