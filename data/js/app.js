'use strict';

/* App Module */

var lovebeatApp = angular.module('lovebeatApp', [
  'ngRoute',
  'lovebeatControllers',
  'lovebeatServices'
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
        templateUrl: 'partials/service-detail.html',
        controller: 'ServiceDetailCtrl'
      }).
      when('/add-service', {
        templateUrl: 'partials/add-service.html',
        controller: 'AddServiceCtrl'
      }).
      otherwise({
        redirectTo: '/views'
      });
  }]);

lovebeatApp.filter('delta_ago', function() {
  return function(seconds) {
      if (seconds <= 0)
	  return "now";
      return juration.stringify(seconds, {format:'micro'}) + " ago";
  }
});

lovebeatApp.filter('delta', function() {
    return function(seconds) {
	if (seconds < 0)
	    return "not set";
	return juration.stringify(seconds);
  }
});
