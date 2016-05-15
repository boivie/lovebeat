import { combineReducers } from 'redux'
import {
  REQUEST_SERVICES, RECEIVE_SERVICES, REQUEST_VIEWS, RECEIVE_VIEWS,
  ADD_SERVICE, UPDATE_SERVICE, REMOVE_SERVICE, ADD_VIEW, UPDATE_VIEW,
  TOGGLE_SERVICE_CHECKED
} from '../actions'

function services(state = {
  isFetching: false,
  didInvalidate: false,
  items: [],
  checked: {}
}, action) {
  switch (action.type) {
    case REQUEST_SERVICES:
      return Object.assign({}, state, {
        isFetching: true,
        didInvalidate: false
      })
    case RECEIVE_SERVICES:
      return Object.assign({}, state, {
        isFetching: false,
        didInvalidate: false,
        items: action.services,
        lastUpdated: action.receivedAt
      })
    case ADD_SERVICE:
      return Object.assign({}, state, {
        items: [action.service, ...state.items]
      })
    case UPDATE_SERVICE:
      return Object.assign({}, state, {
        items: state.items.map(s => s.name == action.service.name ? action.service : s)
      })
    case REMOVE_SERVICE:
      const newChecked = Object.assign({}, state.checked)
      delete newChecked[action.service.name]
      return Object.assign({}, state, {
        items: state.items.filter(s => s.name != action.service.name),
        checked: newChecked
      })
    case TOGGLE_SERVICE_CHECKED:
      return Object.assign({}, state, {
        checked: Object.assign({}, state.checked,
          { [action.serviceName]: !state.checked[action.serviceName] })
      })
    default:
      return state
  }
}

function servicesByView(state = { }, action) {
  switch (action.type) {
    case RECEIVE_SERVICES:
    case REQUEST_SERVICES:
    case TOGGLE_SERVICE_CHECKED:
      return Object.assign({}, state, {
        [action.view]: services(state[action.view], action)
      })
    case ADD_SERVICE:
    case UPDATE_SERVICE:
    case REMOVE_SERVICE:
      var upd = {}
      for (var i = 0; i < action.service.in_views.length; i++) {
        const view = action.service.in_views[i]
        upd[view] = services(state[view], action)
      }
      return Object.assign({}, state, upd)
    default:
      return state
  }
}

function listOfViews(state = { isFetching: false, didInvalidate: false, items: []}, action) {
  switch (action.type) {
    case REQUEST_VIEWS:
      return Object.assign({}, state, {
        isFetching: true,
        didInvalidate: false
      })
    case RECEIVE_VIEWS:
      return Object.assign({}, state, {
        isFetching: false,
        didInvalidate: false,
        items: action.views,
        lastUpdated: action.receivedAt
      })
    case ADD_VIEW:
      return Object.assign({}, state, {
        items: [action.view, ...state.items]
      })
    case UPDATE_VIEW:
      return Object.assign({}, state, {
        items: state.items.map(v => v.name == action.view.name ? action.view : v)
      })
    default:
      return state
  }
}

const rootReducer = combineReducers({
  servicesByView,
  listOfViews
})

export default rootReducer
