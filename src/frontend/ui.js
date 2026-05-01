// frontend/ui.js

// Switch to the welcome/upload view
function showWelcomeView() {
    document.getElementById("welcomeView").classList.remove("hidden");
    document.getElementById("notesView").classList.add("hidden");
    document.getElementById("processingView").classList.add("hidden");
    document.getElementById("messageBar").classList.add("hidden");
    document.getElementById("topbarTitle").textContent = "AI Notes";
    // hide export and delete buttons when not viewing a document
    document.getElementById("exportBtn").classList.add("hidden");
    document.getElementById("deleteBtn").classList.add("hidden");
}

document.getElementById("newChatBtn").addEventListener("click", showWelcomeView);

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
