import request from '../utils/request';

/**
 * Get main configuration
 */
export function getMainConfig() {
    return request({
        method: 'GET',
        url: '/system/config',
    });
}

/**
 * Get all configuration settings
 */
export function getAllConfig() {
    return request({
        method: 'GET',
        url: '/system/config',
    });
}

/**
 * Update main configuration
 */
export function updateConfig(data) {
    return request({
        method: 'PUT',
        url: '/system/config',
        data,
    });
}

/**
 * Save configuration to disk
 */
export function saveConfig() {
    return request({
        method: 'POST',
        url: '/system/config/save',
    });
}

/**
 * Reload configuration from disk
 */
export function reloadConfig() {
    return request({
        method: 'POST',
        url: '/system/config/reload',
    });
}

/**
 * Reset configuration to default values
 */
export function resetConfig() {
    return request({
        method: 'POST',
        url: '/system/config/reset',
    });
}
