'use strict';

/* Services */
angular.module('decisionTreeServices', ['ngResource']).
  factory('DecisionTree', function($resource) { 
    return $resource('/api/decisiontrees/:taskId')
  }).
  factory('WeakLearner', function($resource) { 
    return $resource('/api/decisiontrees/:taskId/trees/:treeId')
  });

