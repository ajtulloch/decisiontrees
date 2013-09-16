module.exports = function (config) {
  config.set({
    basePath: '../',

    files: [
      'lib/angular/angular.js',
      'lib/angular/angular-*.js',
      'lib/angular/angular-mocks.js',
      'js/**/*.js',
      'test/unit/**/*.js'
    ],

    frameworks: ['jasmine'],

    autoWatch: true,

    browsers: ['Chrome'],

    junitReporter: {
      outputFile: 'test_out/unit.xml',
      suite: 'unit'
    }
  });
};
