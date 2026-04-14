jQuery(function ($) {

    // ── Toggle API key visibility ─────────────────────────────
    $('#trunc-toggle-key').on('click', function () {
        var $input = $('#trunc-api-key-input');
        var isPass = $input.attr('type') === 'password';
        $input.attr('type', isPass ? 'text' : 'password');
        $(this).find('.dashicons')
            .toggleClass('dashicons-visibility', !isPass)
            .toggleClass('dashicons-hidden', isPass);
    });

    // ── Test Connection ───────────────────────────────────────
    $('#trunc-test-btn').on('click', function () {
        var key = $('#trunc-api-key-input').val().trim();
        var $result = $('#trunc-test-result');

        if (!key) {
            showError($result, 'Please enter an API key first.');
            return;
        }

        $(this).prop('disabled', true).text('Testing…');

        $.post(truncAjax.ajaxurl, {
            action: 'trunc_test_key',
            nonce: truncAjax.nonce,
            api_key: key,
        }, function (response) {
            if (response.success) {
                var d = response.data;
                showSuccess(
                    $result,
                    'Connected! Plan: ' + d.plan_name + ' — ' +
                    d.current_urls + ' / ' + d.max_urls + ' URLs used this month.'
                );
            } else {
                showError($result, String(response.data));
            }
        }).always(function () {
            $('#trunc-test-btn').prop('disabled', false).text('Test Connection');
        });
    });

    // ── Shorten URL ───────────────────────────────────────────
    $('#trunc-shorten-btn').on('click', function () {
        var url = $('#trunc-original-url').val().trim();
        var customCode = $('#trunc-custom-code').val().trim();
        var $result = $('#trunc-shorten-result');

        if (!url) {
            showError($result, 'Please enter a URL.');
            return;
        }

        var $btn = $(this).prop('disabled', true).text('Shortening…');

        $.post(truncAjax.ajaxurl, {
            action: 'trunc_shorten',
            nonce: truncAjax.nonce,
            url: url,
            custom_code: customCode,
        }, function (response) {
            if (response.success) {
                var shortUrl = String(response.data.short_url);
                var $notice = buildNotice('success');

                // Safe DOM construction — no innerHTML with server data
                var $check = $('<span class="dashicons dashicons-yes-alt"></span>');
                var $text  = $('<span> ✓ Shortened: </span>');
                var $link  = $('<a target="_blank" rel="noopener noreferrer"></a>')
                    .attr('href', shortUrl)
                    .text(shortUrl);
                var $copyBtn = $('<button class="trunc-btn trunc-btn-ghost trunc-copy-inline" style="margin-left:8px"></button>')
                    .attr('data-url', shortUrl)
                    .text('Copy');

                $notice.append($check, $text, $link, $copyBtn);
                $result.empty().append($notice);

                $('#trunc-original-url').val('');
                $('#trunc-custom-code').val('');
                setTimeout(function () { location.reload(); }, 2000);
            } else {
                showError($result, String(response.data));
            }
        }).always(function () {
            $btn.prop('disabled', false).text('Shorten');
        });
    });

    // Allow Enter key in URL input
    $('#trunc-original-url').on('keydown', function (e) {
        if (e.key === 'Enter') $('#trunc-shorten-btn').trigger('click');
    });

    // ── Copy to clipboard ─────────────────────────────────────
    $(document).on('click', '.trunc-copy-btn, .trunc-copy-inline', function () {
        var $btn = $(this);
        var url = String($btn.data('url'));
        if (!url) return;
        navigator.clipboard.writeText(url).then(function () {
            var prev = $btn.text();
            $btn.text('Copied!');
            setTimeout(function () { $btn.text(prev || 'Copy'); }, 2000);
        });
    });

    // ── Helpers ───────────────────────────────────────────────
    function buildNotice(type) {
        return $('<div></div>').addClass('trunc-notice trunc-notice-' + type);
    }

    function showSuccess($el, msg) {
        var $n = buildNotice('success');
        $n.append($('<span class="dashicons dashicons-yes-alt"></span>'));
        $n.append($('<span></span>').text(' ' + msg));
        $el.empty().append($n);
    }

    function showError($el, msg) {
        var $n = buildNotice('error');
        $n.append($('<span class="dashicons dashicons-warning"></span>'));
        $n.append($('<span></span>').text(' ' + msg));
        $el.empty().append($n);
    }

});
