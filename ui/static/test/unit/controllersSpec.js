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

  describe('DecisionTreeListCtrl', function(){
    var scope, ctrl, $httpBackend;
    beforeEach(inject(function(_$httpBackend_, $rootScope, $controller) {
      $httpBackend = _$httpBackend_;
      $httpBackend.expectGET('/api/decisiontrees').
          respond([{_id: '123', forestConfig: { numWeakLearners: 5, }, }])

      scope = $rootScope.$new();
      ctrl = $controller(DecisionTreeListCtrl, {$scope: scope});
    }));

    it('should return a single training row', function() {
      expect(scope.trainingRows).toEqual([]);
      $httpBackend.flush();
      expect(scope.trainingRows).toEqualData([{_id: '123', forestConfig: { numWeakLearners: 5, }}])
    });
  })
});
