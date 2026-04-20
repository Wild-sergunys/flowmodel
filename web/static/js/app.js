document.documentElement.classList.add("js");

window.addEventListener("flowmodel:unauthorized", function () {
  if (window.location.pathname !== "/login") {
    window.sessionStorage.setItem("flowmodel:returnTo", window.location.pathname + window.location.search);
    window.location.assign("/login");
  }
});
