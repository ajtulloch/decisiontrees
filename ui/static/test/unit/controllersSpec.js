'use strict';

/* jasmine specs for controllers go here */
describe('DecisionTree controllers', function() {

  describe('DecisionTreeListCtrl', function(){
    var scope, ctrl, $httpBackend;
    beforeEach(inject(function(_$httpBackend_, $rootScope, $controller) {
      $httpBackend = _$httpBackend_;
      $httpBackend.expectGET('/api/decisiontrees/').
          respond([{_id: '123', forestConfig: { numWeakLearners: 5, }, }])

      scope = $rootScope.$new();
      ctrl = $controller(DecisionTreeListCtrl, {$scope: scope});
    }));


  it('should return a single training row', function() {
    expect(scope.trainingRows).toBeUndefined();
    $httpBackend.flush();

    expect(scope.trainingRows).toEqual([{_id: '123', forestConfig: { numWeakLearners: 5, }}])
  });


  //   it('should set the default value of orderProp model', function() {
  //     expect(scope.orderProp).toBe('age');
  //   });
  // });


  // describe('PhoneDetailCtrl', function(){
  //   var scope, $httpBackend, ctrl;

  //   beforeEach(inject(function(_$httpBackend_, $rootScope, $routeParams, $controller) {
  //     $httpBackend = _$httpBackend_;
  //     $httpBackend.expectGET('phones/xyz.json').respond({name:'phone xyz'});

  //     $routeParams.phoneId = 'xyz';
  //     scope = $rootScope.$new();
  //     ctrl = $controller(PhoneDetailCtrl, {$scope: scope});
  //   }));


  //   it('should fetch phone detail', function() {
  //     expect(scope.phone).toBeUndefined();
  //     $httpBackend.flush();

  //     expect(scope.phone).toEqual({name:'phone xyz'});
  //   });
  // });
  })
});
