<?php
if ( ! defined( 'ABSPATH' ) ) {
    exit;
}

class Trunc_Admin {
    public function init(): void {
        add_action( 'admin_menu',              [ $this, 'register_menu' ] );
        add_action( 'admin_init',              [ $this, 'register_settings' ] );
        add_action( 'admin_enqueue_scripts',   [ $this, 'enqueue_assets' ] );
        add_action( 'wp_ajax_trunc_shorten',   [ $this, 'ajax_shorten' ] );
        add_action( 'wp_ajax_trunc_analytics', [ $this, 'ajax_analytics' ] );
        add_action( 'wp_ajax_trunc_test_key',  [ $this, 'ajax_test_key' ] );
    }

    public function register_menu(): void {
        add_menu_page(
            'Trunc URL Shortener',
            'Trunc Links',
            'manage_options',
            'trunc-url-shortener',
            [ $this, 'render_links_page' ],
            'dashicons-admin-links',
            80
        );
        add_submenu_page(
            'trunc-url-shortener',
            'Trunc Settings',
            'Settings',
            'manage_options',
            'trunc-settings',
            [ $this, 'render_settings_page' ]
        );
    }

    public function register_settings(): void {
        register_setting( 'trunc_options', 'trunc_api_key', [
            'type'              => 'string',
            'sanitize_callback' => [ $this, 'sanitize_api_key' ],
            'default'           => '',
        ] );
    }

    public function sanitize_api_key( string $value ): string {
        $key = sanitize_text_field( $value );
        // Clear cached usage data when key changes
        if ( $key !== get_option( 'trunc_api_key', '' ) ) {
            delete_transient( 'trunc_usage_' . substr( md5( $key ), 0, 8 ) );
        }
        return $key;
    }

