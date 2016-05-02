'use strict';

var lovebeatControllers = angular.module('lovebeatControllers', []);

lovebeatControllers.controller('ServiceListCtrl', ['$scope', '$routeParams',
  'Service', '$http', '$interval', 'LovebeatStream',
  function($scope, $routeParams, Service, $http, $interval, LovebeatStream) {
    $scope.stream = LovebeatStream

    $scope.services = Service.query({
      viewId: $routeParams.viewId
    });
    $scope.viewName = $routeParams.viewId;
    $scope.$on("LovebeatStream::service_changed", function(event, args) {
      $scope.$apply(function() {
        var len = $scope.services.length
        for (var i = 0; i < len; i++) {
          var service = $scope.services[i]
          if (service.name == args.name) {
            service.state = args.state
            break;
          }
        }
      });
    });
    $scope.lbTrigger = function(service) {
        $http({
          method: 'POST',
          url: 'api/services/' + service.name,
          data: '',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        }).success(function(data, status, headers, config) {
          service = service.$get();
        })
      },
      $scope.updater = $interval(function() {
        $scope.services = Service.query({
          viewId: $routeParams.viewId
        });
      }, 60000);
  }
]);

lovebeatControllers.controller('ViewListCtrl', ['$scope', 'View', 'LovebeatStream',
  function($scope, View, LovebeatStream) {
    $scope.stream = LovebeatStream
    $scope.views = View.query();
    $scope.$on("LovebeatStream::view_changed", function(event, args) {
      $scope.$apply(function() {
        var len = $scope.views.length
        for (var i = 0; i < len; i++) {
          var view = $scope.views[i]
          if (view.name == args.name) {
            view.state = args.state
            break;
          }
        }
      });
    });
  }
]);

lovebeatControllers.controller('AddServiceCtrl', ['$scope', '$http',
  function($scope, $http) {
    $scope.service = {}
    $scope.createService = function() {
      var timeout = -1

      try {
        timeout = juration.parse($scope.service.timeout_hr)
      } catch (e) {}

      $http({
        method: 'POST',
        url: 'api/services/' + $scope.service.name,
        data: 'timeout=' + timeout,
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      }).success(function(data, status, headers, config) {
        window.location = "#/"
      })
    }
  }
]);

lovebeatControllers.controller('EditServiceCtrl', ['$scope', '$routeParams', 'Service', '$http',
  function($scope, $routeParams, Service, $http) {
    $scope.service = Service.get({
        serviceId: $routeParams.serviceId
      }, function(service) {
        if (service.timeout > 0) {
          $scope.service.timeout_hr = juration.stringify(service.timeout / 1000);
        } else if (service.timeout == 0) {
          $scope.service.timeout_hr = "0s";
        }
      }),
      $scope.editService = function() {
        var timeout = -1

        try {
          timeout = juration.parse($scope.service.timeout_hr)
        } catch (e) {}

        $http({
          method: 'POST',
          url: 'api/services/' + $scope.service.name,
          data: 'timeout=' + timeout,
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        }).success(function(data, status, headers, config) {
          window.location = "#/"
        })
      },
      $scope.deleteService = function() {
        $http({
          method: 'DELETE',
          url: 'api/services/' + $scope.service.name
        }).success(function(data, status, headers, config) {
          window.location = "#/"
        })
      }

  }
]);
