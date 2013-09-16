'use strict';

/* Controllers */
function DecisionTreeListCtrl($scope, DecisionTree) {
  $scope.trainingRows = DecisionTree.query()
}

//PhoneListCtrl.$inject = ['$scope', '$http'];
function DecisionTreeDetailCtrl($scope, $routeParams, DecisionTree) {
  $scope.trainingRow = DecisionTree.get({taskId: $routeParams.taskId})
}
