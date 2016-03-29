module.exports = function(grunt) {
	grunt.initConfig({
		clean: {
			assets:  ['www/fonts/', 'www/images/', 'www/*.css', 'www/*.js', 'tmp/']
		},

		less: {
			app: {
				options: {
					compress: true
				},
				files: {
					'tmp/app.css': 'assets/app.less',
					'tmp/theme.css': 'assets/theme.less'
				}
			}
		},

		'string-replace': {
			ungooglefont: {
				files: {
					'tmp/theme.css': 'tmp/theme.css',
				},
				options: {
					replacements: [{
						pattern: /@import url\(".+?"\);\s*/i,
						replacement: ''
					}]
				}
			}
		},

		uglify: {
			assets: {
				files: [{
					expand: true,
					cwd: 'assets/',
					src: [
						'vendor/html5upload.js',
						'vendor/jquery-timeago/jquery.timeago.js',
						'vendor/jquery-timeago/locales/jquery.timeago.en.js',
						'vendor/js-cookie/src/js.cookie.js',
						'app.js',
					],
					dest: 'tmp/',
					ext: '.min.js'
				}]
			}
		},

		concat: {
			js: {
				options: {
					separator: '\n;'
				},
				src: [
					'assets/vendor/jquery/dist/jquery.min.js',
					'tmp/vendor/html5upload.min.js',
					'tmp/vendor/jquery-timeago/jquery.min.js',
					'tmp/vendor/jquery-timeago/locales/jquery.min.js',
					'tmp/vendor/js-cookie/src/js.min.js',
					'tmp/app.min.js'
				],
				dest: 'www/app.min.js'
			},

			css: {
				options: {
					separator: '\n'
				},
				src: [
					'tmp/theme.css',
					'tmp/app.css'
				],
				dest: 'www/app.min.css'
			}
		},

		copy: {
			fontawesome: {
				files: [
					{
						expand: true,
						cwd: 'assets/vendor/font-awesome',
						src: ['fonts/*.*'],
						dest: 'www'
					}
				]
			},

			opensans: {
				files: [
					{
						expand: true,
						cwd: 'assets/vendor/lessfonts-open-sans/dist',
						src: ['fonts/**/*.*'],
						dest: 'www'
					}
				]
			},

			images: {
				files: [
					{
						expand: true,
						cwd: 'assets',
						src: ['images/**/*.*'],
						dest: 'www'
					}
				]
			}
		},

		watch: {
			css: {
				files: ['assets/*.less'],
				tasks: ['css']
			},
			app: {
				files: ['assets/*.js', 'assets/vendor/*.js'],
				tasks: ['js']
			}
		}
	});

	// load tasks
	grunt.loadNpmTasks('grunt-contrib-clean');
	grunt.loadNpmTasks('grunt-contrib-concat');
	grunt.loadNpmTasks('grunt-contrib-copy');
	grunt.loadNpmTasks('grunt-contrib-less');
	grunt.loadNpmTasks('grunt-contrib-uglify');
	grunt.loadNpmTasks('grunt-contrib-watch');
	grunt.loadNpmTasks('grunt-string-replace');

	// register custom tasks
	grunt.registerTask('css',     ['less', 'string-replace', 'concat:css', 'copy:fontawesome', 'copy:opensans']);
	grunt.registerTask('js',      ['uglify', 'concat:js']);
	grunt.registerTask('images',  ['copy:images']);
	grunt.registerTask('default', ['clean', 'css', 'js', 'images']);
};
