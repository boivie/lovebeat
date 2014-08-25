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

lovebeatControllers.controller('AddServiceCtrl', ['$scope', '$http',
  function($scope, $http) {
      $scope.service = {}
      $scope.createService = function() {
	  var err_tmo = -1
	  var warn_tmo = -1

	  try {
	      err_tmo = juration.parse($scope.service.err_tmo_hr)
	  } catch (e) {
	  }

	  try {
	      warn_tmo = juration.parse($scope.service.warn_tmo_hr)
	  } catch (e) {
	  }

          $http({
              method : 'POST',
              url : '/api/services/' + $scope.service.name,
              data : 'err-tmo=' + err_tmo + '&warn-tmo=' + warn_tmo,
              headers : {
                  'Content-Type' : 'application/x-www-form-urlencoded'
              }
          }).success(function(data, status, headers, config) {
	      window.location = "#/"
	  })
      }}]);

lovebeatControllers.controller('ServiceDetailCtrl', ['$scope', '$routeParams', 'Service',
  function($scope, $routeParams, Phone) {
    $scope.service = Service.get({serviceId: $routeParams.serviceId}, function(service) {
      $scope.mainImageUrl = phone.images[0];
    });

    $scope.setImage = function(imageUrl) {
      $scope.mainImageUrl = imageUrl;
    }
  }]);
