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
        const data = await response.json();
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
            const data = await api('/auth/login', {
                method: 'POST',
                body: JSON.stringify({
                    email: document.getElementById('email').value,
                    password: document.getElementById('password').value
                })
            });
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
    const stats = await api('/admin/stats');
    document.getElementById('tab-content').innerHTML = ` + "`" + `
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="bg-white rounded-xl shadow p-6 text-center">
            <div class="text-3xl font-bold text-blue-600">${stats.users}</div>
            <div class="text-gray-500 mt-1">Пользователи</div>
        </div>
        <div class="bg-white rounded-xl shadow p-6 text-center">
            <div class="text-3xl font-bold text-green-600">${stats.patients}</div>
            <div class="text-gray-500 mt-1">Пациенты</div>
        </div>
        <div class="bg-white rounded-xl shadow p-6 text-center">
            <div class="text-3xl font-bold text-purple-600">${stats.districts}</div>
            <div class="text-gray-500 mt-1">Районы</div>
        </div>
        <div class="bg-white rounded-xl shadow p-6 text-center">
            <div class="text-3xl font-bold text-orange-600">${stats.surgeries}</div>
            <div class="text-gray-500 mt-1">Операции</div>
        </div>
    </div>
    ` + "`" + `;
}

async function renderDistricts() {
    const response = await api('/districts');
    const districts = response.data || response;
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
                <td class="px-4 py-3">${d.id}</td>
                <td class="px-4 py-3 font-medium">${d.name}</td>
                <td class="px-4 py-3">${d.region}</td>
                <td class="px-4 py-3"><span class="bg-gray-100 px-2 py-0.5 rounded text-xs">${d.code}</span></td>
                <td class="px-4 py-3">${d.timezone}</td>
                <td class="px-4 py-3 space-x-2">
                    <button onclick='editDistrict(${JSON.stringify(d).replace(/'/g,"&#39;")})' class="text-blue-600 hover:underline text-xs">Изменить</button>
                    <button onclick="deleteDistrict(${d.id})" class="text-red-600 hover:underline text-xs">Удалить</button>
                </td>
            </tr>` + "`" + `;
        });
    }
    html += '</tbody></table></div>';
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
            <input id="df-name" placeholder="Название" value="${d.name||''}" class="border rounded px-3 py-2 text-sm">
            <input id="df-region" placeholder="Регион" value="${d.region||''}" class="border rounded px-3 py-2 text-sm">
            <input id="df-code" placeholder="Код" value="${d.code||''}" class="border rounded px-3 py-2 text-sm">
            <input id="df-tz" placeholder="Часовой пояс" value="${d.timezone||''}" class="border rounded px-3 py-2 text-sm">
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

async function renderUsers() {
    const usersResponse = await api('/admin/users');
    const users = usersResponse.data || usersResponse;
    const districtsResponse = await api('/districts');
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
    const roleBadge = { ADMIN: 'bg-red-100 text-red-700', SURGEON: 'bg-blue-100 text-blue-700', DISTRICT_DOCTOR: 'bg-green-100 text-green-700', PATIENT: 'bg-gray-100 text-gray-700' };
    if (Array.isArray(users)) {
        users.forEach(u => {
            const dist = districts.find(d => d.id === u.district_id);
            html += ` + "`" + `<tr class="border-t">
                <td class="px-4 py-3">${u.id}</td>
                <td class="px-4 py-3 font-medium">${u.name}</td>
                <td class="px-4 py-3">${u.email}</td>
                <td class="px-4 py-3">${u.phone || '—'}</td>
                <td class="px-4 py-3"><span class="px-2 py-0.5 rounded text-xs ${roleBadge[u.role]||'bg-gray-100'}">${u.role}</span></td>
                <td class="px-4 py-3">${dist ? dist.name : '—'}</td>
                <td class="px-4 py-3">${u.is_active ? '<span class="text-green-600">✓</span>' : '<span class="text-red-600">✗</span>'}</td>
                <td class="px-4 py-3 space-x-2">
                    <button onclick='editUser(${JSON.stringify(u).replace(/'/g,"&#39;")})' class="text-blue-600 hover:underline text-xs">Изменить</button>
                    <button onclick="deleteUser(${u.id})" class="text-red-600 hover:underline text-xs">Удалить</button>
                </td>
            </tr>` + "`" + `;
        });
    }
    html += '</tbody></table></div>';
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
            <input id="uf-email" placeholder="Email" value="${u.email||''}" class="border rounded px-3 py-2 text-sm" ${u.id?'disabled':''}>
            <input id="uf-name" placeholder="Полное имя" value="${u.name||''}" class="border rounded px-3 py-2 text-sm">
            <input id="uf-fname" placeholder="Имя" value="${u.first_name||''}" class="border rounded px-3 py-2 text-sm">
            <input id="uf-lname" placeholder="Фамилия" value="${u.last_name||''}" class="border rounded px-3 py-2 text-sm">
            <input id="uf-mname" placeholder="Отчество" value="${u.middle_name||''}" class="border rounded px-3 py-2 text-sm">
            <input id="uf-phone" placeholder="Телефон" value="${u.phone||''}" class="border rounded px-3 py-2 text-sm">
            <select id="uf-role" class="border rounded px-3 py-2 text-sm">
                <option value="PATIENT" ${u.role==='PATIENT'?'selected':''}>PATIENT</option>
                <option value="DISTRICT_DOCTOR" ${u.role==='DISTRICT_DOCTOR'?'selected':''}>DISTRICT_DOCTOR</option>
                <option value="SURGEON" ${u.role==='SURGEON'?'selected':''}>SURGEON</option>
                <option value="ADMIN" ${u.role==='ADMIN'?'selected':''}>ADMIN</option>
            </select>
            <select id="uf-district" class="border rounded px-3 py-2 text-sm">
                <option value="">Без района</option>
                ${districts.map(d => ` + "`" + `<option value="${d.id}" ${u.district_id===d.id?'selected':''}>${d.name}</option>` + "`" + `).join('')}
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

