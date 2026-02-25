package server

const adminHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Oculus-Feldsher Admin</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        .tab-active { border-bottom: 2px solid #3b82f6; color: #3b82f6; font-weight: 600; }
        .loader { border: 3px solid #f3f4f6; border-top: 3px solid #3b82f6; border-radius: 50%; width: 24px; height: 24px; animation: spin 0.8s linear infinite; display: inline-block; }
        @keyframes spin { to { transform: rotate(360deg); } }
        .status-red { background: #fee; color: #c00; border-left: 4px solid #c00; }
        .status-yellow { background: #ffc; color: #960; border-left: 4px solid #fa0; }
        .status-green { background: #efe; color: #060; border-left: 4px solid #0a0; }
        .circular-chart { transform: rotate(-90deg); }
        .circle { stroke-linecap: round; transition: stroke-dasharray 0.3s ease; }
        @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
        @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
        .animate-fadeIn { animation: fadeIn 0.2s ease-out; }
        .animate-slideUp { animation: slideUp 0.3s ease-out; }
    </style>
</head>
<body class="bg-gray-50 min-h-screen">

<div id="app"></div>

<script>
const API = '/api/v1';
let token = localStorage.getItem('admin_token') || '';
let refreshToken = localStorage.getItem('admin_refresh_token') || '';
let currentTab = 'dashboard';
let isRefreshing = false;
let currentPage = 1;
let pageSize = 20;

// Helper function to safely display values
function safe(value, fallback = '—') {
    if (value === undefined || value === null || value === '') return fallback;
    return value;
}

const roleNames = {
    'ADMIN': 'Администратор',
    'CALL_CENTER': 'Колл-центр',
    'DISTRICT_DOCTOR': 'Районный врач',
    'SURGEON': 'Хирург',
    'PATIENT': 'Пациент'
};

const statusColors = {
    'NEW': 'status-yellow',
    'PREPARATION': 'status-yellow',
    'REVIEW_NEEDED': 'status-yellow',
    'APPROVED': 'status-green',
    'REJECTED': 'status-red',
    'SCHEDULED': 'status-green'
};

function getRoleName(role) {
    return roleNames[role] || role;
}

function getStatusColor(status) {
    return statusColors[status] || '';
}

async function refreshAccessToken() {
    if (isRefreshing) return;
    isRefreshing = true;
    try {
        const response = await fetch(API + '/auth/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken })
        });
        if (!response.ok) throw new Error('Refresh failed');
        const jsonResponse = await response.json();
        const data = jsonResponse.data || jsonResponse;
        token = data.access_token;
        refreshToken = data.refresh_token;
        localStorage.setItem('admin_token', token);
        localStorage.setItem('admin_refresh_token', refreshToken);
        return true;
    } catch (err) {
        logout();
        return false;
    } finally {
        isRefreshing = false;
    }
}

function api(path, opts = {}) {
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;
    return fetch(API + path, { ...opts, headers }).then(async r => {
        if (r.status === 401 && refreshToken && !opts._retry) {
            const refreshed = await refreshAccessToken();
            if (refreshed) {
                opts._retry = true;
                return api(path, opts);
            }
        }
        const data = await r.json();
        if (!r.ok) throw new Error(data.error || r.statusText);
        return data;
    });
}

function render() {
    if (!token || !refreshToken) return renderLogin();
    renderApp();
}

function renderLogin() {
    document.getElementById('app').innerHTML = ` + "`" + `
    <div class="flex items-center justify-center min-h-screen">
        <div class="bg-white rounded-xl shadow-lg p-8 w-full max-w-sm">
            <h1 class="text-2xl font-bold text-center mb-2">Oculus-Feldsher</h1>
            <p class="text-gray-500 text-center mb-6">Панель администратора</p>
            <div id="login-error" class="hidden bg-red-50 text-red-600 rounded p-3 mb-4 text-sm"></div>
            <form id="login-form" class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
                    <input id="email" type="email" value="admin@gmail.com" class="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500" required>
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
                    <input id="password" type="password" class="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500" required>
                </div>
                <button type="submit" class="w-full bg-blue-600 text-white rounded-lg py-2 font-medium hover:bg-blue-700 transition">Войти</button>
            </form>
        </div>
    </div>
    ` + "`" + `;
    document.getElementById('login-form').onsubmit = async (e) => {
        e.preventDefault();
        const errEl = document.getElementById('login-error');
        errEl.classList.add('hidden');
        try {
            const response = await api('/auth/login', {
                method: 'POST',
                body: JSON.stringify({
                    email: document.getElementById('email').value,
                    password: document.getElementById('password').value
                })
            });
            const data = response.data || response;
            if (data.user.role !== 'ADMIN') throw new Error('Доступ только для администраторов');
            token = data.access_token;
            refreshToken = data.refresh_token;
            localStorage.setItem('admin_token', token);
            localStorage.setItem('admin_refresh_token', refreshToken);
            render();
        } catch (err) {
            errEl.textContent = err.message;
            errEl.classList.remove('hidden');
        }
    };
}

function logout() {
    token = '';
    refreshToken = '';
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_refresh_token');
    render();
}

function switchTab(tab) {
    currentTab = tab;
    renderApp();
}

async function renderApp() {
    document.getElementById('app').innerHTML = ` + "`" + `
    <nav class="bg-white shadow">
        <div class="max-w-7xl mx-auto px-4 flex items-center justify-between h-14">
            <span class="font-bold text-lg">Oculus-Feldsher Admin</span>
            <button onclick="logout()" class="text-sm text-red-600 hover:text-red-800">Выйти</button>
        </div>
    </nav>
    <div class="max-w-7xl mx-auto px-4 mt-4">
        <div class="flex gap-6 border-b mb-6">
            <button onclick="switchTab('dashboard')" class="pb-2 px-1 text-sm ${currentTab==='dashboard'?'tab-active':'text-gray-500 hover:text-gray-700'}">Дашборд</button>
            <button onclick="switchTab('districts')" class="pb-2 px-1 text-sm ${currentTab==='districts'?'tab-active':'text-gray-500 hover:text-gray-700'}">Районы</button>
            <button onclick="switchTab('users')" class="pb-2 px-1 text-sm ${currentTab==='users'?'tab-active':'text-gray-500 hover:text-gray-700'}">Пользователи</button>
            <button onclick="switchTab('patients')" class="pb-2 px-1 text-sm ${currentTab==='patients'?'tab-active':'text-gray-500 hover:text-gray-700'}">Пациенты</button>
            <button onclick="switchTab('surgeries')" class="pb-2 px-1 text-sm ${currentTab==='surgeries'?'tab-active':'text-gray-500 hover:text-gray-700'}">Операции</button>
        </div>
        <div id="tab-content"><div class="loader mx-auto mt-12"></div></div>
    </div>
    ` + "`" + `;
    try {
        if (currentTab === 'dashboard') await renderDashboard();
        else if (currentTab === 'districts') await renderDistricts();
        else if (currentTab === 'users') await renderUsers();
        else if (currentTab === 'patients') await renderPatients();
        else if (currentTab === 'surgeries') await renderSurgeries();
    } catch (err) {
        if (err.message.includes('expired') || err.message.includes('invalid')) {
            logout();
        } else {
            document.getElementById('tab-content').innerHTML = '<p class="text-red-500">Ошибка: ' + err.message + '</p>';
        }
    }
}

async function renderDashboard() {
    const statsResponse = await api('/admin/stats');
    const stats = statsResponse.data || statsResponse;
    document.getElementById('tab-content').innerHTML = ` + "`" + `
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="bg-white rounded-xl shadow p-6 text-center">
            <div class="text-3xl font-bold text-blue-600">${safe(stats.users, 0)}</div>
            <div class="text-gray-500 mt-1">Пользователи</div>
        </div>
        <div class="bg-white rounded-xl shadow p-6 text-center">
            <div class="text-3xl font-bold text-green-600">${safe(stats.patients, 0)}</div>
            <div class="text-gray-500 mt-1">Пациенты</div>
        </div>
        <div class="bg-white rounded-xl shadow p-6 text-center">
            <div class="text-3xl font-bold text-purple-600">${safe(stats.districts, 0)}</div>
            <div class="text-gray-500 mt-1">Районы</div>
        </div>
        <div class="bg-white rounded-xl shadow p-6 text-center">
            <div class="text-3xl font-bold text-orange-600">${safe(stats.surgeries, 0)}</div>
            <div class="text-gray-500 mt-1">Операции</div>
        </div>
    </div>
    ` + "`" + `;
}

async function renderDistricts(page = 1) {
    const response = await api('/districts?page=' + page + '&limit=' + pageSize);
    const districts = response.data || response;
    const meta = response.meta || {};
    let html = ` + "`" + `
    <div class="mb-4">
        <button onclick="showCreateDistrict()" class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-700">+ Добавить район</button>
    </div>
    <div id="district-form-area"></div>
    <div class="bg-white rounded-xl shadow overflow-hidden">
        <table class="w-full text-sm">
            <thead class="bg-gray-50">
                <tr>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">ID</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Название</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Регион</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Код</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Часовой пояс</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Действия</th>
                </tr>
            </thead>
            <tbody>
    ` + "`" + `;
    if (Array.isArray(districts)) {
        districts.forEach(d => {
            html += ` + "`" + `<tr class="border-t">
                <td class="px-4 py-3">${safe(d.id)}</td>
                <td class="px-4 py-3 font-medium">${safe(d.name)}</td>
                <td class="px-4 py-3">${safe(d.region)}</td>
                <td class="px-4 py-3"><span class="bg-gray-100 px-2 py-0.5 rounded text-xs">${safe(d.code)}</span></td>
                <td class="px-4 py-3">${safe(d.timezone)}</td>
                <td class="px-4 py-3 space-x-2">
                    <button onclick='editDistrict(${JSON.stringify(d).replace(/'/g,"&#39;")})' class="text-blue-600 hover:underline text-xs">Изменить</button>
                    <button onclick="deleteDistrict(${safe(d.id)})" class="text-red-600 hover:underline text-xs">Удалить</button>
                </td>
            </tr>` + "`" + `;
        });
    }
    html += '</tbody></table>';
    if (meta.total_pages > 1) {
        html += '<div class="px-4 py-3 border-t flex items-center justify-between">';
        html += '<div class="text-sm text-gray-600">Страница ' + page + ' из ' + meta.total_pages + '</div>';
        html += '<div class="flex gap-2">';
        if (page > 1) html += '<button onclick="renderDistricts(' + (page-1) + ')" class="px-3 py-1 border rounded text-sm hover:bg-gray-50">Назад</button>';
        if (page < meta.total_pages) html += '<button onclick="renderDistricts(' + (page+1) + ')" class="px-3 py-1 border rounded text-sm hover:bg-gray-50">Вперёд</button>';
        html += '</div></div>';
    }
    html += '</div>';
    document.getElementById('tab-content').innerHTML = html;
}

function showCreateDistrict() {
    document.getElementById('district-form-area').innerHTML = districtForm({}, 'createDistrict');
}

function editDistrict(d) {
    document.getElementById('district-form-area').innerHTML = districtForm(d, 'updateDistrict');
}

function districtForm(d, fn) {
    return ` + "`" + `
    <div class="bg-white rounded-xl shadow p-4 mb-4">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
            <input id="df-name" placeholder="Название" value="${safe(d.name, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="df-region" placeholder="Регион" value="${safe(d.region, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="df-code" placeholder="Код" value="${safe(d.code, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="df-tz" placeholder="Часовой пояс" value="${safe(d.timezone, '')}" class="border rounded px-3 py-2 text-sm">
        </div>
        <div class="mt-3 flex gap-2">
            <button onclick="${fn}(${d.id||0})" class="bg-blue-600 text-white px-4 py-1.5 rounded text-sm hover:bg-blue-700">Сохранить</button>
            <button onclick="document.getElementById('district-form-area').innerHTML=''" class="bg-gray-200 px-4 py-1.5 rounded text-sm hover:bg-gray-300">Отмена</button>
        </div>
    </div>
    ` + "`" + `;
}

async function createDistrict() {
    await api('/districts', {
        method: 'POST',
        body: JSON.stringify({
            name: document.getElementById('df-name').value,
            region: document.getElementById('df-region').value,
            code: document.getElementById('df-code').value,
            timezone: document.getElementById('df-tz').value
        })
    });
    await renderDistricts();
}

async function updateDistrict(id) {
    await api('/districts/' + id, {
        method: 'PATCH',
        body: JSON.stringify({
            name: document.getElementById('df-name').value,
            region: document.getElementById('df-region').value,
            code: document.getElementById('df-code').value,
            timezone: document.getElementById('df-tz').value
        })
    });
    await renderDistricts();
}

async function deleteDistrict(id) {
    if (!confirm('Удалить район?')) return;
    await api('/districts/' + id, { method: 'DELETE' });
    await renderDistricts();
}

async function renderUsers(page = 1) {
    const usersResponse = await api('/admin/users?page=' + page + '&limit=' + pageSize);
    const users = usersResponse.data || usersResponse;
    const meta = usersResponse.meta || {};
    const districtsResponse = await api('/districts?limit=100');
    const districts = districtsResponse.data || districtsResponse;
    let html = ` + "`" + `
    <div class="mb-4">
        <button onclick="showCreateUser()" class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-700">+ Добавить пользователя</button>
    </div>
    <div id="user-form-area"></div>
    <div class="bg-white rounded-xl shadow overflow-hidden">
        <table class="w-full text-sm">
            <thead class="bg-gray-50">
                <tr>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">ID</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Имя</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Email</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Телефон</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Роль</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Район</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Статус</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Действия</th>
                </tr>
            </thead>
            <tbody>
    ` + "`" + `;
    const roleBadge = { ADMIN: 'bg-red-100 text-red-700', CALL_CENTER: 'bg-yellow-100 text-yellow-700', SURGEON: 'bg-blue-100 text-blue-700', DISTRICT_DOCTOR: 'bg-green-100 text-green-700', PATIENT: 'bg-gray-100 text-gray-700' };
    if (Array.isArray(users)) {
        users.forEach(u => {
            const dist = districts.find(d => d.id === u.district_id);
            html += ` + "`" + `<tr class="border-t">
                <td class="px-4 py-3">${safe(u.id)}</td>
                <td class="px-4 py-3 font-medium">${safe(u.name)}</td>
                <td class="px-4 py-3">${safe(u.email)}</td>
                <td class="px-4 py-3">${safe(u.phone)}</td>
                <td class="px-4 py-3"><span class="px-2 py-0.5 rounded text-xs ${roleBadge[u.role]||'bg-gray-100'}">${getRoleName(u.role)}</span></td>
                <td class="px-4 py-3">${dist ? dist.name : '—'}</td>
                <td class="px-4 py-3">${u.is_active ? '<span class="text-green-600">✓</span>' : '<span class="text-red-600">✗</span>'}</td>
                <td class="px-4 py-3 space-x-2">
                    <button onclick='editUser(${JSON.stringify(u).replace(/'/g,"&#39;")})' class="text-blue-600 hover:underline text-xs">Изменить</button>
                    <button onclick="deleteUser(${safe(u.id)})" class="text-red-600 hover:underline text-xs">Удалить</button>
                </td>
            </tr>` + "`" + `;
        });
    }
    html += '</tbody></table>';
    if (meta.total_pages > 1) {
        html += '<div class="px-4 py-3 border-t flex items-center justify-between">';
        html += '<div class="text-sm text-gray-600">Страница ' + page + ' из ' + meta.total_pages + '</div>';
        html += '<div class="flex gap-2">';
        if (page > 1) html += '<button onclick="renderUsers(' + (page-1) + ')" class="px-3 py-1 border rounded text-sm hover:bg-gray-50">Назад</button>';
        if (page < meta.total_pages) html += '<button onclick="renderUsers(' + (page+1) + ')" class="px-3 py-1 border rounded text-sm hover:bg-gray-50">Вперёд</button>';
        html += '</div></div>';
    }
    html += '</div>';
    document.getElementById('tab-content').innerHTML = html;
    window.allDistricts = districts;
}

