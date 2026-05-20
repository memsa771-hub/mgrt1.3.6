(function () {
  const btn = document.getElementById("rdSettingsBtn");
  const menu = document.getElementById("rdSettingsMenu");
  const wrap = document.getElementById("rdSettingsWrap");
  if (!btn || !menu || !wrap) return;
  function setOpen(open) {
    menu.classList.toggle("hidden", !open);
    btn.setAttribute("aria-expanded", open ? "true" : "false");
  }
  btn.addEventListener("click", (e) => {
    e.stopPropagation();
    setOpen(menu.classList.contains("hidden"));
  });
  menu.addEventListener("click", (e) => e.stopPropagation());
  document.addEventListener("click", () => setOpen(false));
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") setOpen(false);
  });
})();
