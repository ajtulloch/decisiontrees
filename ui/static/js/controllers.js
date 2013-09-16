'use strict';

/* Controllers */

function DecisionTreeListCtrl($scope, $http) {
  $http.get('/api/decisiontrees/').success(function(data) {
    $scope.trainingRows = data
  });
}

//PhoneListCtrl.$inject = ['$scope', '$http'];
function DecisionTreeLDetailCtrl($scope, $routeParams, $http) {
  $scope.taskId = $routeParams.taskId
  $http.get('/api/decisiontrees/' + $scope.taskId).
    success(function(data) {
      $scope.trainingRow = data
    });
}
