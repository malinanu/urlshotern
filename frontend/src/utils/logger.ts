/**
 * Production-safe logging utility
 *
 * Replaces console.log statements with environment-aware logging.
 * Fixes Issue #11 - Remove console logging from production code.
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LoggerConfig {
  enabled: boolean;
  level: LogLevel;
  prefix?: string;
}

const LOG_LEVELS: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
};

class Logger {
  private config: LoggerConfig;

  constructor(config?: Partial<LoggerConfig>) {
    // Logging is only ever active in development. In production builds,
    // NEXT_PUBLIC_* env vars are baked into the bundle, so we hard-gate
    // on NODE_ENV rather than a runtime toggle to prevent info leakage.
    const isDevelopment = process.env.NODE_ENV === 'development';

    this.config = {
      enabled: isDevelopment,
      level: isDevelopment ? 'debug' : 'error',
      prefix: config?.prefix,
      ...config,
    };
  }

  private shouldLog(level: LogLevel): boolean {
    if (!this.config.enabled) return false;
    return LOG_LEVELS[level] >= LOG_LEVELS[this.config.level];
  }

  private formatMessage(level: LogLevel, message: string, ...args: any[]): any[] {
    const timestamp = new Date().toISOString();
    const prefix = this.config.prefix ? `[${this.config.prefix}]` : '';
    return [`[${timestamp}] ${prefix}[${level.toUpperCase()}]`, message, ...args];
  }

  debug(message: string, ...args: any[]): void {
    if (this.shouldLog('debug')) {
      console.debug(...this.formatMessage('debug', message, ...args));
    }
  }

  info(message: string, ...args: any[]): void {
    if (this.shouldLog('info')) {
      console.info(...this.formatMessage('info', message, ...args));
    }
  }

  warn(message: string, ...args: any[]): void {
    if (this.shouldLog('warn')) {
      console.warn(...this.formatMessage('warn', message, ...args));
    }
  }

  error(message: string, ...args: any[]): void {
    if (this.shouldLog('error')) {
      console.error(...this.formatMessage('error', message, ...args));
    }
  }

  /**
   * Group logging for better organization in dev tools
   */
  group(label: string, fn: () => void): void {
    if (this.config.enabled) {
      console.group(label);
      fn();
      console.groupEnd();
    } else {
      fn();
    }
  }

  /**
   * Performance timing utility
   */
  time(label: string): void {
    if (this.config.enabled) {
      console.time(label);
    }
  }

  timeEnd(label: string): void {
    if (this.config.enabled) {
      console.timeEnd(label);
    }
  }
}

// Export default logger instance
export const logger = new Logger();

// Export logger factory for creating prefixed loggers
export function createLogger(prefix: string, config?: Partial<LoggerConfig>): Logger {
  return new Logger({ ...config, prefix });
}

// Export for backward compatibility (can be used as drop-in replacement)
export default logger;
