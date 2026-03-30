// frontend/ui.js
async function updateAuthButton() {
    const response = await fetch("http://localhost:8080/api/auth/me", {
        credentials: "include"
    });

    const signInBtn = document.getElementById("signInBtn");

    if (response.ok) {
        signInBtn.textContent = "Sign Out";
        signInBtn.href = "#";
        signInBtn.addEventListener("click", async (e) => {
            e.preventDefault();
            await fetch("http://localhost:8080/api/auth/logout", {
                method: "POST",
                credentials: "include"
            });
            window.location.reload()
        });
    }
}

updateAuthButton();
