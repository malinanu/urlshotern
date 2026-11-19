<?php
/**
 * Plugin Name: Trunc URL Shortener
 * Plugin URI: https://trunc.site/integrations
 * Description: Shorten URLs and track clicks with Trunc analytics directly from your WordPress dashboard.
 * Version: 1.1.0
 * Requires at least: 6.0
 * Requires PHP: 8.0
 * Author: Trunc
 * Author URI: https://trunc.site
 * License: GPL-2.0-or-later
 * License URI: https://www.gnu.org/licenses/gpl-2.0.html
 * Text Domain: trunc-url-shortener
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

define( 'TRUNC_VERSION', '1.1.0' );
define( 'TRUNC_PLUGIN_DIR', plugin_dir_path( __FILE__ ) );
define( 'TRUNC_PLUGIN_URL', plugin_dir_url( __FILE__ ) );
define( 'TRUNC_DEFAULT_SERVER', 'https://trunc.site' );

/**
 * The Trunc server this site talks to: trunc.site, or a self-hosted server
 * set under Trunc Links → Settings.
 */
function trunc_server(): string {
	$server = get_option( 'trunc_server_url', '' );
	return $server ? untrailingslashit( $server ) : TRUNC_DEFAULT_SERVER;
}

function trunc_api_base(): string {
	return trunc_server() . '/api/v1';
}

require_once TRUNC_PLUGIN_DIR . 'includes/class-trunc-api.php';
require_once TRUNC_PLUGIN_DIR . 'includes/class-trunc-connect.php';
require_once TRUNC_PLUGIN_DIR . 'includes/class-trunc-admin.php';
require_once TRUNC_PLUGIN_DIR . 'includes/class-trunc-block.php';

register_activation_hook( __FILE__, 'trunc_activate' );

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

/**
 * Creates or updates the table when the plugin version changes, so updates
 * that skip the activation hook (and a missing table) are handled.
 */
function trunc_maybe_upgrade() {
	if ( get_option( 'trunc_db_version' ) !== TRUNC_VERSION ) {
		trunc_activate();
		update_option( 'trunc_db_version', TRUNC_VERSION, false );
	}
}

add_action(
	'plugins_loaded',
	function () {
		trunc_maybe_upgrade();
		( new Trunc_Connect() )->init();
		( new Trunc_Admin() )->init();
		( new Trunc_Block() )->init();
	}
);
