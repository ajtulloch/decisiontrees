'use strict';

/* Controllers */
function DecisionTreeListCtrl($scope, DecisionTree) {
  $scope.trainingRows = DecisionTree.query()
}

//PhoneListCtrl.$inject = ['$scope', '$http'];
function DecisionTreeDetailCtrl($scope, $routeParams, $log, DecisionTree) {
  $scope.TREE_ROW_SIZE = 3
  $scope.trainingRow = DecisionTree.get(
    {taskId: $routeParams.taskId}, 
    function(row, headers) {
      var lineChart = function(label, accessor) {
        return {
          "key": label,
          "values": row.trainingResults.epochResults.map(function(e, i) {
            return [
              // decision trees are enumerated from [1, numWeakLearners]
              i + 1, accessor(e).toFixed(3)]
          })
        }
      }

      $scope.epochResults = [
        lineChart("ROC", function(e) { return e.roc }),
        lineChart("Calibration", function(e) { return e.calibration }),
        lineChart("Normalized Entropy", function(e) { 
          return e.normalizedEntropy 
        }),
      ]

      
      var treeRows = []
      var currentRow = []
      for (var i = 0; i < row.forest.trees.length; i++) {
        row.forest.trees[i].index = i;
        currentRow.push(row.forest.trees[i])
        if (currentRow.length >= $scope.TREE_ROW_SIZE) {
            treeRows.push(currentRow)
            currentRow = []
        }
      }

      if (currentRow.length) {
        treeRows.push(currentRow)
      }

      $scope.treeRows = treeRows
      $log.info($scope.treeRows)
    }
  )
}

function DecisionTreeWeakLearnerCtrl($scope, $routeParams, $log, WeakLearner) {
  $scope.tree = WeakLearner.get({taskId: $routeParams.taskId, treeId: $routeParams.treeId})
}
