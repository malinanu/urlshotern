<?php
/**
 * Removes everything the plugin stored when it is deleted from WordPress.
 * Short links on Trunc itself are kept; they belong to the Trunc account.
 */

if ( ! defined( 'WP_UNINSTALL_PLUGIN' ) ) {
	exit;
}

delete_option( 'trunc_api_key' );
delete_option( 'trunc_connected_via' );
delete_option( 'trunc_server_url' );
delete_option( 'trunc_db_version' );

global $wpdb;
$wpdb->query( "DROP TABLE IF EXISTS {$wpdb->prefix}trunc_links" ); // phpcs:ignore WordPress.DB.DirectDatabaseQuery
