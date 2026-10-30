async function logout() {
    await fetch('/php/class/Auth/logout.php', { method: 'POST' });
    window.location.href = '/Auth/login.html';
}

// Mostra os elementos conforme a sessão (não redireciona se não estiver logado):
//   data-auth="user"  → só para usuários logados
//   data-auth="guest" → só para visitantes
function applyAuthState(autenticado) {
    document.querySelectorAll('[data-auth]').forEach((el) => {
        el.hidden = el.dataset.auth !== (autenticado ? 'user' : 'guest');
    });
}

async function loadUserInfo() {
    try {
        const res  = await fetch('/php/class/session_check.php');
        const data = await res.json();

        const el = document.getElementById('nome-usuario');
        if (el) el.textContent = data.autenticado ? data.usuario.name : '';

        applyAuthState(data.autenticado);
    } catch {
        applyAuthState(false); // falha na checagem → trata como visitante
    }
}

document.addEventListener('DOMContentLoaded', loadUserInfo);