    public function render_settings_page(): void {
        $api_key    = get_option( 'trunc_api_key', '' );
        $usage      = null;
        $usage_err  = '';

        if ( ! empty( $api_key ) ) {
            $api    = new Trunc_API( $api_key );
            $result = $api->get_usage();
            if ( is_wp_error( $result ) ) {
                $usage_err = $result->get_error_message();
            } else {
                $usage = $result;
            }
        }

        $plan_labels = [
            'free'              => [ 'label' => 'Free',       'class' => 'trunc-badge-free' ],
            'premium'           => [ 'label' => 'Premium',    'class' => 'trunc-badge-premium' ],
            'premium_yearly'    => [ 'label' => 'Premium',    'class' => 'trunc-badge-premium' ],
            'enterprise'        => [ 'label' => 'Enterprise', 'class' => 'trunc-badge-enterprise' ],
            'enterprise_yearly' => [ 'label' => 'Enterprise', 'class' => 'trunc-badge-enterprise' ],
        ];
        ?>
        <div class="trunc-wrap">
            <!-- Header -->
            <div class="trunc-header">
                <div class="trunc-header-logo">
                    <span class="dashicons dashicons-admin-links"></span>
                    <strong>Trunc</strong>
                </div>
                <p class="trunc-header-sub">URL Shortener &amp; Analytics</p>
            </div>

            <!-- API Key Card -->
            <div class="trunc-card">
                <h2 class="trunc-card-title">API Key</h2>
                <p class="trunc-card-desc">
                    Get your free API key at
                    <a href="https://trunc.site/dashboard/api-keys" target="_blank">trunc.site/dashboard/api-keys</a>.
                </p>
                <form method="post" action="options.php" id="trunc-settings-form">
                    <?php settings_fields( 'trunc_options' ); ?>
                    <div class="trunc-key-row">
                        <div class="trunc-key-input-wrap">
                            <input
                                type="password"
                                id="trunc-api-key-input"
                                name="trunc_api_key"
                                value="<?php echo esc_attr( $api_key ); ?>"
                                placeholder="us_••••••••••••••••"
                                class="trunc-key-input"
                                autocomplete="off"
                            />
                            <button type="button" id="trunc-toggle-key" class="trunc-btn-icon" title="Show/hide key">
                                <span class="dashicons dashicons-visibility"></span>
                            </button>
                        </div>
                        <button type="button" id="trunc-test-btn" class="trunc-btn trunc-btn-secondary">Test Connection</button>
                        <?php submit_button( 'Save', 'primary', 'submit', false, [ 'class' => 'trunc-btn trunc-btn-primary' ] ); ?>
                    </div>
                    <div id="trunc-test-result"></div>
                </form>
            </div>

            <?php if ( $usage_err && ! empty( $api_key ) ) : ?>
            <div class="trunc-notice trunc-notice-error">
                <span class="dashicons dashicons-warning"></span>
                Could not connect to Trunc: <?php echo esc_html( $usage_err ); ?>
            </div>
            <?php endif; ?>

            <?php if ( $usage ) :
                $plan_type  = $usage['plan_type'];
                $badge      = $plan_labels[ $plan_type ] ?? [ 'label' => ucfirst( $plan_type ), 'class' => 'trunc-badge-free' ];
                $max_urls   = $usage['max_urls'] === -1 ? '∞' : number_format( $usage['max_urls'] );
                $pct        = $usage['max_urls'] === -1 ? 0 : min( 100, $usage['url_pct'] );
                $bar_class  = $pct >= 90 ? 'trunc-bar-danger' : ( $pct >= 70 ? 'trunc-bar-warning' : 'trunc-bar-ok' );
            ?>
            <!-- Account Card -->
            <div class="trunc-card trunc-account-card">
                <div class="trunc-account-top">
                    <div>
                        <h2 class="trunc-card-title" style="margin:0;">Your Account</h2>
                        <p class="trunc-card-desc" style="margin:4px 0 0;">Connected successfully</p>
                    </div>
                    <span class="trunc-badge <?php echo esc_attr( $badge['class'] ); ?>"><?php echo esc_html( $badge['label'] ); ?></span>
                </div>

                <!-- URL Usage -->
                <div class="trunc-usage-block">
                    <div class="trunc-usage-label">
                        <span>Short URLs this month</span>
                        <strong><?php echo esc_html( number_format( $usage['current_urls'] ) ); ?> / <?php echo esc_html( $max_urls ); ?></strong>
                    </div>
                    <?php if ( $usage['max_urls'] !== -1 ) : ?>
                    <div class="trunc-bar-track">
                        <div class="trunc-bar-fill <?php echo esc_attr( $bar_class ); ?>" style="width:<?php echo esc_attr( $pct ); ?>%"></div>
                    </div>
                    <?php endif; ?>
                </div>

                <!-- Click Usage -->
                <div class="trunc-usage-block">
                    <div class="trunc-usage-label">
                        <span>Clicks this month</span>
                        <strong><?php echo esc_html( number_format( $usage['monthly_clicks'] ) ); ?> / <?php echo $usage['max_clicks'] === -1 ? '∞' : esc_html( number_format( $usage['max_clicks'] ) ); ?></strong>
                    </div>
                </div>

                <!-- Plan Features -->
                <div class="trunc-features-grid">
                    <?php
                    $features = [
                        [ 'label' => 'Analytics',      'on' => true ],
                        [ 'label' => 'API Access',     'on' => true ],
                        [ 'label' => 'Custom Domain',  'on' => $usage['custom_domain'] ],
                        [ 'label' => 'Team Members',   'on' => $usage['team_members'] > 0 ],
                    ];
                    foreach ( $features as $f ) :
                    ?>
                    <div class="trunc-feature-item <?php echo $f['on'] ? 'trunc-feature-on' : 'trunc-feature-off'; ?>">
                        <span class="dashicons <?php echo $f['on'] ? 'dashicons-yes-alt' : 'dashicons-dismiss'; ?>"></span>
                        <?php echo esc_html( $f['label'] ); ?>
                    </div>
                    <?php endforeach; ?>
                </div>

                <?php if ( $plan_type === 'free' ) : ?>
                <div class="trunc-upgrade-prompt">
                    <p>Upgrade to Premium for 1,000 URLs/month, 365-day analytics, API access, and custom domains.</p>
                    <a href="https://trunc.site/pricing" target="_blank" class="trunc-btn trunc-btn-primary">Upgrade Plan</a>
                </div>
                <?php endif; ?>
            </div>
            <?php endif; ?>

            <?php if ( empty( $api_key ) ) : ?>
            <div class="trunc-card trunc-onboarding-card">
                <div class="trunc-onboarding-icon">🔗</div>
                <h3>Get started in 2 steps</h3>
                <ol class="trunc-steps">
                    <li>Create a free account at <a href="https://trunc.site/register" target="_blank">trunc.site</a></li>
                    <li>Copy your API key from <a href="https://trunc.site/dashboard/api-keys" target="_blank">Dashboard → API Keys</a> and paste it above</li>
                </ol>
            </div>
            <?php endif; ?>
        </div>
        <?php
    }

