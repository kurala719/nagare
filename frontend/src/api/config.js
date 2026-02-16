import request from '../utils/request';

/**
 * Get main configuration
 */
export function getMainConfig() {
    return request({
        method: 'GET',
        url: '/config',
    });
}

/**
 * Get all configuration settings
 */
export function getAllConfig() {
    return request({
        method: 'GET',
        url: '/config',
    });
}

/**
 * Update main configuration
 */
export function updateConfig(data) {
    return request({
        method: 'PUT',
        url: '/config',
        data,
    });
}

/**
 * Save configuration to disk
 */
export function saveConfig() {
    return request({
        method: 'POST',
        url: '/config/save',
    });
}

/**
 * Reload configuration from disk
 */
export function reloadConfig() {
    return request({
        method: 'POST',
        url: '/config/reload',
    });
}
