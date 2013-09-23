'use strict';

/* Services */
angular.module('decisionTreeServices', ['ngResource']).
  factory('DecisionTree', function($resource) { 
    return $resource('/api/decisiontrees/:taskId')
  }).
  factory('WeakLearner', function($resource) { 
    return $resource('/api/decisiontrees/:taskId/trees/:treeId')
  }).
  factory('Page', function() {
    var title = 'Decision Tree Trainer';
    var subtitle = 'Optimization, Monitoring, and Visualization';
    return {
      title: function() { return title; },
      setTitle: function(newTitle) { title = newTitle },
      subtitle: function() { return subtitle; },
      setSubtitle: function(newSubtitle) { subtitle = newSubtitle },
    };
  });
