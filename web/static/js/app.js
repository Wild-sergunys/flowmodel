document.documentElement.classList.add("js");

(function() {
  const publicPaths = ['/login'];
  const currentPath = window.location.pathname;
  
  if (!publicPaths.includes(currentPath)) {
    const token = sessionStorage.getItem('flowmodel_token');
    if (!token) {
      sessionStorage.setItem("flowmodel:returnTo", currentPath);
      window.location.href = '/login';
    }
  }
})();

window.addEventListener("flowmodel:unauthorized", function () {
  if (window.location.pathname === "/login") return;
  
  sessionStorage.clear();
  window.location.href = '/login';
});