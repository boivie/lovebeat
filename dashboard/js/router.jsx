import React from 'react';
import { Router, Route, browserHistory, IndexRoute } from 'react-router';

import MainLayout from './containers/main-layout';
import Home from './components/home';
import AlarmDetails from './containers/alarm-details';
import AllServices from './containers/all-services'
import { Provider } from 'react-redux'
import configureStore from './store/'

const store = configureStore()

export default (
  <Provider store={store}>
    <Router history={browserHistory}>
      <Route component={MainLayout}>
        <Route path="/" component={Home} />
        <Route path="services" component={AllServices} />
        <Route path="alarms">
          <Route path=":alarmId" component={AlarmDetails} />
        </Route>

      </Route>
    </Router>
  </Provider>
);
