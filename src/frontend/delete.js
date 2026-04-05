import { getCurrentUploadId } from './notes.js';

const UPLOADS_BASE_URL = "http://localhost:8080/api/uploads";

export function addDeleteButtonListeners() {
    const deleteBtn = document.getElementById("deleteBtn");
    const deleteModal = document.getElementById("deleteModal");
    const closeDeleteModal = document.getElementById("closeDeleteModal");
    const confirmDeleteBtn = document.getElementById("confirmDeleteBtn");
    const cancelDeleteBtn = document.getElementById("cancelDeleteBtn");

    deleteBtn.addEventListener("click", () => {
        deleteModal.classList.remove("hidden");
    });

    const closeModal = () => {
        deleteModal.classList.add("hidden");
    };

    closeDeleteModal.addEventListener("click", closeModal);
    cancelDeleteBtn.addEventListener("click", closeModal);

    deleteModal.addEventListener("click", (e) => {
        if (e.target === deleteModal) {
            closeModal();
        }
    });

    confirmDeleteBtn.addEventListener("click", async () => {
        closeModal();
        await deleteRow();
    });
}

async function deleteRow() {
    const id = getCurrentUploadId();
    if (!id) {
        alert("No document selected for deletion.");
        return;
    }

    try {
        const response = await fetch(`${UPLOADS_BASE_URL}/${id}`, {
            method: 'DELETE',
            credentials: "include"
        });

        if (response.status === 204) {
            window.location.reload();
        } else {
            alert("Failed to delete document");
        }
    } catch (error) {
        alert('Failed to connect to server: ' + error.message)
    }
}
