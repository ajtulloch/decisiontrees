'use strict';

/* App Module */
angular.module('decisiontrees', ['decisiontreeDirectives']).
  config(['$routeProvider', function($routeProvider) {
  $routeProvider.
      when('/decisiontrees', {templateUrl: 'static/partials/decisiontree-list.html',   controller: DecisionTreeListCtrl}).
      when('/decisiontrees/:taskId', {templateUrl: 'static/partials/decisiontree-detail.html', controller: DecisionTreeLDetailCtrl}).
      otherwise({redirectTo: '/decisiontrees'});
}]);
