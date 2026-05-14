document.documentElement.classList.add("js");

// Проверка авторизации через API (cookie отправится автоматически)
(function() {
  const publicPaths = ['/login'];
  const currentPath = window.location.pathname;
  
  if (!publicPaths.includes(currentPath)) {
    // Пробуем получить информацию о пользователе
    fetch('/api/auth/me', { credentials: 'include' })
      .then(res => {
        if (!res.ok) {
          sessionStorage.setItem("flowmodel:returnTo", currentPath);
          window.location.href = '/login';
        }
      })
      .catch(() => {
        sessionStorage.setItem("flowmodel:returnTo", currentPath);
        window.location.href = '/login';
      });
  }
})();

window.addEventListener("flowmodel:unauthorized", function () {
  if (window.location.pathname === "/login") return;
  
  sessionStorage.removeItem('flowmodel_role');
  window.location.href = '/login';
});