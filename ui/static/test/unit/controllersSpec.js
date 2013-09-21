'use strict';

/* jasmine specs for controllers go here */
describe('DecisionTree controllers', function() {
  beforeEach(function(){
    this.addMatchers({
      toEqualData: function(expected) {
        return angular.equals(this.actual, expected);
      }
    });
  });

  beforeEach(module('decisionTreeServices'));
  
  var TEST_DATA = [{
    _id: '123', 
    forestConfig: { 
      numWeakLearners: 5, 
    }, 
    trainingResults: {
      epochResults: [{
        roc: 0.5,
        logScore: 0.3,
        normalizedEntropy: 0.5,
        calibration: 1.0,
      }]
    }
  }]

  describe('DecisionTreeListCtrl', function(){
    var scope, ctrl, $httpBackend;
    beforeEach(inject(function(_$httpBackend_, $rootScope, $controller) {
      $httpBackend = _$httpBackend_;
      $httpBackend.expectGET('/api/decisiontrees').respond(TEST_DATA)

      scope = $rootScope.$new();
      ctrl = $controller(DecisionTreeListCtrl, {$scope: scope});
    }));

    it('should return a single training row', function() {
      expect(scope.trainingRows).toEqual([]);
      $httpBackend.flush();
      expect(scope.trainingRows).toEqualData(TEST_DATA)
    });


    it('should should construct ROC graph', function() {
      // TODO(tulloch) - figure this out
    });
  })
});
