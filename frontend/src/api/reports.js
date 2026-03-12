import request from '@/utils/request'

export function fetchReports(params) {
  return request({
    url: '/analysis/reports',
    method: 'get',
    params
  })
}

export function getReport(id) {
  return request({
    url: `/analysis/reports/${id}`,
    method: 'get'
  })
}

export function fetchReportContent(id) {
  return request({
    url: `/analysis/reports/${id}/content`,
    method: 'get'
  })
}

export function generateDailyReport() {
  return request({
    url: '/analysis/report-generation/daily',
    method: 'post'
  })
}

export function generateWeeklyReport() {
  return request({
    url: '/analysis/report-generation/weekly',
    method: 'post'
  })
}

export function generateMonthlyReport() {
  return request({
    url: '/analysis/report-generation/monthly',
    method: 'post'
  })
}

export function generateCustomReport(data) {
  return request({
    url: '/analysis/report-generation/custom',
    method: 'post',
    data
  })
}

export function deleteReport(id) {
  return request({
    url: `/analysis/reports/${id}`,
    method: 'delete'
  })
}

export function bulkDeleteReports(ids) {
  return Promise.all(ids.map(id => deleteReport(id)))
}

export function downloadReport(id) {
  return request({
    url: `/analysis/report-download/${id}`,
    method: 'get',
    responseType: 'blob'
  })
}

export function getReportConfig() {
  return request({
    url: '/analysis/report-config',
    method: 'get'
  })
}

export function updateReportConfig(data) {
  return request({
    url: '/analysis/report-config',
    method: 'put',
    data
  })
}
