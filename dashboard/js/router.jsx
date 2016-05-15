import React from 'react';
import { Router, Route, browserHistory, IndexRoute } from 'react-router';

import MainLayout from './containers/main-layout';
import Home from './components/home';
import ViewDetails from './containers/view-details';
import { Provider } from 'react-redux'
import configureStore from './store/'
import fetchViews from './actions/'

const store = configureStore()

export default (
  <Provider store={store}>
    <Router history={browserHistory}>
      <Route component={MainLayout}>
        <Route path="/" component={Home} />

        <Route path="views">
          <Route path=":viewId" component={ViewDetails} />
        </Route>

      </Route>
    </Router>
  </Provider>
);
