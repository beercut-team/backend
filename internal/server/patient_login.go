package server

const patientLoginHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Вход для пациента</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gradient-to-br from-blue-50 to-indigo-100 min-h-screen">

<div class="container mx-auto px-4 py-8 max-w-md">
    <div class="bg-white rounded-2xl shadow-xl p-8">
        <div class="text-center mb-8">
            <div class="mb-4">
                <svg class="w-20 h-20 mx-auto text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                </svg>
            </div>
            <h1 class="text-3xl font-bold text-gray-800 mb-2">Вход для пациента</h1>
            <p class="text-gray-600">Введите ваш код доступа</p>
        </div>

        <div id="error" class="hidden bg-red-50 text-red-600 rounded-lg p-3 mb-4 text-sm"></div>

        <form id="login-form" class="space-y-4">
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">Код доступа</label>
                <input
                    id="access-code"
                    type="text"
                    placeholder="Введите код доступа"
                    class="w-full border-2 border-gray-300 rounded-lg px-4 py-3 text-center text-lg font-mono focus:outline-none focus:border-blue-500"
                    required
                    autocomplete="off">
            </div>

            <button
                type="submit"
                id="submit-btn"
                class="w-full bg-blue-600 text-white rounded-lg py-3 font-medium hover:bg-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed">
                Войти
            </button>
        </form>

        <div class="mt-6 text-center">
            <p class="text-sm text-gray-500">Код доступа выдаётся вашим лечащим врачом</p>
            <a href="/patient" class="text-sm text-blue-600 hover:underline mt-2 inline-block">Проверить статус без входа</a>
        </div>
    </div>
</div>

<script>
const API = '/api/v1';

document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();

    const accessCode = document.getElementById('access-code').value.trim();
    const submitBtn = document.getElementById('submit-btn');
    const errorDiv = document.getElementById('error');

    if (!accessCode) return;

    // Disable button and show loading
    submitBtn.disabled = true;
    submitBtn.textContent = 'Вход...';
    errorDiv.classList.add('hidden');

    try {
        const response = await fetch(API + '/auth/patient-login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ access_code: accessCode })
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Неверный код доступа');
        }

        const data = await response.json();

        // Store tokens
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
        localStorage.setItem('user', JSON.stringify(data.user));

        // Redirect to patient portal
        window.location.href = '/patient/portal';

    } catch (err) {
        errorDiv.textContent = err.message;
        errorDiv.classList.remove('hidden');
        submitBtn.disabled = false;
        submitBtn.textContent = 'Войти';
    }
});
</script>

</body>
</html>
`
