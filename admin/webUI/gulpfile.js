var gulp = require('gulp'),
  autoprefixer = require('gulp-autoprefixer'),
  cssnano = require('gulp-cssnano'),
  jshint = require('gulp-jshint'),
  uglify = require('gulp-uglify'),
  rename = require('gulp-rename'),
  concat = require('gulp-concat'),
  order = require('gulp-order'),
  del = require("del"),
  browserSync = require('browser-sync').create();

var lib_root_path = "assets/lib/";
var style_files = ['assets/css/**/*.css', '!assets/css/main.css', ];
var scriptFiles = [
  'assets/lib/js/angular.min.js',
  'assets/lib/js/angular-sanitize.min.js',
  'assets/lib/js/angular-route.min.js?v=20161018-1',
  'assets/lib/js/angular-ui-router.min.js',
  'assets/lib/js/angular-css.min.js',
  'assets/lib/js/angular-select.min.js',
  'assets/lib/js/jquery-2.1.1.min.js',
  'assets/lib/js/material.min.js',
  'assets/lib/js/ocLazyLoad.min.js',
  'assets/lib/js/duration.js',
  'assets/lib/js/ui-codemirror.min.js',
];

// Clean old compiled files
gulp.task('cleanOldFiles', function() {
  return del(['assets/compiled/css/**/*.css', "assets/compiled/js/**/*/*.js"]);
});

// Minify CSS 
gulp.task('styles', function() {
  return gulp.src(style_files)
    .pipe(autoprefixer('last 2 version'))
    .pipe(cssnano())
    .pipe(concat('gru.min.css'))
    .pipe(gulp.dest('assets/compiled/css'))
});

// Minify Javascript
// gulp.task('lib-js', function() {
//   return streamqueue({ objectMode: true },
//       gulp.src('assets/lib/js/angular.min.js'),
//       gulp.src('assets/lib/js/angular-sanitize.min.js'),
//       gulp.src('assets/lib/js/angular-route.min.js?v=20161018-1'),
//       gulp.src('assets/lib/js/angular-ui-router.min.js'),
//       gulp.src('assets/lib/js/angular-css.min.js'),
//       gulp.src('assets/lib/js/angular-select.min.js'),
//       gulp.src('assets/lib/js/jquery-2.1.1.min.js'),
//       gulp.src('assets/lib/js/material.min.js'),
//       gulp.src('assets/lib/js/ocLazyLoad.min.js'),
//       gulp.src('assets/lib/js/duration.js'),
//       gulp.src('assets/lib/js/ui-codemirror.min.js')
//     )
//     .pipe(concat('vendor.js'))
//     .pipe(gulp.dest('assets/compiled/js'))
//     .pipe(rename({ suffix: '.min' }))
//     .pipe(uglify())
//     .pipe(gulp.dest('assets/compiled/js'))
// });

// gulp.task('lib-js', function() {
//   console.log("ddjf");
//   return gulp.src(scriptFiles)
//     .pipe(jshint())
//     .pipe(jshint.reporter('default'))
//     .pipe(order(scriptFiles, { base: './' }))
//     .pipe(concat('vendor.js'))
//     .pipe(gulp.dest('assets/compiled/js'))
//     .pipe(rename({ suffix: '.min' }))
//     .pipe(uglify())
//     .pipe(gulp.dest('assets/compiled/js'))
// });

gulp.task('browser-sync', function() {
  browserSync.init(null, {
    proxy: 'http://localhost:2020',
    files: ['app/**/*.html', 'assets/compiled/css/**/*.css', "assets/compiled/js/**/*/*.js", ],
    browser: 'google chrome',
    port: 5001,
    open: false
  });
});

// Watch on file changes
gulp.task('watch', function() {
  // Watch .css files
  gulp.watch(style_files, ['styles']);
  gulp.watch(scriptFiles, ['lib-js']);
});

// Default task
gulp.task('default', ['cleanOldFiles', 'watch', 'browser-sync'], function() {
  gulp.start('styles');
  gulp.start('lib-js');
  console.log("Started listing for changes..")
});
