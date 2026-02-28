import request from '../utils/request';

// ============= User Authentication APIs =============

export function loginUser(data) {
    return request({
        method: 'POST',
        url: '/user/auth/login',
        data,
    })
}

export function registerUser(data) {
    return request({
        method: 'POST',
        url: '/user/auth/register',
        data,
    })
}

export function sendVerificationCode(data) {
    return request({
        method: 'POST',
        url: '/user/auth/send-code',
        data,
    })
}

export function resetPassword(data) {
    return request({
        method: 'POST',
        url: '/user/auth/reset',
        data,
    })
}

// ============= User Management APIs (Admin) =============

export function getUsers() {
    return request({
        method: 'GET',
        url: '/user/users',
    })
}

export function searchUsers(params) {
    return request({
        method: 'GET',
        url: '/user/users',
        params,
    })
}

export function getUserByID(id) {
    return request({
        method: 'GET',
        url: `/user/users/${id}`,
    })
}

export function updateUser(id, data) {
    return request({
        method: 'PUT',
        url: `/user/users/${id}`,
        data,
    })
}

export function addUser(data) {
    return request({
        method: 'POST',
        url: '/user/users',
        data,
    })
}

export function deleteUser(id) {
    return request({
        method: 'DELETE',
        url: `/user/users/${id}`,
    })
}

// ============= Register Application APIs (Superadmin) =============

export function searchRegisterApplications(params) {
    return request({
        method: 'GET',
        url: '/user/register-applications',
        params,
    })
}

export function approveRegisterApplication(id) {
    return request({
        method: 'PUT',
        url: `/user/register-applications/${id}/approve`,
    })
}

export function rejectRegisterApplication(id, data) {
    return request({
        method: 'PUT',
        url: `/user/register-applications/${id}/reject`,
        data,
    })
}

// ============= Password Reset Application APIs (Superadmin) =============

export function searchResetApplications(params) {
    return request({
        method: 'GET',
        url: '/user/reset-applications',
        params,
    })
}

export function approveResetApplication(id) {
    return request({
        method: 'PUT',
        url: `/user/reset-applications/${id}/approve`,
    })
}

export function rejectResetApplication(id, data) {
    return request({
        method: 'PUT',
        url: `/user/reset-applications/${id}/reject`,
        data,
    })
}

// ============= User Profile APIs (Current User) =============

export function getUserProfile() {
    return request({
        method: 'GET',
        url: '/user/user-info/me',
    })
}

export function updateUserProfile(data) {
    return request({
        method: 'PUT',
        url: '/user/user-info/me',
        data,
    })
}

export function uploadAvatar(formData, onUploadProgress) {
    return request({
        method: 'POST',
        url: '/user/user-info/me/avatar',
        data: formData,
        headers: {
            'Content-Type': 'multipart/form-data'
        },
        onUploadProgress
    })
}


// ============= Legacy/Redundant Profile APIs (Admin) =============
// These are kept for backward compatibility if needed, but they map to the same backend logic

export function getUserInformation() {
    return getUserProfile();
}

export function updateUserInformation(data) {
    return updateUserProfile(data);
}

export function createUserInformation(data) {
    return updateUserProfile(data); // Map to update
}

export function getUserInformationByUserID(id) {
    return request({
        method: 'GET',
        url: `/user/user-info/users/${id}`,
    })
}

export function updateUserInformationByUserID(id, data) {
    return request({
        method: 'PUT',
        url: `/user/user-info/users/${id}`,
        data,
    })
}
