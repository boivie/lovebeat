var lovebeatServices = angular.module('lovebeatServices', ['ngResource']);

lovebeatServices.factory('Service', ['$resource',
  function($resource) {
    return $resource('api/services/:serviceId?view=:viewId', {
      serviceId: '@name'
    }, {
      get: {
        method: 'GET'
      },
      query: {
        method: 'GET',
        isArray: true
      }
    });
  }
]);

lovebeatServices.factory('View', ['$resource',
  function($resource) {
    return $resource('api/views/:viewId?', {
      viewId: '@name'
    }, {
      query: {
        method: 'GET',
        isArray: true
      }
    });
  }
]);
