export const REQUEST_ALARM = 'REQUEST_ALARM'
export const RECEIVE_ALARM = 'RECEIVE_ALARM'
export const REQUEST_ALARMS = 'REQUEST_ALARMS'
export const RECEIVE_ALARMS = 'RECEIVE_ALARMS'
export const REQUEST_ALL_SERVICES = 'REQUEST_ALL_SERVICES'
export const RECEIVE_ALL_SERVICES = 'RECEIVE_ALL_SERVICES'
export const ADD_SERVICE = 'ADD_SERVICE'
export const UPDATE_SERVICE = 'UPDATE_SERVICE'
export const REMOVE_SERVICE = 'REMOVE_SERVICE'
export const ADD_ALARM = 'ADD_ALARM'
export const UPDATE_ALARM = 'UPDATE_ALARM'
export const REMOVE_ALARM = 'REMOVE_ALARM'
export const SET_TIMEOUT = 'SET_TIMEOUT'
export const TOGGLE_SERVICE_CHECKED = 'TOGGLE_SERVICE_CHECKED'
export const TRIGGER_BEAT = 'TRIGGER_BEAT'
export const MUTE_SERVICE = 'MUTE_SERVICE'
export const UNMUTE_SERVICE = 'UNMUTE_SERVICE'
export const DELETE_SERVICE = 'DELETE_SERVICE'

function requestAlarm(alarmId) {
  return {
    type: REQUEST_ALARM,
    alarmId
  }
}

export function toggleServiceChecked(alarmId, serviceName) {
  return dispatch => {
      dispatch({
      type: TOGGLE_SERVICE_CHECKED,
      alarmId,
      serviceName
    })
  }
}

function receiveAlarm(alarmId, json) {
  return {
    type: RECEIVE_ALARM,
    alarmId,
    alarm: json.alarm,
    services: json.services,
    serverTs: json.now,
    receivedAt: Date.now()
  }
}

export function fetchAlarm(alarmId) {
  return dispatch => {
    dispatch(requestAlarm(alarmId))
    return fetch(`/api/alarms/${alarmId}`)
      .then(response => response.json())
      .then(json => dispatch(receiveAlarm(alarmId, json)))
  }
}

function requestAllServices() {
  return {
    type: REQUEST_ALL_SERVICES
  }
}

function receiveAllServices(json) {
  return {
    type: RECEIVE_ALL_SERVICES,
    services: json.services,
    receivedAt: Date.now()
  }
}

export function fetchAllServices() {
  return dispatch => {
    dispatch(requestAllServices())
    return fetch(`/api/services`)
      .then(response => response.json())
      .then(json => dispatch(receiveAllServices(json)))
  }
}

function requestAlarms() {
  return {
    type: REQUEST_ALARMS
  }
}

function receiveAlarms(json) {
  return {
    type: RECEIVE_ALARMS,
    alarms: json.alarms,
    serverTs: json.now,
    receivedAt: Date.now()
  }
}

export function fetchAlarms() {
  return dispatch => {
    dispatch(requestAlarms())
    return fetch(`/api/alarms/`)
      .then(response => response.json())
      .then(json => dispatch(receiveAlarms(json)))
  }
}

export const setServiceTimeout = (serviceName, timeout) => {
  return dispatch => {
    dispatch({
      type: SET_TIMEOUT,
      serviceName,
      timeout
    })
    return fetch('/api/services/' + serviceName, {
      method: 'post',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({timeout, beat: false})
    })
      .then(response => response.json())
      .then(json => console.log(json))
  }
}

export const triggerBeat = (serviceName) => {
  return dispatch => {
    dispatch({
      type: TRIGGER_BEAT,
      serviceName
    })
    return fetch('/api/services/' + serviceName, {
      method: 'post',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({beat: true})
    })
      .then(response => response.json())
      .then(json => console.log(json))
  }
}


export const muteService = (serviceName) => {
  return dispatch => {
    dispatch({
      type: MUTE_SERVICE,
      serviceName
    })
    return fetch('/api/services/' + serviceName + '/mute', {
      method: 'post',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({})
    })
      .then(response => response.json())
      .then(json => console.log(json))
  }
}


export const unmuteService = (serviceName) => {
  return dispatch => {
    dispatch({
      type: UNMUTE_SERVICE,
      serviceName
    })
    return fetch('/api/services/' + serviceName + '/unmute', {
      method: 'post',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({})
    })
      .then(response => response.json())
      .then(json => console.log(json))
  }
}


export const deleteService = (serviceName) => {
  return dispatch => {
    dispatch({
      type: DELETE_SERVICE,
      serviceName
    })
    return fetch('/api/services/' + serviceName, {
      method: 'delete',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      }
    })
      .then(response => response.json())
      .then(json => console.log(json))
  }
}
