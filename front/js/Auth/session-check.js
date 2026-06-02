// Esconde a página imediatamente para evitar flash de conteúdo protegido
document.documentElement.style.visibility = 'hidden';

(async () => {
    try {
        const res  = await fetch('/php/class/session_check.php');
        const data = await res.json();

        if (!data.autenticado) {
            window.location.replace('/Auth/login.html');
            return;
        }

        // Disponibiliza os dados do usuário globalmente para a página
        window.__session = data.usuario;
        document.documentElement.style.visibility = '';
    } catch {
        window.location.replace('/Auth/login.html');
    }
})();
