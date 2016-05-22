import React, { PropTypes, Component } from 'react'
import { connect } from 'react-redux'
import Service from '../components/service'
import { setServiceTimeout, toggleServiceChecked } from '../actions/'
import ReactCSSTransitionGroup from 'react-addons-css-transition-group'

function compare(a,b) {
  if (a.name < b.name)
    return -1;
  else if (a.name > b.name)
    return 1;
  else
    return 0;
}

class Services extends Component {

  render() {
    const services = this.props.services
    services.sort(compare)
    return (<ul className="services">
      <ReactCSSTransitionGroup transitionName="services" transitionEnterTimeout={500} transitionLeaveTimeout={400}>
        {services.map(service => <Service key={service.name} service={service} checked={this.props.checked[service.name] || false}
          toggleChecked={() => this.props.toggleServiceChecked(this.props.alarmId, service.name)}
          setServiceTimeout={this.props.setServiceTimeout}/>)}
      </ReactCSSTransitionGroup>
    </ul>)
  }
}

Services.propTypes = {
  alarmId: PropTypes.string.isRequired,
  services: PropTypes.array.isRequired,
  checked: PropTypes.object.isRequired,
  setServiceTimeout: PropTypes.func.isRequired,
  toggleServiceChecked: PropTypes.func.isRequired
}

function mapStateToProps(state) {
  return {}
}

export default connect(mapStateToProps, { setServiceTimeout, toggleServiceChecked })(Services)