async function renderPatients() {
    const patientsResponse = await api('/patients');
    const patients = patientsResponse.data || patientsResponse;
    const districtsResponse = await api('/districts');
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
        html += ` + "`" + `<tr class="border-t">
            <td class="px-4 py-3">${p.id}</td>
            <td class="px-4 py-3 font-medium">${p.last_name} ${p.first_name} ${p.middle_name||''}</td>
            <td class="px-4 py-3">${p.phone || '—'}</td>
            <td class="px-4 py-3 max-w-xs truncate">${p.diagnosis || '—'}</td>
            <td class="px-4 py-3"><span class="text-xs">${p.operation_type || '—'}</span></td>
            <td class="px-4 py-3">${p.eye || '—'}</td>
            <td class="px-4 py-3"><span class="px-2 py-0.5 rounded text-xs ${statusBadge[p.status]||'bg-gray-100'}">${p.status}</span></td>
            <td class="px-4 py-3 space-x-2">
                <button onclick='editPatient(${JSON.stringify(p).replace(/'/g,"&#39;")})' class="text-blue-600 hover:underline text-xs">Изменить</button>
                <button onclick="deletePatient(${p.id})" class="text-red-600 hover:underline text-xs">Удалить</button>
            </td>
        </tr>` + "`" + `;
    });
    html += '</tbody></table></div>';
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
            <input id="pf-fname" placeholder="Имя" value="${p.first_name||''}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-lname" placeholder="Фамилия" value="${p.last_name||''}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-mname" placeholder="Отчество" value="${p.middle_name||''}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-phone" placeholder="Телефон" value="${p.phone||''}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-email" placeholder="Email" value="${p.email||''}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-dob" type="date" placeholder="Дата рождения" value="${p.date_of_birth||''}" class="border rounded px-3 py-2 text-sm">
            <input id="pf-diagnosis" placeholder="Диагноз" value="${p.diagnosis||''}" class="border rounded px-3 py-2 text-sm col-span-2">
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
                ${districts.map(d => ` + "`" + `<option value="${d.id}" ${p.district_id===d.id?'selected':''}>${d.name}</option>` + "`" + `).join('')}
            </select>
            <textarea id="pf-notes" placeholder="Заметки" class="border rounded px-3 py-2 text-sm col-span-2">${p.notes||''}</textarea>
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
    if (!confirm('Удалить пациента?')) return;
    alert('Удаление пациентов пока не реализовано в API');
}

async function renderSurgeries() {
    const surgeriesResponse = await api('/surgeries');
    const patientsResponse = await api('/patients');
    const usersResponse = await api('/admin/users');
    const surgeries = surgeriesResponse.data || surgeriesResponse;
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
            <td class="px-4 py-3">${s.id}</td>
            <td class="px-4 py-3 font-medium">${patient}</td>
            <td class="px-4 py-3">${surgeon}</td>
            <td class="px-4 py-3">${date}</td>
            <td class="px-4 py-3 text-xs">${s.operation_type || '—'}</td>
            <td class="px-4 py-3">${s.eye || '—'}</td>
            <td class="px-4 py-3"><span class="px-2 py-0.5 rounded text-xs">${s.status}</span></td>
            <td class="px-4 py-3 space-x-2">
                <button onclick='editSurgery(${JSON.stringify(s).replace(/'/g,"&#39;")})' class="text-blue-600 hover:underline text-xs">Изменить</button>
                <button onclick="deleteSurgery(${s.id})" class="text-red-600 hover:underline text-xs">Удалить</button>
            </td>
        </tr>` + "`" + `;
    });
    html += '</tbody></table></div>';
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
                ${patients.map(p => ` + "`" + `<option value="${p.id}" ${s.patient_id===p.id?'selected':''}>${p.last_name} ${p.first_name}</option>` + "`" + `).join('')}
            </select>
            <select id="sf-surgeon" class="border rounded px-3 py-2 text-sm">
                <option value="">Выберите хирурга</option>
                ${surgeons.map(u => ` + "`" + `<option value="${u.id}" ${s.surgeon_id===u.id?'selected':''}>${u.name}</option>` + "`" + `).join('')}
            </select>
            <input id="sf-date" type="date" value="${s.scheduled_date||''}" class="border rounded px-3 py-2 text-sm">
            <select id="sf-status" class="border rounded px-3 py-2 text-sm">
                <option value="SCHEDULED" ${s.status==='SCHEDULED'?'selected':''}>SCHEDULED</option>
                <option value="IN_PROGRESS" ${s.status==='IN_PROGRESS'?'selected':''}>IN_PROGRESS</option>
                <option value="COMPLETED" ${s.status==='COMPLETED'?'selected':''}>COMPLETED</option>
                <option value="CANCELLED" ${s.status==='CANCELLED'?'selected':''}>CANCELLED</option>
            </select>
            <textarea id="sf-notes" placeholder="Заметки" class="border rounded px-3 py-2 text-sm col-span-2">${s.notes||''}</textarea>
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
