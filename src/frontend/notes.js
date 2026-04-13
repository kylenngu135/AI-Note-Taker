import { marked } from "https://cdn.jsdelivr.net/npm/marked/lib/marked.esm.js";

const API_BASE_URL = "http://localhost:8080/api";
const UPLOADS_URL = `${API_BASE_URL}/uploads`;

// module-level state
let currentUploadId = null;
let currentUploadName = null;

export function getCurrentUploadId() {
    return currentUploadId;
}

// fetch and render all uploads into the library list
async function loadUploads() {
    const list = document.getElementById("recentsList");

    try {
        const response = await fetch(UPLOADS_URL);
        const uploads = await response.json();

        if (!uploads || uploads.length === 0) {
            list.innerHTML = `
                <div class="empty-state">
                    <span class="empty-icon">◎</span>
                    <span>No uploads yet. Add something above.</span>
                </div>`;
            return;
        }

        list.innerHTML = uploads.map(upload => {
            const ext = upload.filename.split(".").pop().toLowerCase();
            const date = new Date(upload.created_at).toLocaleDateString("en-US", {
                month: "short", day: "numeric", year: "numeric"
            });

            return `
                <div class="upload-item" data-id="${upload.id}" data-name="${upload.filename}">
                    <span class="upload-item-date">${date}</span>
                    <span class="upload-item-name">${upload.filename}</span>
                    <span class="upload-item-type ${ext}">${ext}</span>
                </div>`;
        }).join("");

        // attach click listeners to each upload item
        list.querySelectorAll(".upload-item").forEach(item => {
            item.addEventListener("click", () => loadNotes(item));
        });

    } catch (err) {
        list.innerHTML = `
            <div class="empty-state">
                <span class="empty-icon">⚠</span>
                <span>Failed to load uploads.</span>
            </div>`;
    }
}

// fetch and display the notes for a specific upload
async function loadNotes(item) {
    const id = item.dataset.id;
    const name = item.dataset.name;

    // mark active
    document.querySelectorAll(".recent-item").forEach(el => el.classList.remove("active"));
    item.classList.add("active");

    const messagesArea = document.getElementById("messagesArea");
    const notesView = document.getElementById("notesView");
    const welcomeView = document.getElementById("welcomeView");

    // switch views
    welcomeView.classList.add("hidden");
    notesView.classList.remove("hidden");

    // show export and delete buttons
    document.getElementById("exportBtn").classList.remove("hidden");
    document.getElementById("deleteBtn").classList.remove("hidden");

    // update topbar title
    document.getElementById("topbarTitle").textContent = name;

    // store current upload id for export/delete operations
    currentUploadId = id;
    currentUploadName = name;

    // show loading state
    messagesArea.innerHTML = `<div class="message-bubble"><div class="message-label">Loading…</div></div>`;

    try {
        const response = await fetch(`${UPLOADS_URL}/${id}/notes`, {
            credentials: "include"
        });
        const data = await response.json();

        if (!data || !data.note || !data.note.content) {
            messagesArea.innerHTML = `<div class="message-bubble"><div class="message-label">No study sheet found.</div></div>`;
            return;
        }

        // Show only the study sheet content
        messagesArea.innerHTML = `
            <div class="message-bubble">
                <div>${marked.parse(data.note.content)}</div>
            </div>
        `;
        messagesArea.scrollTop = messagesArea.scrollHeight;

    } catch (err) {
        messagesArea.innerHTML = `<div class="message-bubble"><div class="message-label">Failed to load study sheet.</div></div>`;
    }
}

// load uploads on page load
loadUploads();

// Export modal handling
export function addExportButtonListeners() {
    const exportBtn = document.getElementById("exportBtn");
    const exportModal = document.getElementById("exportModal");
    const closeExportModal = document.getElementById("closeExportModal");

    exportBtn.addEventListener("click", () => {
        exportModal.classList.remove("hidden");
    });

    closeExportModal.addEventListener("click", () => {
        exportModal.classList.add("hidden");
    });

    exportModal.addEventListener("click", (e) => {
        if (e.target === exportModal) {
            exportModal.classList.add("hidden");
        }
    });

    exportModal.querySelectorAll(".modal-option").forEach(btn => {
        btn.addEventListener("click", () => {
            const format = btn.dataset.format;
            exportNotes(format);
            exportModal.classList.add("hidden");
        });
    });
}

async function exportNotes(format) {
    if (!currentUploadId) {
        alert("No document selected for export.");
        return;
    }

    try {
        const response = await fetch(`${UPLOADS_URL}/${currentUploadId}/notes?format=${format}`, {
            credentials: "include"
        });

        if (!response.ok) {
            throw new Error("Failed to export notes");
        }

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `notes-${currentUploadName}.${format}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
    } catch (err) {
        alert("Failed to export notes: " + err.message);
    }
}

// Send button (regenerate) handling
export function addSendButtonListener() {
    const sendBtn = document.getElementById("sendBtn");
    const followUpInput = document.getElementById("followUpInput");

    const handleSend = async () => {
        const prompt = followUpInput.value.trim();
        if (!prompt) return;

        if (!currentUploadId) {
            alert("No document selected.");
            return;
        }

        const messagesArea = document.getElementById("messagesArea");

        // Append user message bubble
        messagesArea.innerHTML += `
            <div class="message-bubble user-message">
                <div>${escapeHtml(prompt)}</div>
            </div>
        `;

        // Clear input
        followUpInput.value = "";

        // Show loading indicator
        const loadingBubble = document.createElement("div");
        loadingBubble.className = "message-bubble";
        loadingBubble.innerHTML = `<div class="message-label">Regenerating…</div>`;
        messagesArea.appendChild(loadingBubble);
        messagesArea.scrollTop = messagesArea.scrollHeight;

        try {
            const response = await fetch(`${UPLOADS_URL}/${currentUploadId}/notes/regenerate`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ prompt }),
                credentials: "include"
            });

            loadingBubble.remove();

            if (!response.ok) {
                throw new Error("Failed to regenerate notes");
            }

            const data = await response.json();

            // Append AI response
            messagesArea.innerHTML += `
                <div class="message-bubble">
                    <div class="message-label">Regenerated Study Sheet</div>
                    <div>${marked.parse(data.content)}</div>
                </div>
            `;
            messagesArea.scrollTop = messagesArea.scrollHeight;
        } catch (err) {
            loadingBubble.remove();
            messagesArea.innerHTML += `
                <div class="message-bubble">
                    <div class="message-label" style="color: var(--danger);">Error</div>
                    <div>${err.message}</div>
                </div>
            `;
        }
    };

    sendBtn.addEventListener("click", handleSend);

    followUpInput.addEventListener("keydown", (e) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    });
}

function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

export { loadUploads };
