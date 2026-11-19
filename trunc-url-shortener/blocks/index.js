/**
 * Trunc Short Link block. Plain JavaScript against the wp.* globals, so the
 * plugin installs from its .zip with no build step. The link itself is
 * rendered on the server (render_callback in includes/class-trunc-block.php).
 */
( function ( wp ) {
	var el = wp.element.createElement;
	var Fragment = wp.element.Fragment;
	var TextControl = wp.components.TextControl;
	var PanelBody = wp.components.PanelBody;
	var InspectorControls = wp.blockEditor.InspectorControls;
	var useBlockProps = wp.blockEditor.useBlockProps;

	wp.blocks.registerBlockType( 'trunc/short-link', {
		edit: function ( props ) {
			var shortUrl = props.attributes.shortUrl;
			var label = props.attributes.label;
			var blockProps = useBlockProps();
			return el(
				Fragment,
				null,
				el(
					InspectorControls,
					null,
					el(
						PanelBody,
						{ title: 'Link settings' },
						el( TextControl, {
							label: 'Short URL',
							value: shortUrl,
							placeholder: 'https://trunc.site/abc123',
							onChange: function ( v ) {
								props.setAttributes( { shortUrl: v } );
							},
						} ),
						el( TextControl, {
							label: 'Link text',
							value: label,
							placeholder: 'Click here',
							onChange: function ( v ) {
								props.setAttributes( { label: v } );
							},
						} )
					)
				),
				el(
					'div',
					blockProps,
					shortUrl
						? el( 'a', { href: shortUrl }, label || shortUrl )
						: el( 'span', { style: { color: '#757575' } }, 'Add a short URL in the block settings →' )
				)
			);
		},
		save: function () {
			return null;
		},
	} );
} )( window.wp );
