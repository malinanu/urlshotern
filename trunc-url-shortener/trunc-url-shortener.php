<?php
/**
 * Plugin Name: Trunc URL Shortener
 * Plugin URI: https://trunc.site
 * Description: Shorten URLs and track clicks with Trunc analytics directly from your WordPress dashboard.
 * Version: 1.0.0
 * Requires at least: 6.0
 * Requires PHP: 8.0
 * Author: Trunc
 * Author URI: https://trunc.site
 * License: GPL-2.0+
 * Text Domain: trunc-url-shortener
 */

if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

define( 'TRUNC_VERSION', '1.0.0' );
define( 'TRUNC_PLUGIN_DIR', plugin_dir_path( __FILE__ ) );
define( 'TRUNC_PLUGIN_URL', plugin_dir_url( __FILE__ ) );
define( 'TRUNC_API_BASE', 'https://trunc.site/api/v1' );

require_once TRUNC_PLUGIN_DIR . 'includes/class-trunc-api.php';
require_once TRUNC_PLUGIN_DIR . 'includes/class-trunc-admin.php';
require_once TRUNC_PLUGIN_DIR . 'includes/class-trunc-block.php';

register_activation_hook( __FILE__, 'trunc_activate' );
register_deactivation_hook( __FILE__, 'trunc_deactivate' );

function trunc_activate() {
    global $wpdb;
    $table   = $wpdb->prefix . 'trunc_links';
    $charset = $wpdb->get_charset_collate();
    $sql     = "CREATE TABLE IF NOT EXISTS $table (
        id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
        original_url TEXT NOT NULL,
        short_code VARCHAR(50) NOT NULL,
        short_url VARCHAR(255) NOT NULL,
        clicks BIGINT(20) UNSIGNED DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (id),
        UNIQUE KEY short_code (short_code)
    ) $charset;";
    require_once ABSPATH . 'wp-admin/includes/upgrade.php';
    dbDelta( $sql );
}

function trunc_deactivate() {
    // No cleanup on deactivate — preserve user data
}

add_action( 'plugins_loaded', function () {
    $admin = new Trunc_Admin();
    $admin->init();
    $block = new Trunc_Block();
    $block->init();
} );
