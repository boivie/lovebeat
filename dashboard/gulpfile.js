// Include gulp
var gulp = require('gulp');

var sass = require('gulp-sass');
var concat = require('gulp-concat');
var uglify = require('gulp-uglify');
var rename = require('gulp-rename');
var minifyCss = require('gulp-minify-css');
var sourcemaps = require('gulp-sourcemaps');
var rename = require("gulp-rename");

gulp.task('css', function() {
  return gulp.src(['css/application.scss'])
    .pipe(sourcemaps.init())
    .pipe(sass())
    .pipe(minifyCss())
    .pipe(rename("lovebeat.css"))
    .pipe(sourcemaps.write("."))
    .pipe(gulp.dest('assets/'));
});

gulp.task('scripts', function() {
  return gulp.src(['bower_components/angular/angular.js',
      'bower_components/angular-resource/angular-resource.js',
      'bower_components/angular-route/angular-route.js',
      'bower_components/angular-websocket/angular-websocket.js',
      'bower_components/bootstrap-sass/assets/javascript/bootstrap.js',
      'bower_components/jquery/dist/jquery.js',
      'bower_components/juration/juration.js',
      'js/app.js',
      'js/controllers.js',
      'js/services.js'
    ])
    .pipe(sourcemaps.init())
    .pipe(concat('lovebeat.js'))
    .pipe(uglify())
    .pipe(rename("lovebeat.js"))
    .pipe(sourcemaps.write("."))
    .pipe(gulp.dest('assets/'));
});

gulp.task('default', ['css', 'scripts']);
