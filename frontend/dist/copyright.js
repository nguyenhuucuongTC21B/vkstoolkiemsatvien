// copyright.js — xử lý mở/đóng popup Bản quyền.
// Popup hiển thị avatarC.jpg + thông tin coder Nguyễn Hữu Cường.

(function() {
  const modal = document.getElementById("copyright-modal");
  const btnOpen = document.getElementById("btn-copyright");
  const btnClose = document.getElementById("btn-close-copyright");
  const backdrop = modal?.querySelector(".modal-backdrop");

  function openModal() {
    if (!modal) return;
    modal.hidden = false;
    modal.classList.add("modal-open");
    document.body.style.overflow = "hidden";
  }
  function closeModal() {
    if (!modal) return;
    modal.classList.remove("modal-open");
    modal.hidden = true;
    document.body.style.overflow = "";
  }

  if (btnOpen) btnOpen.addEventListener("click", openModal);
  if (btnClose) btnClose.addEventListener("click", closeModal);
  if (backdrop) backdrop.addEventListener("click", closeModal);

  // Đóng bằng ESC
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && modal && !modal.hidden) closeModal();
  });
})();
