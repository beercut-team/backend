package server

const patientPortalHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Личный кабинет пациента</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gradient-to-br from-blue-50 to-indigo-100 min-h-screen">

<div id="app" class="container mx-auto px-4 py-8 max-w-4xl"></div>

<script>
const API = '/api/v1';
let accessToken = localStorage.getItem('access_token');
let user = null;

// Check for token in URL (from Telegram bot)
const urlParams = new URLSearchParams(window.location.search);
const telegramToken = urlParams.get('token');

if (telegramToken) {
    // Exchange Telegram token for JWT
    loginWithTelegramToken(telegramToken);
} else {
    // Check authentication
    if (!accessToken) {
        window.location.href = '/patient/login';
    }

    try {
        user = JSON.parse(localStorage.getItem('user'));
    } catch (e) {
        window.location.href = '/patient/login';
    }

    loadPatientData();
}

async function loginWithTelegramToken(token) {
    try {
        const response = await fetch(API + '/auth/telegram-token-login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ token: token })
        });

        if (!response.ok) {
            throw new Error('Недействительный или истёкший токен');
        }

        const data = await response.json();

        // Store tokens
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
        localStorage.setItem('user', JSON.stringify(data.user));

        // Remove token from URL and reload
        window.history.replaceState({}, document.title, '/patient/portal');
        window.location.reload();

    } catch (err) {
        alert('Ошибка входа: ' + err.message);
        window.location.href = '/patient/login';
    }
}

async function fetchWithAuth(url, options = {}) {
    const headers = {
        'Authorization': 'Bearer ' + accessToken,
        'Content-Type': 'application/json',
        ...options.headers
    };

    const response = await fetch(url, { ...options, headers });

    if (response.status === 401) {
        // Try to refresh token
        const refreshed = await refreshToken();
        if (refreshed) {
            headers.Authorization = 'Bearer ' + accessToken;
            return fetch(url, { ...options, headers });
        } else {
            logout();
            return response;
        }
    }

    return response;
}

async function refreshToken() {
    try {
        const refreshToken = localStorage.getItem('refresh_token');
        const response = await fetch(API + '/auth/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken })
        });

        if (!response.ok) return false;

        const data = await response.json();
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
        accessToken = data.access_token;
        return true;
    } catch (e) {
        return false;
    }
}

function logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
    window.location.href = '/patient/login';
}

async function loadPatientData() {
    try {
        const response = await fetchWithAuth(API + '/patients/' + user.id);
        if (!response.ok) throw new Error('Не удалось загрузить данные');

        const result = await response.json();
        const patient = result.data || result;

        renderPortal(patient);
    } catch (err) {
        renderError(err.message);
    }
}

function renderError(message) {
    document.getElementById('app').innerHTML =
        '<div class="bg-white rounded-2xl shadow-xl p-8 text-center">' +
        '<div class="mb-4"><svg class="w-16 h-16 mx-auto text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">' +
        '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>' +
        '</svg></div>' +
        '<h2 class="text-2xl font-bold text-gray-800 mb-2">Ошибка</h2>' +
        '<p class="text-gray-600 mb-6">' + message + '</p>' +
        '<button onclick="logout()" class="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700">Выйти</button>' +
        '</div>';
}

