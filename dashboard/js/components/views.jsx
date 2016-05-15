import React, { PropTypes, Component } from 'react'
import { Link } from 'react-router';
import classNames from 'classnames';

function compare(a,b) {
  if (a.name < b.name)
    return -1;
  else if (a.name > b.name)
    return 1;
  else
    return 0;
}

export default class Views extends Component {
  render() {
    const views = this.props.views
    views.sort(compare)
    var elems = []
    for (var i = 0; i < views.length; i++) {
      var view = views[i]
      var url = "/views/" + view.name
      var tileClasses = classNames({
        'view-tile': true,
        'ok': view.state == 'ok',
        'error': view.state == 'error'
      })
      var failingServices = null
      if (view.failed_services != null && view.failed_services.length > 0) {
        failingServices = (<p>
          <span className="icon-clock">{view.failed_services.length} services in error</span>
        </p>)
      }
      elems.push(
        <li key={i} className="view-li">
          <div className={tileClasses}>
            <h2 className="title">
              <Link to={url}>
                <span className="label-align">{view.name}</span>
              </Link>
            </h2>
          </div>

        </li>)
    }
    return <ul className="views">{elems}</ul>
  }
}

Views.propTypes = {
  views: PropTypes.array.isRequired
}