    public function render_links_page(): void {
        $api_key = get_option( 'trunc_api_key', '' );

        if ( empty( $api_key ) ) {
            echo '<div class="trunc-wrap"><div class="trunc-notice trunc-notice-info">';
            echo '<span class="dashicons dashicons-info"></span> ';
            echo 'Please <a href="' . esc_url( admin_url( 'admin.php?page=trunc-settings' ) ) . '">configure your API key</a> first.';
            echo '</div></div>';
            return;
        }

        $api     = new Trunc_API( $api_key );
        $usage   = $api->get_usage();
        $blocked = false;

        global $wpdb;
        $table = $wpdb->prefix . 'trunc_links';
        $links = $wpdb->get_results( "SELECT * FROM $table ORDER BY created_at DESC LIMIT 100" ); // phpcs:ignore WordPress.DB.PreparedSQL.InterpolatedNotPrepared
        ?>
        <div class="trunc-wrap">
            <div class="trunc-header">
                <div class="trunc-header-logo">
                    <span class="dashicons dashicons-admin-links"></span>
                    <strong>Trunc Links</strong>
                </div>
                <?php if ( ! is_wp_error( $usage ) ) :
                    $plan_labels = [
                        'free' => 'Free', 'premium' => 'Premium', 'premium_yearly' => 'Premium',
                        'enterprise' => 'Enterprise', 'enterprise_yearly' => 'Enterprise',
                    ];
                    $plan_type = $usage['plan_type'];
                    $plan_name = $plan_labels[ $plan_type ] ?? ucfirst( $plan_type );
                    $badge_class = strpos( $plan_type, 'enterprise' ) !== false ? 'trunc-badge-enterprise' : ( $plan_type === 'free' ? 'trunc-badge-free' : 'trunc-badge-premium' );
                    $blocked = $usage['url_limit_reached'];
                ?>
                <div class="trunc-header-meta">
                    <span class="trunc-badge <?php echo esc_attr( $badge_class ); ?>"><?php echo esc_html( $plan_name ); ?></span>
                    <span class="trunc-header-usage">
                        <?php echo esc_html( $usage['current_urls'] ); ?> /
                        <?php echo $usage['max_urls'] === -1 ? '∞' : esc_html( $usage['max_urls'] ); ?> URLs
                    </span>
                </div>
                <?php endif; ?>
            </div>

            <?php if ( ! is_wp_error( $usage ) && $usage['url_limit_reached'] ) : ?>
            <div class="trunc-notice trunc-notice-error">
                <span class="dashicons dashicons-warning"></span>
                You have reached your <?php echo esc_html( $usage['max_urls'] ); ?> URL limit for this month.
                <a href="https://trunc.site/pricing" target="_blank">Upgrade your plan</a> to shorten more links.
            </div>
            <?php elseif ( ! is_wp_error( $usage ) && $usage['url_approaching'] ) : ?>
            <div class="trunc-notice trunc-notice-warning">
                <span class="dashicons dashicons-clock"></span>
                Approaching your monthly URL limit (<?php echo esc_html( $usage['current_urls'] ); ?> of <?php echo esc_html( $usage['max_urls'] ); ?> used).
                <a href="https://trunc.site/pricing" target="_blank">Consider upgrading</a>.
            </div>
            <?php endif; ?>

            <!-- Shorten Form -->
            <div class="trunc-card">
                <h2 class="trunc-card-title">Shorten a URL</h2>
                <div class="trunc-shorten-form <?php echo $blocked ? 'trunc-form-disabled' : ''; ?>">
                    <input
                        type="url"
                        id="trunc-original-url"
                        placeholder="https://example.com/your-long-url"
                        class="trunc-input"
                        <?php echo $blocked ? 'disabled' : ''; ?>
                    />
                    <input
                        type="text"
                        id="trunc-custom-code"
                        placeholder="custom-code (optional)"
                        class="trunc-input trunc-input-short"
                        <?php echo $blocked ? 'disabled' : ''; ?>
                    />
                    <button
                        id="trunc-shorten-btn"
                        class="trunc-btn trunc-btn-primary"
                        <?php echo $blocked ? 'disabled' : ''; ?>
                    >
                        <span class="dashicons dashicons-admin-links"></span> Shorten
                    </button>
                </div>
                <div id="trunc-shorten-result" class="trunc-shorten-result"></div>
            </div>

            <!-- Links Table -->
            <div class="trunc-card">
                <h2 class="trunc-card-title">Your Short Links</h2>
                <?php if ( empty( $links ) ) : ?>
                <div class="trunc-empty-state">
                    <span class="dashicons dashicons-admin-links trunc-empty-icon"></span>
                    <p>No links yet. Shorten your first URL above.</p>
                </div>
                <?php else : ?>
                <table class="trunc-table">
                    <thead>
                        <tr>
                            <th>Short URL</th>
                            <th>Destination</th>
                            <th>Clicks</th>
                            <th>Created</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <?php foreach ( $links as $link ) : ?>
                        <tr>
                            <td class="trunc-td-short">
                                <a href="<?php echo esc_url( $link->short_url ); ?>" target="_blank" class="trunc-short-link">
                                    <?php echo esc_html( $link->short_url ); ?>
                                </a>
                                <button class="trunc-copy-btn" data-url="<?php echo esc_attr( $link->short_url ); ?>" title="Copy">
                                    <span class="dashicons dashicons-clipboard"></span>
                                </button>
                            </td>
                            <td class="trunc-td-dest" title="<?php echo esc_attr( $link->original_url ); ?>">
                                <?php echo esc_html( $link->original_url ); ?>
                            </td>
                            <td class="trunc-td-clicks">
                                <span class="trunc-clicks-badge" data-code="<?php echo esc_attr( $link->short_code ); ?>">
                                    <?php echo esc_html( number_format( (int) $link->clicks ) ); ?>
                                </span>
                            </td>
                            <td class="trunc-td-date">
                                <?php echo esc_html( date_i18n( 'M j, Y', strtotime( $link->created_at ) ) ); ?>
                            </td>
                            <td class="trunc-td-actions">
                                <a href="https://trunc.site/analytics/<?php echo esc_attr( $link->short_code ); ?>" target="_blank" class="trunc-btn trunc-btn-ghost">
                                    <span class="dashicons dashicons-chart-bar"></span> Analytics
                                </a>
                            </td>
                        </tr>
                        <?php endforeach; ?>
                    </tbody>
                </table>
                <?php endif; ?>
            </div>
        </div>
        <?php
    }

