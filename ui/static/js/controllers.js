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

      // Compute the number of trees 
      $scope.trainingRow.forest.trees.forEach(function(n) {
        var numNodes = function recur(t) {
          if (t.splitValue) {
            return recur(t.left) + recur(t.right)
          } else {
            return 1
          }
        }
        n.numNodes = numNodes(n)
      })
    }
  )
}

function DecisionTreeWeakLearnerCtrl($scope, $routeParams, $log, WeakLearner) {
  $scope.tree = WeakLearner.get({taskId: $routeParams.taskId, treeId: $routeParams.treeId})
}
