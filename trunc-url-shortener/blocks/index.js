import { registerBlockType } from '@wordpress/blocks';
import { TextControl, PanelBody } from '@wordpress/components';
import { InspectorControls, useBlockProps } from '@wordpress/block-editor';
import metadata from './block.json';

registerBlockType( metadata.name, {
    edit( { attributes, setAttributes } ) {
        const { shortUrl, label } = attributes;
        const blockProps = useBlockProps();

        return (
            <>
                <InspectorControls>
                    <PanelBody title="Link Settings">
                        <TextControl
                            label="Short URL"
                            value={ shortUrl }
                            onChange={ ( v ) => setAttributes( { shortUrl: v } ) }
                            placeholder="https://trunc.site/abc123"
                        />
                        <TextControl
                            label="Link Label"
                            value={ label }
                            onChange={ ( v ) => setAttributes( { label: v } ) }
                            placeholder="Click here"
                        />
                    </PanelBody>
                </InspectorControls>
                <div { ...blockProps }>
                    { shortUrl
                        ? <a href={ shortUrl }>{ label || shortUrl }</a>
                        : <span style={ { color: '#999' } }>Set a Short URL in the sidebar →</span>
                    }
                </div>
            </>
        );
    },
    save: () => null,
} );
