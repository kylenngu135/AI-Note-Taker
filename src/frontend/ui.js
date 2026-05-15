// frontend/ui.js

const API_BASE_URL = "https://ai-note-taker-j35g.onrender.com";

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
    // hide notes tags bar
    const notesTagsBar = document.getElementById("notesTagsBar");
    if (notesTagsBar) notesTagsBar.classList.add("hidden");
}

document.getElementById("newChatBtn").addEventListener("click", showWelcomeView);

async function updateAuthButton() {
    const response = await fetch(`${API_BASE_URL}/api/auth/me`, {
        credentials: "include"
    });

    const signInBtn = document.getElementById("signInBtn");

    if (response.ok) {
        signInBtn.textContent = "Sign Out";
        signInBtn.href = "#";
        signInBtn.addEventListener("click", async (e) => {
            e.preventDefault();
            await fetch(`${API_BASE_URL}/api/auth/logout`, {
                method: "POST",
                credentials: "include"
            });
            window.location.reload()
        });
    }
}

updateAuthButton();
