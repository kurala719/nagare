import request from '../utils/request';

// ============= User Authentication APIs =============

export function loginUser(data) {
    return request({
        method: 'POST',
        url: '/users/sessions',
        data,
    })
}

export function registerUser(data) {
    return request({
        method: 'POST',
        url: '/users/registrations',
        data,
    })
}

export function sendVerificationCode(data) {
    return request({
        method: 'POST',
        url: '/users/registration-codes',
        data,
    })
}

export function resetPassword(data) {
    return request({
        method: 'POST',
        url: '/users/password-resets',
        data,
    })
}

// ============= User Management APIs (Admin) =============

export function getUsers() {
    return request({
        method: 'GET',
        url: '/users',
    })
}

export function searchUsers(params) {
    return request({
        method: 'GET',
        url: '/users',
        params,
    })
}

export function getUserByID(id) {
    return request({
        method: 'GET',
        url: `/users/${id}`,
    })
}

export function updateUser(id, data) {
    return request({
        method: 'PUT',
        url: `/users/${id}`,
        data,
    })
}

export function addUser(data) {
    return request({
        method: 'POST',
        url: '/users',
        data,
    })
}

export function deleteUser(id) {
    return request({
        method: 'DELETE',
        url: `/users/${id}`,
    })
}

// ============= Register Application APIs (Superadmin) =============

export function searchRegisterApplications(params) {
    return request({
        method: 'GET',
        url: '/users/registration-applications',
        params,
    })
}

export function approveRegisterApplication(id) {
    return request({
        method: 'POST',
        url: `/users/registration-applications/${id}/approvals`,
    })
}

export function rejectRegisterApplication(id, data) {
    return request({
        method: 'POST',
        url: `/users/registration-applications/${id}/rejections`,
        data,
    })
}

// ============= Password Reset Application APIs (Superadmin) =============

export function searchResetApplications(params) {
    return request({
        method: 'GET',
        url: '/users/password-reset-applications',
        params,
    })
}

export function approveResetApplication(id) {
    return request({
        method: 'POST',
        url: `/users/password-reset-applications/${id}/approvals`,
    })
}

export function rejectResetApplication(id, data) {
    return request({
        method: 'POST',
        url: `/users/password-reset-applications/${id}/rejections`,
        data,
    })
}

// ============= User Profile APIs (Current User) =============

export function getUserProfile() {
    return request({
        method: 'GET',
        url: '/users/profile',
    })
}

export function updateUserProfile(data) {
    return request({
        method: 'PUT',
        url: '/users/profile',
        data,
    })
}

export function uploadAvatar(formData, onUploadProgress) {
    return request({
        method: 'POST',
        url: '/users/profile/avatar',
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
        url: `/users/profiles/${id}`,
    })
}

export function updateUserInformationByUserID(id, data) {
    return request({
        method: 'PUT',
        url: `/users/profiles/${id}`,
        data,
    })
}

export function getPublicStatus() {
    return request({
        method: 'GET',
        url: '/users/status',
    })
}
