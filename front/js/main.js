async function logout() {
    await fetch('/php/class/Auth/logout.php', { method: 'POST' });
    window.location.href = '/Auth/login.html';
}

// Exibe o nome do usuário se estiver autenticado (não redireciona se não estiver)
async function loadUserInfo() {
    try {
        const res  = await fetch('/php/class/session_check.php');
        const data = await res.json();
        const el   = document.getElementById('nome-usuario');
        if (el) el.textContent = data.autenticado ? data.usuario.usuario : '';
    } catch {
        // página pública — ignora silenciosamente
    }
}

document.addEventListener('DOMContentLoaded', loadUserInfo);
