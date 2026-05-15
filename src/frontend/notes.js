import { marked } from "https://cdn.jsdelivr.net/npm/marked/lib/marked.esm.js";

const API_BASE_URL = "https://ai-note-taker-j35g.onrender.com";
const UPLOADS_URL = `${API_BASE_URL}/api/uploads`;

const TAG_COLORS = ['#ef4444', '#f97316', '#eab308', '#22c55e', '#3b82f6', '#8b5cf6', '#ec4899', '#6b7280'];

// module-level state
let currentUploadId = null;
let currentUploadName = null;
let currentTags = [];
let uploadsCache = [];
let activeTagFilter = null;
let selectedTagColor = TAG_COLORS[0];

export function getCurrentUploadId() {
    return currentUploadId;
}

// --- Sidebar tag filter ---

function setTagFilter(tagName) {
    activeTagFilter = (activeTagFilter === tagName) ? null : tagName;
    loadUploads();
    updateTagFilterIndicator();
}

function updateTagFilterIndicator() {
    const indicator = document.getElementById("tagFilterIndicator");
    const filterName = document.getElementById("tagFilterName");
    if (!indicator) return;
    if (activeTagFilter) {
        indicator.classList.remove("hidden");
        if (filterName) filterName.textContent = `#${activeTagFilter}`;
    } else {
        indicator.classList.add("hidden");
    }
}

function updateSidebarItemTags(uploadId, tags) {
    const item = document.querySelector(`.upload-item[data-id="${uploadId}"]`);
    if (!item) return;

    let tagsContainer = item.querySelector('.upload-item-tags');
    if (!tagsContainer) {
        tagsContainer = document.createElement('div');
        tagsContainer.className = 'upload-item-tags';
        item.appendChild(tagsContainer);
    }

    tagsContainer.innerHTML = tags.map(tag => {
        const colorStyle = tag.color ? ` style="--tag-color: ${tag.color}"` : '';
        return `<span class="tag-pill ${tag.type}"${colorStyle} data-tag-name="${escapeHtml(tag.name)}">${escapeHtml(tag.name)}</span>`;
    }).join('');

    tagsContainer.querySelectorAll('.tag-pill').forEach(pill => {
        pill.addEventListener('click', (e) => {
            e.stopPropagation();
            setTagFilter(pill.dataset.tagName);
        });
    });
}

// --- Fetch and render uploads list ---

async function loadUploads() {
    const list = document.getElementById("recentsList");
    const url = activeTagFilter
        ? `${UPLOADS_URL}?tag=${encodeURIComponent(activeTagFilter)}`
        : UPLOADS_URL;

    try {
        const response = await fetch(url, { credentials: "include" });
        const uploads = await response.json();

        uploadsCache = uploads || [];

        if (!uploadsCache.length) {
            list.innerHTML = activeTagFilter
                ? `<div class="empty-state"><span class="empty-icon">◎</span><span>No uploads tagged #${escapeHtml(activeTagFilter)}.</span></div>`
                : `<div class="empty-state"><span class="empty-icon">◎</span><span>No uploads yet. Add something above.</span></div>`;
            return null;
        }

        list.innerHTML = uploadsCache.map(upload => {
            const ext = upload.filename.split(".").pop().toLowerCase();
            const date = new Date(upload.created_at).toLocaleDateString("en-US", {
                month: "short", day: "numeric", year: "numeric"
            });

            const tagPills = (upload.tags || []).map(tag => {
                const colorStyle = tag.color ? ` style="--tag-color: ${tag.color}"` : '';
                return `<span class="tag-pill ${tag.type}"${colorStyle} data-tag-name="${escapeHtml(tag.name)}">${escapeHtml(tag.name)}</span>`;
            }).join('');

            const isActive = upload.id === currentUploadId ? ' active' : '';

            return `
                <div class="upload-item${isActive}" data-id="${upload.id}" data-name="${upload.filename}">
                    <div class="upload-item-header">
                        <span class="upload-item-date">${date}</span>
                        <span class="upload-item-name">${upload.filename}</span>
                        <span class="upload-item-type ${ext}">${ext}</span>
                    </div>
                    ${tagPills ? `<div class="upload-item-tags">${tagPills}</div>` : ''}
                </div>`;
        }).join("");

        list.querySelectorAll(".upload-item").forEach(item => {
            item.addEventListener("click", (e) => {
                if (!e.target.classList.contains("tag-pill")) {
                    loadNotes(item);
                }
            });
            item.querySelectorAll(".tag-pill").forEach(pill => {
                pill.addEventListener("click", (e) => {
                    e.stopPropagation();
                    setTagFilter(pill.dataset.tagName);
                });
            });
        });

        return uploadsCache;

    } catch (err) {
        list.innerHTML = `
            <div class="empty-state">
                <span class="empty-icon">⚠</span>
                <span>Failed to load uploads.</span>
            </div>`;
        return null;
    }
}

