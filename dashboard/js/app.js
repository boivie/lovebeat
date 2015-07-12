'use strict';

var lovebeatApp = angular.module('lovebeatApp', [
  'ngRoute',
  'lovebeatControllers',
  'lovebeatServices',
  'ngWebSocket'
]).factory('LovebeatStream', ['$websocket', '$rootScope',
  function($websocket, $rootScope) {
    var loc = window.location,
      ws_uri;
    if (loc.protocol === "https:") {
      ws_uri = "wss:";
    } else {
      ws_uri = "ws:";
    }
    ws_uri += "//" + loc.host + "/ws";
    var dataStream = $websocket(ws_uri, null, {
      reconnectIfNotNormalClose: true
    });

    dataStream.onMessage(function(message) {
      var payload = JSON.parse(message.data)
      $rootScope.$broadcast("LovebeatStream::" + payload.m, payload.args);
    });

    var methods = {
      get: function() {
        dataStream.send(JSON.stringify({
          action: 'get'
        }));
      }
    };

    return methods;
  }
]);

lovebeatApp.config(['$routeProvider',
  function($routeProvider) {
    $routeProvider.
    when('/views/:viewId/services/', {
      templateUrl: 'partials/service-list.html',
      controller: 'ServiceListCtrl'
    }).
    when('/views', {
      templateUrl: 'partials/view-list.html',
      controller: 'ViewListCtrl'
    }).
    when('/services/:serviceId', {
      templateUrl: 'partials/edit-service.html',
      controller: 'EditServiceCtrl'
    }).
    when('/views/:viewId', {
      templateUrl: 'partials/edit-view.html',
      controller: 'EditViewCtrl'
    }).
    when('/add-service', {
      templateUrl: 'partials/add-service.html',
      controller: 'AddServiceCtrl'
    }).
    when('/add-view', {
      templateUrl: 'partials/add-view.html',
      controller: 'AddViewCtrl'
    }).
    otherwise({
      redirectTo: '/views'
    });
  }
]);

lovebeatApp.filter('delta_ago', function() {
  return function(milliseconds) {
    if (milliseconds <= 0)
      return "now";
    return juration.stringify(milliseconds / 1000, {
      format: 'micro',
      units: 2
    }) + " ago";
  }
});

lovebeatApp.filter('delta', function() {
  return function(milliseconds) {
    if (milliseconds < 0)
      return "not set";
    return juration.stringify(milliseconds / 1000);
  }
});
