<?php
if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class Trunc_Block {
    public function init(): void {
        add_action( 'init', [ $this, 'register_block' ] );
        add_shortcode( 'trunc', [ $this, 'render_shortcode' ] );
    }

    public function register_block(): void {
        register_block_type(
            TRUNC_PLUGIN_DIR . 'blocks',
            [ 'render_callback' => [ $this, 'render_block' ] ]
        );
    }

    /**
     * Server-side render for the Gutenberg block.
     * Attributes: shortUrl, label
     */
    public function render_block( array $attrs ): string {
        $url   = $attrs['shortUrl'] ?? '';
        $label = $attrs['label']    ?? $url;
        if ( empty( $url ) ) {
            return '';
        }
        return '<a href="' . esc_url( $url ) . '" rel="nofollow">' . esc_html( $label ) . '</a>';
    }

    /**
     * Shortcode: [trunc url="https://trunc.site/abc" label="Click here"]
     */
    public function render_shortcode( array $atts ): string {
        $atts = shortcode_atts(
            [ 'url' => '', 'label' => '' ],
            $atts
        );
        if ( empty( $atts['url'] ) ) {
            return '';
        }
        $label = ! empty( $atts['label'] ) ? $atts['label'] : $atts['url'];
        return '<a href="' . esc_url( $atts['url'] ) . '" rel="nofollow">' . esc_html( $label ) . '</a>';
    }
}
