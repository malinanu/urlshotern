<?php
/**
 * "Connect with Trunc": OAuth 2.0 authorization code + PKCE.
 *
 * The admin is sent to <server>/connect, approves the site, and comes back to
 * the settings page with a one-time code. The code is exchanged here, server
 * side, for an API key of the site's own (client id: trunc-wordpress).
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

class Trunc_Connect {
	const CLIENT_ID = 'trunc-wordpress';
	const SETTINGS  = 'trunc-settings';

	public function init(): void {
		add_action( 'admin_post_trunc_connect', array( $this, 'start' ) );
		add_action( 'admin_post_trunc_disconnect', array( $this, 'disconnect' ) );
		add_action( 'admin_init', array( $this, 'maybe_finish' ) );
	}

	/** Where Trunc sends the admin back to. Must be identical at both steps. */
	public static function redirect_uri(): string {
		return admin_url( 'admin.php?page=' . self::SETTINGS );
	}

	/** Connect with Trunc needs https (or localhost while developing). */
	public static function site_can_connect(): bool {
		$parts = wp_parse_url( self::redirect_uri() );
		$host  = $parts['host'] ?? '';
		return ( $parts['scheme'] ?? '' ) === 'https' || in_array( $host, array( 'localhost', '127.0.0.1' ), true );
	}

	private static function base64url( string $bytes ): string {
		return rtrim( strtr( base64_encode( $bytes ), '+/', '-_' ), '=' ); // phpcs:ignore WordPress.PHP.DiscouragedPHPFunctions.obfuscation_base64_encode -- PKCE encoding (RFC 7636), not obfuscation.
	}

	private static function settings_url( array $args = array() ): string {
		return add_query_arg( $args, self::redirect_uri() );
	}

	/** Sends the admin to Trunc to approve this site. */
	public function start(): void {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( esc_html__( 'You are not allowed to do that.', 'trunc-url-shortener' ), 403 );
		}
		check_admin_referer( 'trunc_connect' );

		if ( ! self::site_can_connect() ) {
			wp_safe_redirect( self::settings_url( array( 'trunc_notice' => 'needs_https' ) ) );
			exit;
		}

		$verifier = self::base64url( random_bytes( 48 ) );
		$state    = wp_generate_password( 32, false, false );
		set_transient(
			'trunc_connect_' . $state,
			array(
				'verifier' => $verifier,
				'user'     => get_current_user_id(),
			),
			10 * MINUTE_IN_SECONDS
		);

		$url = add_query_arg(
			array(
				'response_type'         => 'code',
				'client_id'             => self::CLIENT_ID,
				'redirect_uri'          => rawurlencode( self::redirect_uri() ),
				'state'                 => $state,
				'scope'                 => 'links',
				'code_challenge'        => self::base64url( hash( 'sha256', $verifier, true ) ),
				'code_challenge_method' => 'S256',
			),
			trunc_server() . '/connect'
		);
		// The Trunc server is chosen by the admin in settings; wp_safe_redirect would refuse it.
		wp_redirect( $url ); // phpcs:ignore WordPress.Security.SafeRedirect.wp_redirect_wp_redirect
		exit;
	}

	/** Handles the return from Trunc on the settings page. */
	public function maybe_finish(): void {
		// phpcs:disable WordPress.Security.NonceVerification.Recommended -- the OAuth state (stored per admin) is the CSRF check here.
		if ( ! isset( $_GET['page'], $_GET['state'] ) || self::SETTINGS !== $_GET['page'] ) {
			return;
		}
		if ( ! isset( $_GET['code'] ) && ! isset( $_GET['error'] ) ) {
			return;
		}
		if ( ! current_user_can( 'manage_options' ) ) {
			return;
		}
		$state = sanitize_text_field( wp_unslash( $_GET['state'] ) );
		$saved = get_transient( 'trunc_connect_' . $state );
		delete_transient( 'trunc_connect_' . $state );

		if ( ! is_array( $saved ) || (int) $saved['user'] !== get_current_user_id() ) {
			wp_safe_redirect( self::settings_url( array( 'trunc_notice' => 'expired' ) ) );
			exit;
		}
		if ( isset( $_GET['error'] ) ) {
			$err = sanitize_key( wp_unslash( $_GET['error'] ) );
			wp_safe_redirect( self::settings_url( array( 'trunc_notice' => 'access_denied' === $err ? 'cancelled' : 'failed' ) ) );
			exit;
		}
		$code = sanitize_text_field( wp_unslash( $_GET['code'] ) );
		// phpcs:enable

		$response = wp_remote_post(
			trunc_api_base() . '/oauth/token',
			array(
				'timeout' => 15,
				'body'    => array(
					'grant_type'    => 'authorization_code',
					'code'          => $code,
					'redirect_uri'  => self::redirect_uri(),
					'client_id'     => self::CLIENT_ID,
					'code_verifier' => $saved['verifier'],
				),
			)
		);
		$data = is_wp_error( $response ) ? array() : json_decode( wp_remote_retrieve_body( $response ), true );
		if ( empty( $data['access_token'] ) ) {
			wp_safe_redirect( self::settings_url( array( 'trunc_notice' => 'failed' ) ) );
			exit;
		}

		update_option( 'trunc_api_key', sanitize_text_field( $data['access_token'] ), false );
		update_option( 'trunc_connected_via', 'signin', false );
		wp_safe_redirect( self::settings_url( array( 'trunc_notice' => 'connected' ) ) );
		exit;
	}

	/** Forgets the key; a key from Connect with Trunc is also revoked on Trunc. */
	public function disconnect(): void {
		if ( ! current_user_can( 'manage_options' ) ) {
			wp_die( esc_html__( 'You are not allowed to do that.', 'trunc-url-shortener' ), 403 );
		}
		check_admin_referer( 'trunc_disconnect' );

		$key = get_option( 'trunc_api_key', '' );
		if ( $key && 'signin' === get_option( 'trunc_connected_via', '' ) ) {
			wp_remote_post(
				trunc_api_base() . '/oauth/revoke',
				array(
					'timeout' => 10,
					'body'    => array(
						'token'     => $key,
						'client_id' => self::CLIENT_ID,
					),
				)
			);
		}
		delete_option( 'trunc_api_key' );
		delete_option( 'trunc_connected_via' );
		wp_safe_redirect( self::settings_url( array( 'trunc_notice' => 'disconnected' ) ) );
		exit;
	}
}
