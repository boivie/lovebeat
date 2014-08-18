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