function showCreateUser() {
    document.getElementById('user-form-area').innerHTML = userForm({}, 'createUser');
}

function editUser(u) {
    document.getElementById('user-form-area').innerHTML = userForm(u, 'updateUser');
}

function userForm(u, fn) {
    const districts = window.allDistricts || [];
    return ` + "`" + `
    <div class="bg-white rounded-xl shadow p-4 mb-4">
        <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <input id="uf-email" placeholder="Email" value="${safe(u.email, '')}" class="border rounded px-3 py-2 text-sm" ${u.id?'disabled':''}>
            <input id="uf-name" placeholder="Полное имя" value="${safe(u.name, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="uf-fname" placeholder="Имя" value="${safe(u.first_name, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="uf-lname" placeholder="Фамилия" value="${safe(u.last_name, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="uf-mname" placeholder="Отчество" value="${safe(u.middle_name, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="uf-phone" placeholder="Телефон" value="${safe(u.phone, '')}" class="border rounded px-3 py-2 text-sm">
            <select id="uf-role" class="border rounded px-3 py-2 text-sm">
                <option value="PATIENT" ${u.role==='PATIENT'?'selected':''}>Пациент</option>
                <option value="DISTRICT_DOCTOR" ${u.role==='DISTRICT_DOCTOR'?'selected':''}>Районный врач</option>
                <option value="SURGEON" ${u.role==='SURGEON'?'selected':''}>Хирург</option>
                <option value="CALL_CENTER" ${u.role==='CALL_CENTER'?'selected':''}>Колл-центр</option>
                <option value="ADMIN" ${u.role==='ADMIN'?'selected':''}>Администратор</option>
            </select>
            <select id="uf-district" class="border rounded px-3 py-2 text-sm">
                <option value="">Без района</option>
                ${districts.map(d => ` + "`" + `<option value="${safe(d.id)}" ${u.district_id===d.id?'selected':''}>${safe(d.name)}</option>` + "`" + `).join('')}
            </select>
            ${!u.id ? '<input id="uf-password" type="password" placeholder="Пароль" class="border rounded px-3 py-2 text-sm">' : ''}
        </div>
        <div class="mt-3 flex gap-2">
            <button onclick="${fn}(${u.id||0})" class="bg-blue-600 text-white px-4 py-1.5 rounded text-sm hover:bg-blue-700">Сохранить</button>
            <button onclick="document.getElementById('user-form-area').innerHTML=''" class="bg-gray-200 px-4 py-1.5 rounded text-sm hover:bg-gray-300">Отмена</button>
        </div>
    </div>
    ` + "`" + `;
}

async function createUser() {
    const districtId = document.getElementById('uf-district').value;
    await api('/auth/register', {
        method: 'POST',
        body: JSON.stringify({
            email: document.getElementById('uf-email').value,
            password: document.getElementById('uf-password').value,
            name: document.getElementById('uf-name').value,
            first_name: document.getElementById('uf-fname').value,
            last_name: document.getElementById('uf-lname').value,
            middle_name: document.getElementById('uf-mname').value,
            phone: document.getElementById('uf-phone').value,
            role: document.getElementById('uf-role').value,
            district_id: districtId ? parseInt(districtId) : null
        })
    });
    await renderUsers();
}

async function updateUser(id) {
    alert('Редактирование пользователей пока не реализовано в API');
}

async function deleteUser(id) {
    if (!confirm('Удалить пользователя?')) return;
    alert('Удаление пользователей пока не реализовано в API');
}

async function renderPatients(page = 1) {
    const patientsResponse = await api('/patients?page=' + page + '&limit=' + pageSize);
    const patients = patientsResponse.data || patientsResponse;
    const meta = patientsResponse.meta || {};
    const districtsResponse = await api('/districts?limit=100');
    const districts = districtsResponse.data || districtsResponse;
    let items = Array.isArray(patients) ? patients : (patients.patients || []);
    let html = ` + "`" + `
    <div class="mb-4">
        <button onclick="showCreatePatient()" class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-700">+ Добавить пациента</button>
    </div>
    <div id="patient-form-area"></div>
    <div class="bg-white rounded-xl shadow overflow-hidden">
        <table class="w-full text-sm">
            <thead class="bg-gray-50">
                <tr>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">ID</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">ФИО</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Телефон</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Диагноз</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Операция</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Глаз</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Статус</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Действия</th>
                </tr>
            </thead>
            <tbody>
    ` + "`" + `;
    const statusBadge = { PREPARATION: 'bg-yellow-100 text-yellow-700', REVIEW_NEEDED: 'bg-blue-100 text-blue-700', APPROVED: 'bg-green-100 text-green-700', REJECTED: 'bg-red-100 text-red-700', SCHEDULED: 'bg-purple-100 text-purple-700' };
    items.forEach(p => {
        const rowColor = getStatusColor(p.status);
        html += ` + "`" + `<tr class="border-t ${rowColor} hover:bg-blue-50 cursor-pointer" onclick="showPatientDetails(${safe(p.id)})">
            <td class="px-4 py-3">${safe(p.id)}</td>
            <td class="px-4 py-3 font-medium">${safe(p.last_name)} ${safe(p.first_name)} ${safe(p.middle_name, '')}</td>
            <td class="px-4 py-3">${safe(p.phone)}</td>
            <td class="px-4 py-3 max-w-xs truncate">${safe(p.diagnosis)}</td>
            <td class="px-4 py-3"><span class="text-xs">${safe(p.operation_type)}</span></td>
            <td class="px-4 py-3">${safe(p.eye)}</td>
            <td class="px-4 py-3"><span class="px-2 py-0.5 rounded text-xs ${statusBadge[p.status]||'bg-gray-100'}">${p.status}</span></td>
            <td class="px-4 py-3 space-x-2" onclick="event.stopPropagation()">
                <button onclick='editPatient(${JSON.stringify(p).replace(/'/g,"&#39;")})' class="text-blue-600 hover:underline text-xs">Изменить</button>
                <button onclick="deletePatient(${safe(p.id)})" class="text-red-600 hover:underline text-xs">Удалить</button>
            </td>
        </tr>` + "`" + `;
    });
    html += '</tbody></table>';
    if (meta.total_pages > 1) {
        html += '<div class="px-4 py-3 border-t flex items-center justify-between">';
        html += '<div class="text-sm text-gray-600">Страница ' + page + ' из ' + meta.total_pages + '</div>';
        html += '<div class="flex gap-2">';
        if (page > 1) html += '<button onclick="renderPatients(' + (page-1) + ')" class="px-3 py-1 border rounded text-sm hover:bg-gray-50">Назад</button>';
        if (page < meta.total_pages) html += '<button onclick="renderPatients(' + (page+1) + ')" class="px-3 py-1 border rounded text-sm hover:bg-gray-50">Вперёд</button>';
        html += '</div></div>';
    }
    html += '</div>';
    document.getElementById('tab-content').innerHTML = html;
    window.allDistricts = districts;
}

function showCreatePatient() {
    document.getElementById('patient-form-area').innerHTML = patientForm({}, 'createPatient');
}

function editPatient(p) {
    document.getElementById('patient-form-area').innerHTML = patientForm(p, 'updatePatient');
}

function patientForm(p, fn) {
    const districts = window.allDistricts || [];
    return ` + "`" + `
    <div class="bg-white rounded-xl shadow p-4 mb-4">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
            <input id="pf-fname" placeholder="Имя" value="${safe(p.first_name, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-lname" placeholder="Фамилия" value="${safe(p.last_name, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-mname" placeholder="Отчество" value="${safe(p.middle_name, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-phone" placeholder="Телефон" value="${safe(p.phone, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-email" placeholder="Email" value="${safe(p.email, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-dob" type="date" placeholder="Дата рождения" value="${safe(p.date_of_birth, '')}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-diagnosis" placeholder="Диагноз" value="${safe(p.diagnosis, '')}" class="border rounded px-3 py-2 text-sm col-span-2">
            <select id="pf-optype" class="border rounded px-3 py-2 text-sm">
                <option value="PHACO" ${p.operation_type==='PHACO'?'selected':''}>PHACO</option>
                <option value="ANTIGLAUCOMA" ${p.operation_type==='ANTIGLAUCOMA'?'selected':''}>ANTIGLAUCOMA</option>
                <option value="VITRECTOMY" ${p.operation_type==='VITRECTOMY'?'selected':''}>VITRECTOMY</option>
                <option value="LASER" ${p.operation_type==='LASER'?'selected':''}>LASER</option>
            </select>
            <select id="pf-eye" class="border rounded px-3 py-2 text-sm">
                <option value="OD" ${p.eye==='OD'?'selected':''}>OD (правый)</option>
                <option value="OS" ${p.eye==='OS'?'selected':''}>OS (левый)</option>
                <option value="OU" ${p.eye==='OU'?'selected':''}>OU (оба)</option>
            </select>
            <select id="pf-district" class="border rounded px-3 py-2 text-sm">
                ${districts.map(d => ` + "`" + `<option value="${safe(d.id)}" ${p.district_id===d.id?'selected':''}>${safe(d.name)}</option>` + "`" + `).join('')}
            </select>
            <textarea id="pf-notes" placeholder="Заметки" class="border rounded px-3 py-2 text-sm col-span-2">${safe(p.notes, '')}</textarea>
        </div>
        <div class="mt-3 flex gap-2">
            <button onclick="${fn}(${p.id||0})" class="bg-blue-600 text-white px-4 py-1.5 rounded text-sm hover:bg-blue-700">Сохранить</button>
            <button onclick="document.getElementById('patient-form-area').innerHTML=''" class="bg-gray-200 px-4 py-1.5 rounded text-sm hover:bg-gray-300">Отмена</button>
        </div>
    </div>
    ` + "`" + `;
}

async function createPatient() {
    await api('/patients', {
        method: 'POST',
        body: JSON.stringify({
            first_name: document.getElementById('pf-fname').value,
            last_name: document.getElementById('pf-lname').value,
            middle_name: document.getElementById('pf-mname').value,
            phone: document.getElementById('pf-phone').value,
            email: document.getElementById('pf-email').value,
            date_of_birth: document.getElementById('pf-dob').value,
            diagnosis: document.getElementById('pf-diagnosis').value,
            operation_type: document.getElementById('pf-optype').value,
            eye: document.getElementById('pf-eye').value,
            district_id: parseInt(document.getElementById('pf-district').value),
            notes: document.getElementById('pf-notes').value
        })
    });
    await renderPatients();
}

async function updatePatient(id) {
    await api('/patients/' + id, {
        method: 'PATCH',
        body: JSON.stringify({
            phone: document.getElementById('pf-phone').value,
            email: document.getElementById('pf-email').value,
            diagnosis: document.getElementById('pf-diagnosis').value,
            notes: document.getElementById('pf-notes').value
        })
    });
    await renderPatients();
}

async function deletePatient(id) {
    if (!confirm('⚠️ Вы уверены, что хотите удалить этого пациента?\n\nЭто действие необратимо и удалит:\n- Карту пациента\n- Все чек-листы\n- Историю статусов\n- Связанные файлы\n\nПродолжить?')) return;

    try {
        await api('/patients/' + id, { method: 'DELETE' });
        alert('✅ Пациент успешно удалён');
        await renderPatients();
    } catch (err) {
        alert('❌ Ошибка удаления: ' + err.message);
    }
}

async function showPatientDetails(id) {
    try {
        const patientResponse = await api('/patients/' + id);
        const patient = patientResponse.data || patientResponse;
        const checklistResponse = await api('/checklists/patient/' + id).catch(() => []);

        // Обработка разных форматов ответа API
        let checklistItems = [];
        if (Array.isArray(checklistResponse)) {
            checklistItems = checklistResponse;
        } else if (checklistResponse && Array.isArray(checklistResponse.items)) {
            checklistItems = checklistResponse.items;
        } else if (checklistResponse && Array.isArray(checklistResponse.data)) {
            checklistItems = checklistResponse.data;
        }

        const modal = document.createElement('div');
        modal.id = 'patient-modal';
        modal.className = 'fixed inset-0 bg-black bg-opacity-60 backdrop-blur-sm flex items-center justify-center z-50 p-4 animate-fadeIn';
        modal.onclick = (e) => { if (e.target === modal) closePatientModal(); };

        const completedItems = checklistItems.filter(i => i.status === 'COMPLETED').length;
        const totalItems = checklistItems.length;
        const progress = totalItems > 0 ? Math.round((completedItems / totalItems) * 100) : 0;

        const surgeryDate = patient.surgery_date ? new Date(patient.surgery_date).toLocaleDateString('ru-RU') : 'Не назначена';
        const dob = patient.date_of_birth ? new Date(patient.date_of_birth).toLocaleDateString('ru-RU') : '—';

        modal.innerHTML = ` + "`" + `
        <div class="bg-white rounded-2xl shadow-2xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <div class="sticky top-0 bg-gradient-to-r from-blue-600 to-indigo-600 text-white px-6 py-4 flex justify-between items-center rounded-t-2xl">
                <h2 class="text-2xl font-bold">Карта пациента #${safe(patient.id)}</h2>
                <button onclick="closePatientModal()" class="text-white hover:text-gray-200 text-3xl leading-none">&times;</button>
            </div>

            <div class="p-6 space-y-6">
                <!-- Код доступа -->
                <div class="bg-gradient-to-r from-green-50 to-emerald-50 border-2 border-green-300 rounded-xl p-4">
                    <div class="flex items-center justify-between">
                        <div>
                            <div class="text-sm text-gray-600 mb-1">🔑 Код доступа пациента</div>
                            <div class="text-3xl font-mono font-bold text-green-700">${safe(patient.access_code, 'Не задан')}</div>
                            <div class="text-xs text-gray-500 mt-2">Для Telegram: /start ${safe(patient.access_code, 'Не задан')}</div>
                            <div class="text-xs text-gray-500">Для сайта: /patient?code=${safe(patient.access_code, 'Не задан')}</div>
                        </div>
                        <button onclick="copyAccessCode('${safe(patient.access_code, 'Не задан')}')" class="bg-green-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-green-700">
                            📋 Копировать
                        </button>
                    </div>
                </div>

                <!-- Основная информация -->
                <div class="grid md:grid-cols-2 gap-4">
                    <div class="bg-gray-50 rounded-lg p-4">
                        <h3 class="font-bold text-gray-700 mb-3 flex items-center">
                            <span class="text-blue-600 mr-2">👤</span> Личные данные
                        </h3>
                        <div class="space-y-2 text-sm">
                            <div><span class="text-gray-600">ФИО:</span> <span class="font-medium">${safe(patient.last_name)} ${safe(patient.first_name)} ${safe(patient.middle_name, '')}</span></div>
                            <div><span class="text-gray-600">Дата рождения:</span> <span class="font-medium">${dob}</span></div>
                            <div><span class="text-gray-600">Телефон:</span> <span class="font-medium">${safe(patient.phone)}</span></div>
                            <div><span class="text-gray-600">Email:</span> <span class="font-medium">${safe(patient.email)}</span></div>
                            <div><span class="text-gray-600">Адрес:</span> <span class="font-medium">${safe(patient.address)}</span></div>
                        </div>
                    </div>

                    <div class="bg-gray-50 rounded-lg p-4">
                        <h3 class="font-bold text-gray-700 mb-3 flex items-center">
                            <span class="text-red-600 mr-2">📋</span> Документы
                        </h3>
                        <div class="space-y-2 text-sm">
                            <div><span class="text-gray-600">СНИЛС:</span> <span class="font-medium">${safe(patient.snils)}</span></div>
                            <div><span class="text-gray-600">Паспорт:</span> <span class="font-medium">${safe(patient.passport_series, '')} ${safe(patient.passport_number)}</span></div>
                            <div><span class="text-gray-600">Полис ОМС:</span> <span class="font-medium">${safe(patient.policy_number)}</span></div>
                            <div><span class="text-gray-600">Район:</span> <span class="font-medium">${patient.district ? safe(patient.district.name) : '—'}</span></div>
                        </div>
                    </div>
                </div>

                <!-- Медицинская информация -->
                <div class="bg-blue-50 rounded-lg p-4">
                    <h3 class="font-bold text-gray-700 mb-3 flex items-center">
                        <span class="text-blue-600 mr-2">🏥</span> Медицинская информация
                    </h3>
                    <div class="grid md:grid-cols-2 gap-4 text-sm">
                        <div><span class="text-gray-600">Диагноз:</span> <span class="font-medium">${safe(patient.diagnosis)}</span></div>
                        <div><span class="text-gray-600">Тип операции:</span> <span class="font-medium">${safe(patient.operation_type)}</span></div>
                        <div><span class="text-gray-600">Глаз:</span> <span class="font-medium">${safe(patient.eye)}</span></div>
                        <div><span class="text-gray-600">Дата операции:</span> <span class="font-medium">${surgeryDate}</span></div>
                        <div class="md:col-span-2"><span class="text-gray-600">Заметки:</span> <span class="font-medium">${safe(patient.notes)}</span></div>
                    </div>
                </div>

                <!-- Статус -->
                <div class="bg-yellow-50 rounded-lg p-4">
                    <h3 class="font-bold text-gray-700 mb-3 flex items-center">
                        <span class="text-yellow-600 mr-2">📊</span> Статус подготовки
                    </h3>
                    <div class="flex items-center justify-between">
                        <div>
                            <div class="text-2xl font-bold text-gray-800">${safe(patient.status)}</div>
                            <div class="text-sm text-gray-600 mt-1">Прогресс чек-листа: ${completedItems}/${totalItems} (${progress}%)</div>
                        </div>
                        <div class="w-32 h-32">
                            <svg viewBox="0 0 36 36" class="circular-chart">
                                <path class="circle-bg" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#e5e7eb" stroke-width="3"/>
                                <path class="circle" stroke-dasharray="${progress}, 100" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#10b981" stroke-width="3"/>
                                <text x="18" y="20.35" class="percentage" text-anchor="middle" font-size="8" font-weight="bold" fill="#374151">${progress}%</text>
                            </svg>
                        </div>
                    </div>
                </div>

                <!-- Чек-лист -->
                ${checklistItems.length > 0 ? ` + "`" + `
                <div class="bg-gray-50 rounded-lg p-4">
                    <h3 class="font-bold text-gray-700 mb-3 flex items-center">
                        <span class="text-purple-600 mr-2">✓</span> Чек-лист подготовки
                    </h3>
                    <div class="space-y-2 max-h-64 overflow-y-auto">
                        ${checklistItems.map(item => {
                            const statusIcon = item.status === 'COMPLETED' ? '✅' : item.status === 'PENDING' ? '⏳' : '❌';
                            const statusColor = item.status === 'COMPLETED' ? 'text-green-600' : item.status === 'PENDING' ? 'text-yellow-600' : 'text-red-600';
                            return ` + "`" + `<div class="flex items-start gap-2 text-sm p-2 bg-white rounded">
                                <span class="text-lg">${statusIcon}</span>
                                <div class="flex-1">
                                    <div class="font-medium ${statusColor}">${safe(item.title)}</div>
                                    ${item.description ? ` + "`" + `<div class="text-xs text-gray-500">${safe(item.description, '')}</div>` + "`" + ` : ''}
                                </div>
                            </div>` + "`" + `;
                        }).join('')}
                    </div>
                </div>
                ` + "`" + ` : ''}

                <!-- Действия -->
                <div class="flex gap-3 pt-4 border-t">
                    <a href="/patient?code=${safe(patient.access_code, 'Не задан')}" target="_blank" class="flex-1 bg-blue-600 text-white text-center px-4 py-2 rounded-lg hover:bg-blue-700">
                        🔗 Открыть публичную страницу
                    </a>
                    <button onclick="closePatientModal()" class="flex-1 bg-gray-200 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-300">
                        Закрыть
                    </button>
                </div>
            </div>
        </div>
        ` + "`" + `;

        document.body.appendChild(modal);
    } catch (err) {
        alert('Ошибка загрузки данных пациента: ' + err.message);
    }
}

function closePatientModal() {
    const modal = document.getElementById('patient-modal');
    if (modal) modal.remove();
}

function copyAccessCode(code) {
    navigator.clipboard.writeText(code).then(() => {
        alert('Код доступа скопирован: ' + code);
    }).catch(() => {
        alert('Не удалось скопировать код');
    });
}

async function renderSurgeries(page = 1) {
    const surgeriesResponse = await api('/surgeries?page=' + page + '&limit=' + pageSize);
    const patientsResponse = await api('/patients?limit=100');
    const usersResponse = await api('/admin/users?limit=100');
    const surgeries = surgeriesResponse.data || surgeriesResponse;
    const meta = surgeriesResponse.meta || {};
    const patients = patientsResponse.data || patientsResponse;
    const users = usersResponse.data || usersResponse;
    let items = Array.isArray(surgeries) ? surgeries : (surgeries.surgeries || []);
    let patientsData = Array.isArray(patients) ? patients : (patients.patients || []);
    let html = ` + "`" + `
    <div class="mb-4">
        <button onclick="showCreateSurgery()" class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-700">+ Запланировать операцию</button>
    </div>
    <div id="surgery-form-area"></div>
    <div class="bg-white rounded-xl shadow overflow-hidden">
        <table class="w-full text-sm">
            <thead class="bg-gray-50">
                <tr>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">ID</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Пациент</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Хирург</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Дата</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Тип</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Глаз</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Статус</th>
                    <th class="text-left px-4 py-3 font-medium text-gray-600">Действия</th>
                </tr>
            </thead>
            <tbody>
    ` + "`" + `;
    if (items.length === 0) {
        html += '<tr><td colspan="8" class="px-4 py-8 text-center text-gray-400">Нет операций</td></tr>';
    }
    items.forEach(s => {
        const patient = s.patient ? (s.patient.last_name + ' ' + s.patient.first_name) : ('ID ' + s.patient_id);
        const surgeon = s.surgeon ? s.surgeon.name : ('ID ' + s.surgeon_id);
        const date = s.scheduled_date ? new Date(s.scheduled_date).toLocaleDateString('ru-RU') : '—';
        html += ` + "`" + `<tr class="border-t">
            <td class="px-4 py-3">${safe(s.id)}</td>
            <td class="px-4 py-3 font-medium">${patient}</td>
            <td class="px-4 py-3">${surgeon}</td>
            <td class="px-4 py-3">${date}</td>
            <td class="px-4 py-3 text-xs">${safe(s.operation_type)}</td>
            <td class="px-4 py-3">${safe(s.eye)}</td>
            <td class="px-4 py-3"><span class="px-2 py-0.5 rounded text-xs">${safe(s.status)}</span></td>
            <td class="px-4 py-3 space-x-2">
                <button onclick='editSurgery(${JSON.stringify(s).replace(/'/g,"&#39;")})' class="text-blue-600 hover:underline text-xs">Изменить</button>
                <button onclick="deleteSurgery(${safe(s.id)})" class="text-red-600 hover:underline text-xs">Удалить</button>
            </td>
        </tr>` + "`" + `;
    });
    html += '</tbody></table>';
    if (meta.total_pages > 1) {
        html += '<div class="px-4 py-3 border-t flex items-center justify-between">';
        html += '<div class="text-sm text-gray-600">Страница ' + page + ' из ' + meta.total_pages + '</div>';
        html += '<div class="flex gap-2">';
        if (page > 1) html += '<button onclick="renderSurgeries(' + (page-1) + ')" class="px-3 py-1 border rounded text-sm hover:bg-gray-50">Назад</button>';
        if (page < meta.total_pages) html += '<button onclick="renderSurgeries(' + (page+1) + ')" class="px-3 py-1 border rounded text-sm hover:bg-gray-50">Вперёд</button>';
        html += '</div></div>';
    }
    html += '</div>';
    document.getElementById('tab-content').innerHTML = html;
    window.allPatients = patientsData;
    window.allUsers = users;
}

