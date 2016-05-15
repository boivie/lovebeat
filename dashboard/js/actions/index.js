export const REQUEST_SERVICES = 'REQUEST_SERVICES'
export const RECEIVE_SERVICES = 'RECEIVE_SERVICES'
export const REQUEST_VIEWS = 'REQUEST_VIEWS'
export const RECEIVE_VIEWS = 'RECEIVE_VIEWS'
export const ADD_SERVICE = 'ADD_SERVICE'
export const UPDATE_SERVICE = 'UPDATE_SERVICE'
export const REMOVE_SERVICE = 'REMOVE_SERVICE'
export const ADD_VIEW = 'ADD_VIEW'
export const UPDATE_VIEW = 'UPDATE_VIEW'
export const SET_TIMEOUT = 'SET_TIMEOUT'
export const TOGGLE_SERVICE_CHECKED = 'TOGGLE_SERVICE_CHECKED'
export const TRIGGER_BEAT = 'TRIGGER_BEAT'
export const MUTE_SERVICE = 'MUTE_SERVICE'
export const UNMUTE_SERVICE = 'UNMUTE_SERVICE'
export const DELETE_SERVICE = 'DELETE_SERVICE'

function requestServices(view) {
  return {
    type: REQUEST_SERVICES,
    view
  }
}

export function toggleServiceChecked(view, serviceName) {
  return dispatch => {
      console.log("TOOGLE", view, serviceName)
      dispatch({
      type: TOGGLE_SERVICE_CHECKED,
      view,
      serviceName
    })
  }
}

function receiveServices(view, json) {
  return {
    type: RECEIVE_SERVICES,
    view,
    services: json.services,
    serverTs: json.now,
    receivedAt: Date.now()
  }
}

export function fetchServices(view) {
  return dispatch => {
    dispatch(requestServices(view))
    return fetch(`/api/services/?view=${view}`)
      .then(response => response.json())
      .then(json => dispatch(receiveServices(view, json)))
  }
}

function requestViews() {
  return {
    type: REQUEST_VIEWS
  }
}

function receiveViews(json) {
  return {
    type: RECEIVE_VIEWS,
    views: json.views,
    serverTs: json.now,
    receivedAt: Date.now()
  }
}

export function fetchViews() {
  return dispatch => {
    dispatch(requestViews())
    return fetch(`/api/views/`)
      .then(response => response.json())
      .then(json => dispatch(receiveViews(json)))
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
