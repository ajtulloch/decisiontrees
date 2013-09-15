'use strict';

/* Controllers */

function DecisionTreeListCtrl($scope, $http) {
  $http.get('http://localhost:5000/api/decisiontrees/').success(function(data) {
    console.log(data)
    $scope.trees = data
  });
}

//PhoneListCtrl.$inject = ['$scope', '$http'];
function DecisionTreeLDetailCtrl($scope, $routeParams, $http) {
  $scope.taskId = $routeParams.taskId
  $http.get('http://localhost:5000/api/decisiontrees/' + $scope.taskId).
    success(function(data) {
      $scope.trainingRow = data
    });
}
