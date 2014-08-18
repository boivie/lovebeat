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
      otherwise({
        redirectTo: '/views'
      });
  }]);

lovebeatApp.filter('delta_ago', function() {
  return function(seconds) {
    return moment.duration(-1 * seconds * 1000).humanize(true);
  }
});

lovebeatApp.filter('delta', function() {
    return function(seconds) {
	if (seconds < 0)
	    return "not set";
	var output = []
	var numyears = Math.floor(seconds / 31536000);
	if (numyears > 0) {
	    output.push(numyears + " yrs")
	}
	var numdays = Math.floor((seconds % 31536000) / 86400);
	if (numdays > 0) {
	    output.push(numdays + " days")
	}
	var numhours = Math.floor(((seconds % 31536000) % 86400) / 3600);
	if (numhours > 0) {
	    output.push(numhours + " hrs")
	}
	var numminutes = Math.floor((((seconds % 31536000) % 86400) % 3600) / 60);
	if (numminutes > 0) {
	    output.push(numminutes + " min")
	}
	var numseconds = (((seconds % 31536000) % 86400) % 3600) % 60;
	if (numseconds > 0) {
	    output.push(numseconds + " sec")
	}
	return output.join(", ")
  }
});
