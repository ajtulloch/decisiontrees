'use strict';

/* Services */
angular.module('decisionTreeServices', ['ngResource']).
  factory('DecisionTree', function($resource) { 
    return $resource('/api/decisiontrees/:taskId')
  });
