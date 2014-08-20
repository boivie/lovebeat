'use strict';

/* Controllers */

var lovebeatControllers = angular.module('lovebeatControllers', []);

lovebeatControllers.controller('ServiceListCtrl', ['$scope', '$routeParams', 'Service',
  function($scope, $routeParams, Service) {
      $scope.services = Service.query({viewId: $routeParams.viewId});
      $scope.viewName = $routeParams.viewId;
      $scope.lbTrigger = function (service) {
	  service.$trigger();
	  service = service.$get();
      }
  }]);

lovebeatControllers.controller('ViewListCtrl', ['$scope', 'View',
  function($scope, View) {
    $scope.views = View.query();
  }]);

lovebeatControllers.controller('ServiceDetailCtrl', ['$scope', '$routeParams', 'Service',
  function($scope, $routeParams, Phone) {
    $scope.service = Service.get({serviceId: $routeParams.serviceId}, function(service) {
      $scope.mainImageUrl = phone.images[0];
    });

    $scope.setImage = function(imageUrl) {
      $scope.mainImageUrl = imageUrl;
    }
  }]);
