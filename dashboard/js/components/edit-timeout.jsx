import React, { Component, PropTypes } from 'react'
import classnames from 'classnames'

class EditTimeout extends Component {
  constructor(props, context) {
    super(props, context)
    this.state = {
      text: this.props.text || ''
    }
  }

  handleSubmit(e) {
    const text = e.target.value.trim()
    if (e.which === 13) {
      this.props.onSave(text)
    }
  }

  handleChange(e) {
    this.setState({ text: e.target.value })
  }

  handleBlur(e) {
    this.props.onSave(e.target.value)
  }

  handleFocus(e) {
    var target = e.target;
    setTimeout(function() {
      target.select();
    }, 0);
  }

  render() {
    return (
      <input className="timeout-input"
        type="text"
        placeholder={this.props.placeholder}
        autoFocus="true"
        value={this.state.text}
        onBlur={this.handleBlur.bind(this)}
        onChange={this.handleChange.bind(this)}
        onKeyDown={this.handleSubmit.bind(this)}
        onFocus={this.handleFocus}/>
    )
  }
}

EditTimeout.propTypes = {
  onSave: PropTypes.func.isRequired,
  text: PropTypes.string,
  placeholder: PropTypes.string
}

export default EditTimeout