function showCreateSurgery() {
    document.getElementById('surgery-form-area').innerHTML = surgeryForm({}, 'createSurgery');
}

function editSurgery(s) {
    document.getElementById('surgery-form-area').innerHTML = surgeryForm(s, 'updateSurgery');
}

function surgeryForm(s, fn) {
    const patients = window.allPatients || [];
    const users = window.allUsers || [];
    const surgeons = users.filter(u => u.role === 'SURGEON' || u.role === 'ADMIN');
    return ` + "`" + `
    <div class="bg-white rounded-xl shadow p-4 mb-4">
        <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <select id="sf-patient" class="border rounded px-3 py-2 text-sm">
                <option value="">Выберите пациента</option>
                ${patients.map(p => ` + "`" + `<option value="${safe(p.id)}" ${s.patient_id===p.id?'selected':''}>${p.last_name} ${p.first_name}</option>` + "`" + `).join('')}
            </select>
            <select id="sf-surgeon" class="border rounded px-3 py-2 text-sm">
                <option value="">Выберите хирурга</option>
                ${surgeons.map(u => ` + "`" + `<option value="${safe(u.id)}" ${s.surgeon_id===u.id?'selected':''}>${safe(u.name)}</option>` + "`" + `).join('')}
            </select>
            <input id="sf-date" type="date" value="${safe(s.scheduled_date, '')}" class="border rounded px-3 py-2 text-sm">
            <select id="sf-status" class="border rounded px-3 py-2 text-sm">
                <option value="SCHEDULED" ${s.status==='SCHEDULED'?'selected':''}>SCHEDULED</option>
                <option value="IN_PROGRESS" ${s.status==='IN_PROGRESS'?'selected':''}>IN_PROGRESS</option>
                <option value="COMPLETED" ${s.status==='COMPLETED'?'selected':''}>COMPLETED</option>
                <option value="CANCELLED" ${s.status==='CANCELLED'?'selected':''}>CANCELLED</option>
            </select>
            <textarea id="sf-notes" placeholder="Заметки" class="border rounded px-3 py-2 text-sm col-span-2">${safe(s.notes, '')}</textarea>
        </div>
        <div class="mt-3 flex gap-2">
            <button onclick="${fn}(${s.id||0})" class="bg-blue-600 text-white px-4 py-1.5 rounded text-sm hover:bg-blue-700">Сохранить</button>
            <button onclick="document.getElementById('surgery-form-area').innerHTML=''" class="bg-gray-200 px-4 py-1.5 rounded text-sm hover:bg-gray-300">Отмена</button>
        </div>
    </div>
    ` + "`" + `;
}

async function createSurgery() {
    await api('/surgeries', {
        method: 'POST',
        body: JSON.stringify({
            patient_id: parseInt(document.getElementById('sf-patient').value),
            scheduled_date: document.getElementById('sf-date').value,
            notes: document.getElementById('sf-notes').value
        })
    });
    await renderSurgeries();
}

async function updateSurgery(id) {
    await api('/surgeries/' + id, {
        method: 'PATCH',
        body: JSON.stringify({
            scheduled_date: document.getElementById('sf-date').value,
            status: document.getElementById('sf-status').value,
            notes: document.getElementById('sf-notes').value
        })
    });
    await renderSurgeries();
}

async function deleteSurgery(id) {
    if (!confirm('Удалить операцию?')) return;
    alert('Удаление операций пока не реализовано в API');
}

render();
</script>
</body>
</html>`
