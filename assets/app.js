jQuery(function($) {
	$('body').on('click', 'a.postlink', function() {
		var
			self   = $(this),
			parts  = self.attr('href').split('?'),
			action = parts[0],
			params = parts[1] || '',
			form   = $('<form>').attr('action', action).attr('method', 'post'),
			token  = $('html').data('csrf');

		$('body').append(form);

		if (params !== '') {
			params = params.split('&');

			for (var i in params) {
				var tmp   = params[i].split('=');
				var key   = tmp[0];
				var value = decodeURIComponent(tmp[1]);

				$('<input>')
					.attr('type', 'hidden')
					.attr('name', key)
					.attr('value', value)
					.appendTo(form);
			}
		}

		if (token.length > 0) {
			$('<input>')
				.attr('type', 'hidden')
				.attr('name', 'csrftoken')
				.attr('value', token)
				.appendTo(form);
		}

		form.submit();
		return false;
	});

	var cookieName = 'settings';

	function readCookie() {
		var val    = Cookies.get(cookieName);
		var result = {
			v: 'public', // visibility
			s: false,    // selfdestruct
			e: 'never',  // expire,
			f: 'text',   // paste filetype
		};

		if (val) {
			var parts = val.split('|');

			if (parts.length === 4) {
				result.v = parts[0];
				result.s = parts[1] === 'true';
				result.e = parts[2];
				result.f = parts[3];
			}
		}

		return result;
	}

	function writeCookie() {
		var value = [
			userSettings.v,
			userSettings.s ? 'true' : 'false',
			userSettings.e,
			userSettings.f
		].join('|');

		Cookies.set(cookieName, value, {
			expires: 365, // days
			path: '/',
		});
	}

	function updateStatusIcons() {
		// not really a status icon, but who cares
		$('#selfdestruct').prop('disabled', userSettings.v === 'private');

		$('.statusicons .status-expire').toggleClass('disabled', userSettings.e === 'never');
		$('.statusicons .status-selfdestruct').toggleClass('disabled', !userSettings.s || userSettings.v === 'private');
		$('.statusicons .status-visibility i').addClass('hidden').filter('.visibility-' + userSettings.v).removeClass('hidden');
	}

	var userSettings = readCookie();

	$('#expire').val(userSettings.e).on('change', function() {
		userSettings.e = $(this).val();
		writeCookie();
		updateStatusIcons();
	});

	$('select[name=filetype]').val(userSettings.f).on('change', function() {
		userSettings.f = $(this).val();
		writeCookie();
	});

	$('#selfdestruct').prop('checked', userSettings.s).on('change', function() {
		userSettings.s = $(this).prop('checked');
		writeCookie();
		updateStatusIcons();
	});

	$('input[name=visibility][value="' + userSettings.v + '"]').prop('checked', true);
	$('body').on('change', 'input[name=visibility]', function() {
		userSettings.v = $(this).val();
		writeCookie();
		updateStatusIcons();
	});

	updateStatusIcons();

	$('.opttoggle').on('click', function() {
		var group = $('.ext-options');

		group.slideToggle('fast');
		$(this).toggleClass('active');
	});

	$('body').on('click', '.remover', function() {
		var row = $(this).closest('li');

		if (row.is('.finished')) {
			row.slideUp('fast', function() {
				row.remove();
			});
		}
	});

	// drag&drop file upload logic
	// based on http://stackoverflow.com/a/33917000/564807

	var dropZone = $('.drop-overlay')[0];
	var dropHelp = $('.drop-help')[0];

	function showDropZone() {
		dropZone.style.visibility = 'visible';
		dropHelp.style.visibility = 'visible';
	}

	function hideDropZone() {
		dropZone.style.visibility = 'hidden';
		dropHelp.style.visibility = 'hidden';
	}

	function allowDrag(e) {
		if (true) {  // Test that the item being dragged is a valid one
			e.dataTransfer.dropEffect = 'copy';
			e.preventDefault();
		}
	}

	function handleDrop(e) {
		e.preventDefault();
		hideDropZone();
	}

	window.addEventListener('dragenter', showDropZone);

	dropZone.addEventListener('dragenter', allowDrag);
	dropZone.addEventListener('dragover', allowDrag);
	dropZone.addEventListener('dragleave', hideDropZone);
	dropZone.addEventListener('drop', handleDrop);

	// uploading functionality

	function setProgress(node, percent) {
		var complete = percent >= 100;

		$(node).find('.state').text(percent + '%');
		$(node).find('.progress-bar')
			.css('width', percent + '%')
			.attr('aria-valuenow', percent)
			.toggleClass('active', !complete)
			.toggleClass('progress-bar-info', !complete)
			.toggleClass('progress-bar-success', complete);
	}

	html5upload.initialize({
		maxSimultaneousUploads: 3,
		uploadUrl: '/upload',
		dropContainer: dropZone,
		inputField: $('.btn-file input')[0],
		key: 'file',

		data: function() {
			return {
				expire:       $('#expire').val() || 'never',
				selfdestruct: $('#selfdestruct').prop('checked') ? 1 : 0,
				visibility:   $('input[name=visibility]:checked').val() || 'public',
				csrftoken:    $('html').data('csrf')
			};
		},

		onFileAdded: function (file) {
			var node = $($('#upload-item').html());

			node.find('.name span').text(file.fileName);
			$('#uploads').append(node);

			file.on({
				// Called after received response from the server
				onCompleted: function(response) {
					try {
						response = JSON.parse(response);
					}
					catch (e) {
						return;
					}

					setProgress(node, 100);
					node.removeClass('uploading').addClass('finished');

					if (response.status === 'ok') {
						node.find('.name span').after(
							$('<a>').text(file.fileName).attr('href', response.uri).attr('target', '_blank')
						).remove();
					}
					else {
						node.find('.progress-bar').removeClass('progress-bar-success').removeClass('progress-bar-info').addClass('progress-bar-danger');
						node.find('.name span').addClass('text-danger');
					}
				},

				onProgress: function (progress, fileSize, uploadedBytes) {
					setProgress(node, parseInt(progress, 10));
				}
			});
		}
	});

	jQuery('time.rel').timeago();
});
