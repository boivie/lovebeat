'use strict';

/* Controllers */

var lovebeatControllers = angular.module('lovebeatControllers', []);

lovebeatControllers.controller('ServiceListCtrl', ['$scope', '$routeParams', 'Service', '$http', '$interval',
    function($scope, $routeParams, Service, $http, $interval) {
      $scope.services = Service.query({viewId: $routeParams.viewId});
      $scope.viewName = $routeParams.viewId;
      $scope.lbTrigger = function (service) {
	  $http({
              method : 'POST',
              url : '/api/services/' + service.name,
              data : '',
              headers : {
                  'Content-Type' : 'application/x-www-form-urlencoded'
              }
          }).success(function(data, status, headers, config) {
	      service = service.$get();
	  })
      },
      $scope.updater = $interval(function() {
	  $scope.services = Service.query({viewId: $routeParams.viewId});
      }, 60000);
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

lovebeatControllers.controller('AddViewCtrl', ['$scope', '$http',
  function($scope, $http) {
      $scope.view = {}
      $scope.createView = function() {
          $http({
              method : 'POST',
              url : '/api/views/' + $scope.view.name,
              data : 'regexp=' + $scope.view.regexp,
              headers : {
                  'Content-Type' : 'application/x-www-form-urlencoded'
              }
          }).success(function(data, status, headers, config) {
	      window.location = "#/"
	  })
      }}]);

lovebeatControllers.controller('EditServiceCtrl', ['$scope', '$routeParams', 'Service', '$http',
  function($scope, $routeParams, Service, $http) {
      $scope.service = Service.get({serviceId: $routeParams.serviceId}, function (service) {
	  if (service.warning_timeout > 0) {
	      $scope.service.warn_tmo_hr = juration.stringify(service.warning_timeout);
	  }
	  if (service.error_timeout > 0) {
	      $scope.service.err_tmo_hr = juration.stringify(service.error_timeout);
	  }
      }),
      $scope.editService = function() {
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
      },
      $scope.deleteService = function() {
          $http({
              method : 'DELETE',
              url : '/api/services/' + $scope.service.name
          }).success(function(data, status, headers, config) {
	      window.location = "#/"
	  })
      }

  }]);

lovebeatControllers.controller('EditViewCtrl', ['$scope', '$routeParams', 'View', '$http',
  function($scope, $routeParams, View, $http) {
      $scope.view = View.get({viewId: $routeParams.viewId}),
      $scope.editView = function() {
          $http({
              method : 'POST',
              url : '/api/views/' + $scope.view.name,
              data : 'regexp=' + $scope.view.regexp,
              headers : {
                  'Content-Type' : 'application/x-www-form-urlencoded'
              }
          }).success(function(data, status, headers, config) {
	      window.location = "#/"
	  })
      },
      $scope.deleteView = function() {
          $http({
              method : 'DELETE',
              url : '/api/views/' + $scope.view.name
          }).success(function(data, status, headers, config) {
	      window.location = "#/"
	  })
      }

  }]);
