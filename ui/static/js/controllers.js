'use strict';

/* Controllers */
function DecisionTreeListCtrl($scope, DecisionTree) {
  $scope.trainingRows = DecisionTree.query()
}

//PhoneListCtrl.$inject = ['$scope', '$http'];
function DecisionTreeDetailCtrl($scope, $routeParams, $log, DecisionTree) {
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
    }
  )
}