// --- Notes view tags ---

function renderNotesTagsBar(uploadId, tags) {
    const pillsContainer = document.getElementById("notesTagsPills");
    const addBtn = document.getElementById("tagAddBtn");
    if (!pillsContainer) return;

    pillsContainer.innerHTML = tags.map(tag => {
        const colorStyle = tag.color ? `style="--tag-color: ${tag.color}"` : '';
        if (tag.type === 'user') {
            return `<span class="tag-pill user" ${colorStyle} data-tag-id="${tag.id}">#${escapeHtml(tag.name)}<button class="tag-remove" data-tag-id="${tag.id}" aria-label="Remove tag">✕</button></span>`;
        }
        return `<span class="tag-pill ${tag.type}" ${colorStyle}>#${escapeHtml(tag.name)}</span>`;
    }).join('');

    if (addBtn) addBtn.classList.remove("hidden");

    pillsContainer.querySelectorAll('.tag-remove').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const tagId = btn.dataset.tagId;
            removeTagFromUpload(uploadId, tagId).catch(err => console.error('Remove tag failed:', err));
        });
    });
}

async function addTagToUpload(uploadId, name, color) {
    const resp = await fetch(`${UPLOADS_URL}/${uploadId}/tags`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, color }),
        credentials: 'include',
    });
    if (!resp.ok) throw new Error('Failed to add tag');
    const tag = await resp.json();

    // Update local state (deduplicate by id)
    currentTags = [...currentTags.filter(t => t.id !== tag.id), tag];
    renderNotesTagsBar(uploadId, currentTags);
    updateSidebarItemTags(uploadId, currentTags);

    const cached = uploadsCache.find(u => u.id === uploadId);
    if (cached) cached.tags = [...currentTags];
}

async function removeTagFromUpload(uploadId, tagId) {
    const resp = await fetch(`${UPLOADS_URL}/${uploadId}/tags/${tagId}`, {
        method: 'DELETE',
        credentials: 'include',
    });
    if (!resp.ok) throw new Error('Failed to remove tag');

    currentTags = currentTags.filter(t => t.id !== tagId);
    renderNotesTagsBar(uploadId, currentTags);
    updateSidebarItemTags(uploadId, currentTags);

    const cached = uploadsCache.find(u => u.id === uploadId);
    if (cached) cached.tags = [...currentTags];
}

// --- render a single message bubble ---
function renderMessageBubble(role, content, label = null) {
    const isUser = role === "user";
    const bubbleClass = isUser ? "message-bubble user-message" : "message-bubble";
    const messageLabel = label || (isUser ? "You" : "AI Study Assistant");

    return `
        <div class="${bubbleClass}">
            <div class="message-label">${escapeHtml(messageLabel)}</div>
            <div>${marked.parse(content)}</div>
        </div>
    `;
}

// --- fetch and display notes for a specific upload ---
export async function loadNotesById(id, name) {
    const messagesArea = document.getElementById("messagesArea");
    const notesView = document.getElementById("notesView");
    const welcomeView = document.getElementById("welcomeView");
    const processingView = document.getElementById("processingView");
    const messageBar = document.getElementById("messageBar");
    const notesTagsBar = document.getElementById("notesTagsBar");

    // switch views
    processingView.classList.add("hidden");
    welcomeView.classList.add("hidden");
    notesView.classList.remove("hidden");
    messageBar.classList.remove("hidden");
    if (notesTagsBar) notesTagsBar.classList.remove("hidden");

    // show export and delete buttons
    document.getElementById("exportBtn").classList.remove("hidden");
    document.getElementById("deleteBtn").classList.remove("hidden");

    // update topbar title
    document.getElementById("topbarTitle").textContent = name;

    // store current upload id/name
    currentUploadId = id;
    currentUploadName = name;

    // reset tags bar
    currentTags = [];
    const pillsContainer = document.getElementById("notesTagsPills");
    const addBtn = document.getElementById("tagAddBtn");
    if (pillsContainer) pillsContainer.innerHTML = '';
    if (addBtn) addBtn.classList.add("hidden");

    // mark active in sidebar
    document.querySelectorAll(".upload-item").forEach(item => {
        item.classList.toggle("active", item.dataset.id === id);
    });

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

        // Render tags below topbar
        currentTags = data.tags || [];
        renderNotesTagsBar(id, currentTags);

        // Build conversation history in chronological order
        let messagesHTML = "";

        if (data.history && data.history.length > 0) {
            const baseName = name.replace(/\.[^.]+$/, '');
            let firstMessageRendered = false;
            for (const item of data.history) {
                if (item.role === "user" && item.prompt) {
                    if (!firstMessageRendered) {
                        messagesHTML += renderMessageBubble("user", `${baseName}_transcription.txt`, "You");
                        firstMessageRendered = true;
                    } else {
                        messagesHTML += renderMessageBubble("user", item.prompt, "You");
                    }
                } else if (item.role === "assistant" && item.content) {
                    messagesHTML += renderMessageBubble("assistant", item.content, "AI Study Assistant");
                }
            }
        } else {
            // Fallback for existing uploads without history
            messagesHTML += renderMessageBubble("assistant", data.note.content, "AI Study Assistant");
        }

        messagesArea.innerHTML = messagesHTML;
        setTimeout(() => {
            notesView.scrollTop = notesView.scrollHeight;
        }, 50);

    } catch (err) {
        messagesArea.innerHTML = `<div class="message-bubble"><div class="message-label">Failed to load study sheet.</div></div>`;
    }
}

