import { marked } from "https://cdn.jsdelivr.net/npm/marked/src/marked.min.js";

const API_BASE_URL = "http://localhost:8080/api";
const UPLOADS_URL = `${API_BASE_URL}/uploads`;

// fetch and render all uploads into the library list
async function loadUploads() {
    const list = document.getElementById("uploadsList");

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
                    <span class="upload-item-name">${upload.filename}</span>
                    <div class="upload-item-meta">
                        <span class="upload-item-date">${date}</span>
                        <span class="upload-item-type ${ext}">${ext}</span>
                    </div>
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
    document.querySelectorAll(".upload-item").forEach(el => el.classList.remove("active"));
    item.classList.add("active");

    const title = document.getElementById("notesPanelTitle");
    const content = document.getElementById("notesContent");

    title.textContent = name;
    content.innerHTML = `<div class="empty-state"><span class="empty-icon">◌</span><span>Loading study sheet…</span></div>`;

    try {
        const response = await fetch(`${UPLOADS_URL}/${id}/notes`);
        const data = await response.json();

        if (!data || !data.content) {
            content.innerHTML = `<div class="empty-state"><span class="empty-icon">◈</span><span>No study sheet found for this upload.</span></div>`;
            return;
        }

        content.textContent = data.content;

    } catch (err) {
        content.innerHTML = marked.parse(data.content);
    }
}

// wire up the refresh button
document.getElementById("printRows").addEventListener("click", loadUploads);

// load uploads on page load
loadUploads();
