var gulp = require('gulp'),
  autoprefixer = require('gulp-autoprefixer'),
  cssnano = require('gulp-cssnano'),
  jshint = require('gulp-jshint'),
  uglify = require('gulp-uglify'),
  rename = require('gulp-rename'),
  concat = require('gulp-concat'),
  order = require('gulp-order'),
  notify = require('gulp-notify'),
  del = require("del"),
  browserSync = require("browser-sync").create();

var lib_root_path = "assets/lib/";
var style_files = ['assets/css/**/*.css', '!assets/css/main.css', ]

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
    .pipe(browserSync.stream())
    // .pipe(notify({ message: 'Minify css task completed.' }));
});

// Minify Javascript 
// gulp.task('lib-js', function() {
//   var scriptFiles = [
//     'assets/lib/js/angular.min.js',
//     'assets/lib/js/angular-sanitize.min.js',
//     'assets/lib/js/angular-route.min.js?v=20161018-1',
//     'assets/lib/js/angular-ui-router.min.js',
//     'assets/lib/js/angular-css.min.js',
//     'assets/lib/js/angular-select.min.js',
//     'assets/lib/js/jquery-2.1.1.min.js',
//     'assets/lib/js/material.min.js',
//     'assets/lib/js/ocLazyLoad.min.js',
//     'assets/lib/js/duration.js',
//     'assets/lib/js/ui-codemirror.min.js',
//   ]

//   return gulp.src(scriptFiles)
//     // .pipe(jshint())
//     // .pipe(jshint.reporter('default'))
//     // .pipe(order(scriptFiles))
//     .pipe(concat('vendor.js'))
//     .pipe(gulp.dest('assets/compiled/js'))
//     .pipe(rename({ suffix: '.min' }))
//     .pipe(uglify())
//     .pipe(gulp.dest('assets/compiled/js'))
//     .pipe(notify({ message: 'Minify scripts task complete' }));
// });

gulp.task('browser-sync', function() {
  browserSync.init(null, {
    proxy: 'localhost:2020',
    files: style_files,
    browser: 'google chrome',
    port: 5000,
    open: false
  });
});

// Watch on changes
gulp.task('watch', function() {
  // Watch .css files
  gulp.watch(style_files, ['styles']);
});

// Default task
gulp.task('default', ['cleanOldFiles', 'watch', 'browser-sync'], function() {
  gulp.start('styles');
  console.log("Started listing for changes..")
});
