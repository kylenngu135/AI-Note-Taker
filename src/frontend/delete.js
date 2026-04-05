const UPLOADS_BASE_URL = "http://localhost:8080/api/uploads";

export function addDeleteButtonListeners() {
    document.getElementById("deleteBtn").addEventListener("click", () => {
        deleteRow();
    });
}

async function deleteRow() {
    const id = window.currentUploadId;
    if (!id) {
        alert("No document selected for deletion.");
        return;
    }

    if (!confirm("Are you sure you want to delete this document?")) {
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
