<?php
if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class Trunc_API {
    private string $api_key;
    private string $base_url;

    public function __construct( string $api_key ) {
        $this->api_key  = $api_key;
        $this->base_url = TRUNC_API_BASE;
    }

    /**
     * Shorten a URL.
     *
     * @return array{short_code: string, short_url: string}|WP_Error
     */
    public function shorten( string $url, string $custom_code = '' ): array|WP_Error {
        $body = [ 'url' => $url ];
        if ( $custom_code !== '' ) {
            $body['custom_code'] = $custom_code;
        }

        $response = wp_remote_post(
            $this->base_url . '/shorten',
            [
                'headers' => [
                    'Authorization' => 'Bearer ' . $this->api_key,
                    'Content-Type'  => 'application/json',
                ],
                'body'    => wp_json_encode( $body ),
                'timeout' => 15,
            ]
        );

        if ( is_wp_error( $response ) ) {
            return $response;
        }

        $code = wp_remote_retrieve_response_code( $response );
        $data = json_decode( wp_remote_retrieve_body( $response ), true );

        if ( $code !== 200 && $code !== 201 ) {
            $msg = $data['error'] ?? 'Unknown API error';
            return new WP_Error( 'trunc_api_error', $msg, [ 'status' => $code ] );
        }

        return [
            'short_code' => $data['short_code'] ?? '',
            'short_url'  => $data['short_url']  ?? '',
        ];
    }

    /**
     * Get plan, usage and limits for the current API key owner.
     * Results are cached for 5 minutes via WP transients.
     *
     * @return array|WP_Error
     */
    public function get_usage(): array|WP_Error {
        $cache_key = 'trunc_usage_' . substr( md5( $this->api_key ), 0, 8 );
        $cached    = get_transient( $cache_key );
        if ( $cached !== false ) {
            return $cached;
        }

        $response = wp_remote_get(
            $this->base_url . '/billing/usage',
            [
                'headers' => [ 'Authorization' => 'Bearer ' . $this->api_key ],
                'timeout' => 15,
            ]
        );

        if ( is_wp_error( $response ) ) {
            return $response;
        }

        $code = wp_remote_retrieve_response_code( $response );
        $data = json_decode( wp_remote_retrieve_body( $response ), true );

        if ( $code !== 200 ) {
            $msg = $data['error'] ?? 'Could not fetch account info';
            return new WP_Error( 'trunc_api_error', $msg, [ 'status' => $code ] );
        }

        $result = [
            'plan_type'         => $data['subscription']['plan_type'] ?? 'free',
            'plan_name'         => $data['plan']['name']              ?? 'Free',
            'current_urls'      => $data['usage']['current_urls']     ?? 0,
            'max_urls'          => $data['limits']['max_urls']        ?? 10,
            'monthly_clicks'    => $data['usage']['monthly_clicks']   ?? 0,
            'max_clicks'        => $data['limits']['max_clicks_month'] ?? 5000,
            'url_pct'           => $data['usage_percentages']['urls'] ?? 0,
            'url_limit_reached' => $data['warnings']['url_limit_reached'] ?? false,
            'url_approaching'   => $data['warnings']['url_limit_approaching'] ?? false,
            'custom_domain'     => $data['limits']['custom_domain']   ?? false,
            'api_access'        => $data['limits']['api_access']      ?? false,
            'analytics_days'    => $data['limits']['analytics_days']  ?? 30,
            'team_members'      => $data['limits']['team_members']    ?? 0,
        ];

        set_transient( $cache_key, $result, 5 * MINUTE_IN_SECONDS );
        return $result;
    }

    /**
     * Get basic analytics for a short code.
     *
     * @return array{total_clicks: int, countries: array, daily: array}|WP_Error
     */
    public function get_analytics( string $short_code ): array|WP_Error {
        $response = wp_remote_get(
            $this->base_url . '/analytics/' . rawurlencode( $short_code ),
            [
                'headers' => [
                    'Authorization' => 'Bearer ' . $this->api_key,
                ],
                'timeout' => 15,
            ]
        );

        if ( is_wp_error( $response ) ) {
            return $response;
        }

        $code = wp_remote_retrieve_response_code( $response );
        $data = json_decode( wp_remote_retrieve_body( $response ), true );

        if ( $code !== 200 ) {
            $msg = $data['error'] ?? 'Unknown API error';
            return new WP_Error( 'trunc_api_error', $msg, [ 'status' => $code ] );
        }

        return [
            'total_clicks' => $data['total_clicks'] ?? 0,
            'countries'    => $data['countries']     ?? [],
            'daily'        => $data['daily']         ?? [],
        ];
    }
}
