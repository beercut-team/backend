package server

const patientPublicHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Статус подготовки к операции</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        .status-indicator { width: 12px; height: 12px; border-radius: 50%; display: inline-block; }
        .status-green { background: #10b981; }
        .status-yellow { background: #f59e0b; }
        .status-red { background: #ef4444; }
    </style>
</head>
<body class="bg-gradient-to-br from-blue-50 to-indigo-100 min-h-screen">

<div id="app" class="container mx-auto px-4 py-8 max-w-2xl"></div>

<script>
const API = '/api/v1';
let accessCode = '';

function render() {
    const urlParams = new URLSearchParams(window.location.search);
    accessCode = urlParams.get('code') || '';

    if (!accessCode) {
        renderCodeInput();
    } else {
        loadPatientStatus();
    }
}

function renderCodeInput() {
    document.getElementById('app').innerHTML = '<div class="bg-white rounded-2xl shadow-xl p-8 text-center">' +
        '<div class="mb-6"><svg class="w-20 h-20 mx-auto text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">' +
        '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>' +
        '</svg></div>' +
        '<h1 class="text-3xl font-bold text-gray-800 mb-2">Проверка статуса</h1>' +
        '<p class="text-gray-600 mb-8">Введите ваш код доступа для просмотра статуса подготовки к операции</p>' +
        '<div id="error" class="hidden bg-red-50 text-red-600 rounded-lg p-3 mb-4 text-sm"></div>' +
        '<form id="code-form" class="space-y-4">' +
        '<input id="code-input" type="text" placeholder="Введите код доступа" class="w-full border-2 border-gray-300 rounded-lg px-4 py-3 text-center text-lg font-mono focus:outline-none focus:border-blue-500" required>' +
        '<button type="submit" class="w-full bg-blue-600 text-white rounded-lg py-3 font-medium hover:bg-blue-700 transition">Проверить статус</button>' +
        '</form>' +
        '<p class="text-sm text-gray-500 mt-6">Код доступа выдаётся вашим лечащим врачом</p>' +
        '</div>';

    document.getElementById('code-form').onsubmit = function(e) {
        e.preventDefault();
        const code = document.getElementById('code-input').value.trim();
        if (code) {
            window.location.href = '?code=' + encodeURIComponent(code);
        }
    };
}

async function loadPatientStatus() {
    try {
        const response = await fetch(API + '/patients/public/' + accessCode);
        if (!response.ok) throw new Error('Неверный код доступа');
        const data = await response.json();
        renderPatientStatus(data.data);
    } catch (err) {
        renderError(err.message);
    }
}

function renderError(message) {
    document.getElementById('app').innerHTML = '<div class="bg-white rounded-2xl shadow-xl p-8 text-center">' +
        '<div class="mb-4"><svg class="w-16 h-16 mx-auto text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">' +
        '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>' +
        '</svg></div>' +
        '<h2 class="text-2xl font-bold text-gray-800 mb-2">Ошибка</h2>' +
        '<p class="text-gray-600 mb-6">' + message + '</p>' +
        '<button onclick="window.location.href=\'/patient\'" class="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700">Попробовать снова</button>' +
        '</div>';
}

function renderPatientStatus(patient) {
    const statusInfo = {
        'NEW': { text: 'Новый', color: 'yellow', icon: '📝' },
        'PREPARATION': { text: 'Идёт подготовка', color: 'yellow', icon: '⏳' },
        'REVIEW_NEEDED': { text: 'На проверке у хирурга', color: 'yellow', icon: '👨‍⚕️' },
        'APPROVED': { text: 'Готов к операции', color: 'green', icon: '✅' },
        'REJECTED': { text: 'Требуется дополнительная подготовка', color: 'red', icon: '❌' },
        'SCHEDULED': { text: 'Операция запланирована', color: 'green', icon: '📅' }
    };

    const status = statusInfo[patient.status] || { text: patient.status, color: 'yellow', icon: '📋' };
    const bgColor = status.color === 'green' ? 'green' : status.color === 'red' ? 'red' : 'yellow';

    let html = '<div class="bg-white rounded-2xl shadow-xl overflow-hidden">' +
        '<div class="bg-gradient-to-r from-blue-600 to-indigo-600 p-6 text-white">' +
        '<h1 class="text-2xl font-bold mb-2">' + patient.first_name + ' ' + patient.last_name + '</h1>' +
        '<p class="opacity-90">Код доступа: ' + accessCode + '</p>' +
        '</div>' +
        '<div class="p-6">' +
        '<div class="mb-6 p-4 rounded-xl bg-' + bgColor + '-50 border-l-4 border-' + bgColor + '-500">' +
        '<div class="flex items-center">' +
        '<span class="text-3xl mr-3">' + status.icon + '</span>' +
        '<div><div class="font-semibold text-gray-800">Текущий статус</div>' +
        '<div class="text-lg font-bold text-' + bgColor + '-700">' + status.text + '</div></div>' +
        '</div></div>' +
        '<div class="space-y-4 mb-6">' +
        '<div class="flex justify-between py-2 border-b"><span class="text-gray-600">Операция:</span><span class="font-medium">' + patient.operation_type + '</span></div>' +
        '<div class="flex justify-between py-2 border-b"><span class="text-gray-600">Глаз:</span><span class="font-medium">' + patient.eye + '</span></div>';

    if (patient.surgery_date) {
        html += '<div class="flex justify-between py-2 border-b"><span class="text-gray-600">Дата операции:</span>' +
            '<span class="font-medium text-blue-600">' + new Date(patient.surgery_date).toLocaleDateString('ru-RU') + '</span></div>';
    }

    html += '</div>';

    if (patient.checklist_progress) {
        html += '<div class="mb-6"><h3 class="font-semibold text-gray-800 mb-3">Прогресс подготовки</h3>' +
            '<div class="bg-gray-200 rounded-full h-4 overflow-hidden">' +
            '<div class="bg-blue-600 h-full transition-all" style="width: ' + patient.checklist_progress + '%"></div></div>' +
            '<p class="text-sm text-gray-600 mt-2 text-center">' + patient.checklist_progress + '% выполнено</p></div>';
    }

    html += '<div class="bg-blue-50 rounded-lg p-4 text-sm text-gray-700">' +
        '<p class="font-medium mb-2">💡 Что дальше?</p>';

    if (status.color === 'yellow') {
        html += '<p>Ваш врач работает над подготовкой документов. Вы получите уведомление о любых изменениях.</p>';
    } else if (status.color === 'green' && !patient.surgery_date) {
        html += '<p>Вы готовы к операции! Ожидайте назначения даты.</p>';
    } else if (status.color === 'green' && patient.surgery_date) {
        html += '<p>Операция запланирована. Следуйте рекомендациям вашего врача.</p>';
    } else if (status.color === 'red') {
        html += '<p>Обратитесь к вашему лечащему врачу для уточнения деталей.</p>';
    }

    html += '</div></div></div>' +
        '<div class="text-center mt-6">' +
        '<button onclick="window.location.href=\'/patient\'" class="text-blue-600 hover:underline text-sm">Проверить другой код</button>' +
        '</div>';

    document.getElementById('app').innerHTML = html;
}

render();
</script>

</body>
</html>
`
