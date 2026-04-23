(function() {
  const loginForm = document.getElementById('login-form');
  const loginError = document.getElementById('login-error');

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
})();