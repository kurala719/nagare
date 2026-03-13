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
 * Get AI configuration safe for read-only pages
 */
export function getAIConfig() {
    return request({
        method: 'GET',
        url: '/ai/settings',
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
        url: '/system/config-snapshots',
    });
}

/**
 * Reload configuration from disk
 */
export function reloadConfig() {
    return request({
        method: 'POST',
        url: '/system/config-reloads',
    });
}

/**
 * Reset configuration to default values
 */
export function resetConfig() {
    return request({
        method: 'DELETE',
        url: '/system/config',
    });
}
