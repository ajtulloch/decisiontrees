'use strict';

/* http://docs.angularjs.org/guide/dev_guide.e2e-testing */
describe('DecisionTrees App', function() {
  it('should redirect index.html to index.html#/decisiontrees', function() {
    browser().navigateTo('/');
    expect(browser().location().url()).toBe('/decisiontrees');
  });


  describe('Phone list view', function() {
    beforeEach(function() {
      browser().navigateTo('/');
    });


    it('should filter the phone list as user types into the search box', function() {
    });
  });

});
