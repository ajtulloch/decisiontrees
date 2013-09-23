'use strict';

/* App Module */
angular.module('decisiontrees', ['decisiontreeDirectives', 'decisionTreeServices', 'nvd3ChartDirectives']).
  config(['$routeProvider', function($routeProvider) {
  $routeProvider.
    when('/decisiontrees', {templateUrl: 'static/partials/decisiontree-list.html',   controller: DecisionTreeListCtrl}).
    when('/decisiontrees/:taskId', {templateUrl: 'static/partials/decisiontree-detail.html', controller: DecisionTreeDetailCtrl}).
    when('/decisiontrees/:taskId/trees/:treeId', {templateUrl: 'static/partials/decisiontree-weaklearner.html', controller: DecisionTreeWeakLearnerCtrl}).
    otherwise({redirectTo: '/decisiontrees'});
}]);
