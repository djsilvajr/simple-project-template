document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('form-login');
    const erro = document.getElementById('erro');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        erro.textContent = '';

        const payload = {
            email: document.getElementById('email').value.trim(),
            senha: document.getElementById('senha').value,
        };

        try {
            const res  = await fetch('/php/class/Auth/auth.php', {
                method:  'POST',
                headers: { 'Content-Type': 'application/json' },
                body:    JSON.stringify(payload),
            });

            const data = await res.json();

            if (data.status) {
                window.location.href = '/';
            } else {
                erro.textContent = data.message || 'Credenciais inválidas.';
            }
        } catch {
            erro.textContent = 'Erro de conexão. Tente novamente.';
        }
    });
});
