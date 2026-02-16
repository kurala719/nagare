import request from '../utils/request';

// ============= User Information APIs =============

/**
 * Get current user's information
 */
export function getUserInformation() {
    return request({
        method: 'GET',
        url: '/user-info/me',
    })
}

/**
 * Create current user's information
 */
export function createUserInformation(data) {
    return request({
        method: 'POST',
        url: '/user-info/me',
        data,
    })
}

/**
 * Update current user's information
 */
export function updateUserInformation(data) {
    return request({
        method: 'PUT',
        url: '/user-info/me',
        data,
    })
}

/**
 * Delete current user's information
 */
export function deleteUserInformation() {
    return request({
        method: 'DELETE',
        url: '/user-info/me',
    })
}

/**
 * Get user information by user ID (admin only)
 */
export function getUserInformationByUserID(userId) {
    return request({
        method: 'GET',
        url: `/user-info/users/${userId}`,
    })
}

/**
 * Update user information by user ID (superadmin only)
 */
export function updateUserInformationByUserID(userId, data) {
    return request({
        method: 'PUT',
        url: `/user-info/users/${userId}`,
        data,
    })
}