    public function enqueue_assets( string $hook ): void {
        if ( ! str_contains( $hook, 'trunc' ) ) {
            return;
        }
        wp_enqueue_style( 'trunc-admin', TRUNC_PLUGIN_URL . 'admin/admin.css', [], TRUNC_VERSION );
        wp_enqueue_script( 'trunc-admin', TRUNC_PLUGIN_URL . 'admin/admin.js', [ 'jquery' ], TRUNC_VERSION, true );
        wp_localize_script( 'trunc-admin', 'truncAjax', [
            'ajaxurl' => admin_url( 'admin-ajax.php' ),
            'nonce'   => wp_create_nonce( 'trunc_nonce' ),
        ] );
    }

    public function ajax_test_key(): void {
        check_ajax_referer( 'trunc_nonce', 'nonce' );
        if ( ! current_user_can( 'manage_options' ) ) {
            wp_send_json_error( 'Unauthorized' );
        }

        $api_key = sanitize_text_field( wp_unslash( $_POST['api_key'] ?? '' ) );
        if ( empty( $api_key ) ) {
            wp_send_json_error( 'Please enter an API key first.' );
        }

        $api    = new Trunc_API( $api_key );
        $result = $api->get_usage();

        if ( is_wp_error( $result ) ) {
            wp_send_json_error( $result->get_error_message() );
        }

        $plan_labels = [
            'free'              => 'Free',
            'premium'           => 'Premium',
            'premium_yearly'    => 'Premium',
            'enterprise'        => 'Enterprise',
            'enterprise_yearly' => 'Enterprise',
        ];

        wp_send_json_success( [
            'plan_name'    => $plan_labels[ $result['plan_type'] ] ?? ucfirst( $result['plan_type'] ),
            'current_urls' => $result['current_urls'],
            'max_urls'     => $result['max_urls'] === -1 ? '∞' : $result['max_urls'],
        ] );
    }

