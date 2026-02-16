import request from '../utils/request';

// ============= User Authentication APIs =============

export function loginUser(data) {
    return request({
        method: 'POST',
        url: '/auth/login',
        data,
    })
}

export function registerUser(data) {
    return request({
        method: 'POST',
        url: '/auth/register',
        data,
    })
}

export function resetPassword(data) {
    return request({
        method: 'POST',
        url: '/auth/reset',
        data,
    })
}

// ============= User Management APIs (Admin) =============

export function getUsers() {
    return request({
        method: 'GET',
        url: '/users/',
    })
}

export function searchUsers(params) {
    return request({
        method: 'GET',
        url: '/users/',
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
        url: '/users/',
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
        url: '/register-applications/',
        params,
    })
}

export function approveRegisterApplication(id) {
    return request({
        method: 'PUT',
        url: `/register-applications/${id}/approve`,
    })
}

export function rejectRegisterApplication(id, data) {
    return request({
        method: 'PUT',
        url: `/register-applications/${id}/reject`,
        data,
    })
}