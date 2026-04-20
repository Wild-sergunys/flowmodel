(function() {
  const loginForm = document.getElementById('login-form');
  const loginError = document.getElementById('login-error');
  const registerLink = document.getElementById('register-link');
  const registerDiv = document.getElementById('register-form');
  const registerForm = document.getElementById('register-form-element');
  const registerError = document.getElementById('register-error');

  // Переключение на регистрацию
  if (registerLink) {
    registerLink.addEventListener('click', (e) => {
      e.preventDefault();
      registerDiv.style.display = 'block';
    });
  }

  // Логин
  if (loginForm) {
    loginForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      loginError.textContent = '';
      
      const formData = new FormData(loginForm);
      const data = {
        login: formData.get('login'),
        password: formData.get('password')
      };

      try {
        const result = await FlowModelAPI.client.auth.login(data);
        
        sessionStorage.setItem('flowmodel_token', result.token);
        sessionStorage.setItem('flowmodel_role', result.role);
        
        const returnTo = sessionStorage.getItem('flowmodel:returnTo') || '/';
        sessionStorage.removeItem('flowmodel:returnTo');
        window.location.href = returnTo;
      } catch (e) {
        loginError.textContent = e.message;
      }
    });
  }

  // Регистрация
  if (registerForm) {
    registerForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      registerError.textContent = '';
      registerError.style.color = 'var(--pink)';
      
      const formData = new FormData(registerForm);
      const data = {
        login: formData.get('login'),
        password: formData.get('password')
      };

      try {
        await FlowModelAPI.client.auth.register(data);
        registerError.style.color = '#00c853';
        registerError.textContent = 'Регистрация успешна! Теперь войдите.';
        registerForm.reset();
        
        setTimeout(() => {
          registerDiv.style.display = 'none';
          registerError.textContent = '';
        }, 2000);
      } catch (e) {
        registerError.textContent = e.message;
      }
    });
  }
})();