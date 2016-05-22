import React, { PropTypes, Component } from 'react'
import Alarm from './alarm'

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
    const { alarms } = this.props
    alarms.sort(compare)
    return <ul className="alarms">{alarms.map(a => <Alarm key={a.name} alarm={a}/>)}</ul>
  }
}

Alarms.propTypes = {
  alarms: PropTypes.array.isRequired
}
