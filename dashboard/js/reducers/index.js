import { combineReducers } from 'redux'
import {
  REQUEST_ALARM, RECEIVE_ALARM, REQUEST_ALARMS, RECEIVE_ALARMS,
  ADD_SERVICE, UPDATE_SERVICE, REMOVE_SERVICE, ADD_ALARM, UPDATE_ALARM,
  TOGGLE_SERVICE_CHECKED, REMOVE_ALARM
} from '../actions'

function services(state = {
  isFetching: false,
  didInvalidate: false,
  items: [],
  checked: {}
}, action) {
  switch (action.type) {
    case REQUEST_ALARM:
      return Object.assign({}, state, {
        isFetching: true,
        didInvalidate: false
      })
    case RECEIVE_ALARM:
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

function servicesByAlarm(state = { }, action) {
  switch (action.type) {
    case RECEIVE_ALARM:
    case REQUEST_ALARM:
    case TOGGLE_SERVICE_CHECKED:
      return Object.assign({}, state, {
        [action.alarmId]: services(state[action.alarmId], action)
      })
    case ADD_SERVICE:
    case UPDATE_SERVICE:
    case REMOVE_SERVICE:
      var upd = {}
      const in_alarms = action.service.in_alarms || []
      for (var i = 0; i < in_alarms.length; i++) {
        const alarmId = in_alarms[i]
        upd[alarmId] = services(state[alarmId], action)
      }
      return Object.assign({}, state, upd)
    default:
      return state
  }
}

function listOfAlarms(state = { isFetching: false, didInvalidate: false, items: []}, action) {
  switch (action.type) {
    case REQUEST_ALARMS:
      return Object.assign({}, state, {
        isFetching: true,
        didInvalidate: false
      })
    case RECEIVE_ALARMS:
      return Object.assign({}, state, {
        isFetching: false,
        didInvalidate: false,
        items: action.alarms,
        lastUpdated: action.receivedAt
      })
    case ADD_ALARM:
      return Object.assign({}, state, {
        items: [action.alarm, ...state.items]
      })
    case UPDATE_ALARM:
      return Object.assign({}, state, {
        items: state.items.map(v => v.name == action.alarm.name ? action.alarm : v)
      })
    case REMOVE_ALARM:
      return Object.assign({}, state, {
        items: state.items.filter(v => v.name != action.alarm.name)
      })
    default:
      return state
  }
}

const rootReducer = combineReducers({
  servicesByAlarm,
  listOfAlarms
})

export default rootReducer