// fetch and display the notes for a specific upload (from sidebar click)
async function loadNotes(item) {
    const id = item.dataset.id;
    const name = item.dataset.name;
    await loadNotesById(id, name);
}

// load uploads on page load
loadUploads();

// --- Tag add form ---
export function addTagListeners() {
    const addBtn = document.getElementById("tagAddBtn");
    const addForm = document.getElementById("tagAddForm");
    const addInput = document.getElementById("tagAddInput");
    const addSubmit = document.getElementById("tagAddSubmit");
    const addCancel = document.getElementById("tagAddCancel");
    const colorPicker = document.getElementById("tagColorPicker");
    const filterClear = document.getElementById("tagFilterClear");

    // build color swatches
    if (colorPicker) {
        colorPicker.innerHTML = TAG_COLORS.map((color, i) =>
            `<button class="color-swatch${i === 0 ? ' selected' : ''}" style="background: ${color}" data-color="${color}" type="button" aria-label="${color}"></button>`
        ).join('');

        colorPicker.querySelectorAll('.color-swatch').forEach(swatch => {
            swatch.addEventListener('click', () => {
                colorPicker.querySelectorAll('.color-swatch').forEach(s => s.classList.remove('selected'));
                swatch.classList.add('selected');
                selectedTagColor = swatch.dataset.color;
            });
        });
    }

    if (addBtn && addForm) {
        addBtn.addEventListener('click', () => {
            addForm.classList.remove("hidden");
            addBtn.classList.add("hidden");
            if (addInput) addInput.focus();
        });
    }

    const submitTag = async () => {
        if (!addInput) return;
        const rawName = addInput.value.trim()
            .toLowerCase()
            .replace(/\s+/g, '-')
            .replace(/[^a-z0-9\-]/g, '');
        if (!rawName || !currentUploadId) return;

        addInput.value = '';
        if (addForm) addForm.classList.add("hidden");
        if (addBtn) addBtn.classList.remove("hidden");

        try {
            await addTagToUpload(currentUploadId, rawName, selectedTagColor);
        } catch (err) {
            console.error('Failed to add tag:', err);
        }
    };

    if (addSubmit) addSubmit.addEventListener('click', submitTag);

    if (addInput) {
        addInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') { e.preventDefault(); submitTag(); }
            if (e.key === 'Escape') {
                addInput.value = '';
                if (addForm) addForm.classList.add("hidden");
                if (addBtn) addBtn.classList.remove("hidden");
            }
        });
    }

    if (addCancel) {
        addCancel.addEventListener('click', () => {
            if (addInput) addInput.value = '';
            if (addForm) addForm.classList.add("hidden");
            if (addBtn) addBtn.classList.remove("hidden");
        });
    }

    if (filterClear) {
        filterClear.addEventListener('click', () => {
            activeTagFilter = null;
            loadUploads();
            updateTagFilterIndicator();
        });
    }
}

// --- Export modal handling ---
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

// --- Send button (regenerate) handling ---
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
        messagesArea.innerHTML += renderMessageBubble("user", prompt, "You");

        // Clear input
        followUpInput.value = "";

        // Show loading indicator
        const loadingBubble = document.createElement("div");
        loadingBubble.className = "message-bubble";
        loadingBubble.innerHTML = `<div class="message-label">Generating…</div>`;
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
            messagesArea.innerHTML += renderMessageBubble("assistant", data.content, "AI Study Assistant");
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