    public function ajax_shorten(): void {
        check_ajax_referer( 'trunc_nonce', 'nonce' );
        if ( ! current_user_can( 'manage_options' ) ) {
            wp_send_json_error( 'Unauthorized' );
        }

        $url         = esc_url_raw( wp_unslash( $_POST['url'] ?? '' ) );
        $custom_code = sanitize_text_field( wp_unslash( $_POST['custom_code'] ?? '' ) );

        if ( ! filter_var( $url, FILTER_VALIDATE_URL ) ) {
            wp_send_json_error( 'Please enter a valid URL.' );
        }

        $api_key = get_option( 'trunc_api_key', '' );
        if ( empty( $api_key ) ) {
            wp_send_json_error( 'No API key configured. Go to Trunc Links → Settings.' );
        }

        $api = new Trunc_API( $api_key );

        // Check limit before shortening
        $usage = $api->get_usage();
        if ( ! is_wp_error( $usage ) && $usage['url_limit_reached'] ) {
            wp_send_json_error( 'Monthly URL limit reached. Upgrade your plan at trunc.site/pricing.' );
        }

        $result = $api->shorten( $url, $custom_code );

        if ( is_wp_error( $result ) ) {
            wp_send_json_error( $result->get_error_message() );
        }

        global $wpdb;
        $wpdb->insert(
            $wpdb->prefix . 'trunc_links',
            [
                'original_url' => $url,
                'short_code'   => $result['short_code'],
                'short_url'    => $result['short_url'],
                'clicks'       => 0,
                'created_at'   => current_time( 'mysql' ),
            ]
        );

        // Clear usage cache so count updates
        delete_transient( 'trunc_usage_' . substr( md5( $api_key ), 0, 8 ) );

        wp_send_json_success( $result );
    }

    public function ajax_analytics(): void {
        check_ajax_referer( 'trunc_nonce', 'nonce' );
        if ( ! current_user_can( 'manage_options' ) ) {
            wp_send_json_error( 'Unauthorized' );
        }

        $short_code = sanitize_text_field( wp_unslash( $_POST['short_code'] ?? '' ) );
        $api_key    = get_option( 'trunc_api_key', '' );
        $api        = new Trunc_API( $api_key );
        $result     = $api->get_analytics( $short_code );

        if ( is_wp_error( $result ) ) {
            wp_send_json_error( $result->get_error_message() );
        }

        global $wpdb;
        $wpdb->update(
            $wpdb->prefix . 'trunc_links',
            [ 'clicks' => $result['total_clicks'] ],
            [ 'short_code' => $short_code ]
        );

        wp_send_json_success( $result );
    }
}