function renderPortal(patient) {
    const statusInfo = {
        'NEW': { text: 'Новый', color: 'yellow', icon: '📝', bg: 'yellow' },
        'PREPARATION': { text: 'Идёт подготовка', color: 'yellow', icon: '⏳', bg: 'yellow' },
        'REVIEW_NEEDED': { text: 'На проверке у хирурга', color: 'blue', icon: '👨‍⚕️', bg: 'blue' },
        'APPROVED': { text: 'Готов к операции', color: 'green', icon: '✅', bg: 'green' },
        'REJECTED': { text: 'Требуется дополнительная подготовка', color: 'red', icon: '❌', bg: 'red' },
        'SCHEDULED': { text: 'Операция запланирована', color: 'green', icon: '📅', bg: 'green' }
    };

    const status = statusInfo[patient.status] || { text: patient.status, color: 'gray', icon: '📋', bg: 'gray' };

    let html =
        '<div class="mb-6 flex justify-between items-center">' +
        '<h1 class="text-3xl font-bold text-gray-800">Личный кабинет</h1>' +
        '<button onclick="logout()" class="text-red-600 hover:text-red-700 font-medium">Выйти</button>' +
        '</div>' +

        '<div class="bg-white rounded-2xl shadow-xl overflow-hidden mb-6">' +
        '<div class="bg-gradient-to-r from-blue-600 to-indigo-600 p-6 text-white">' +
        '<h2 class="text-2xl font-bold mb-1">' + (patient.first_name || '') + ' ' + (patient.last_name || '') + '</h2>' +
        '<p class="opacity-90">' + (patient.email || '') + '</p>' +
        '</div>' +

        '<div class="p-6">' +
        '<div class="mb-6 p-4 rounded-xl bg-' + status.bg + '-50 border-l-4 border-' + status.bg + '-500">' +
        '<div class="flex items-center">' +
        '<span class="text-3xl mr-3">' + status.icon + '</span>' +
        '<div><div class="font-semibold text-gray-800">Текущий статус</div>' +
        '<div class="text-lg font-bold text-' + status.bg + '-700">' + status.text + '</div></div>' +
        '</div></div>' +

        '<div class="grid md:grid-cols-2 gap-4 mb-6">' +
        '<div class="bg-gray-50 rounded-lg p-4">' +
        '<div class="text-sm text-gray-600 mb-1">Операция</div>' +
        '<div class="font-semibold text-gray-800">' + (patient.operation_type || '—') + '</div>' +
        '</div>' +
        '<div class="bg-gray-50 rounded-lg p-4">' +
        '<div class="text-sm text-gray-600 mb-1">Глаз</div>' +
        '<div class="font-semibold text-gray-800">' + (patient.eye || '—') + '</div>' +
        '</div>';

    if (patient.surgery_date) {
        html +=
            '<div class="bg-blue-50 rounded-lg p-4 md:col-span-2">' +
            '<div class="text-sm text-gray-600 mb-1">Дата операции</div>' +
            '<div class="font-semibold text-blue-700 text-lg">' + new Date(patient.surgery_date).toLocaleDateString('ru-RU', { year: 'numeric', month: 'long', day: 'numeric' }) + '</div>' +
            '</div>';
    }

    html += '</div>';

    // Checklist progress
    if (patient.checklist_progress !== undefined) {
        html +=
            '<div class="mb-6">' +
            '<h3 class="font-semibold text-gray-800 mb-3 flex items-center">' +
            '<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">' +
            '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>' +
            '</svg>Прогресс подготовки</h3>' +
            '<div class="bg-gray-200 rounded-full h-4 overflow-hidden">' +
            '<div class="bg-blue-600 h-full transition-all duration-500" style="width: ' + patient.checklist_progress + '%"></div></div>' +
            '<p class="text-sm text-gray-600 mt-2 text-center font-medium">' + patient.checklist_progress + '% выполнено</p>' +
            '</div>';
    }

    // Info box
    html +=
        '<div class="bg-blue-50 rounded-lg p-4 text-sm text-gray-700">' +
        '<p class="font-medium mb-2 flex items-center">' +
        '<svg class="w-5 h-5 mr-2 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">' +
        '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>' +
        '</svg>Что дальше?</p>';

    if (status.bg === 'yellow') {
        html += '<p>Ваш врач работает над подготовкой документов. Вы получите уведомление о любых изменениях.</p>';
    } else if (status.bg === 'green' && !patient.surgery_date) {
        html += '<p>Вы готовы к операции! Ожидайте назначения даты.</p>';
    } else if (status.bg === 'green' && patient.surgery_date) {
        html += '<p>Операция запланирована. Следуйте рекомендациям вашего врача.</p>';
    } else if (status.bg === 'red') {
        html += '<p>Обратитесь к вашему лечащему врачу для уточнения деталей.</p>';
    } else if (status.bg === 'blue') {
        html += '<p>Хирург проверяет ваши документы. Это может занять некоторое время.</p>';
    }

    html += '</div></div></div>';

    // Contact info
    html +=
        '<div class="bg-white rounded-2xl shadow-xl p-6">' +
        '<h3 class="font-semibold text-gray-800 mb-4 flex items-center">' +
        '<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">' +
        '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>' +
        '</svg>Контактная информация</h3>' +
        '<div class="space-y-3 text-sm">';

    if (patient.phone) {
        html += '<div class="flex items-center text-gray-700"><svg class="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"></path></svg>' + patient.phone + '</div>';
    }

    if (patient.email) {
        html += '<div class="flex items-center text-gray-700"><svg class="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path></svg>' + patient.email + '</div>';
    }

    html += '</div></div>';

    document.getElementById('app').innerHTML = html;
}

// Load data on page load
loadPatientData();
</script>

</body>
</html>
`;
