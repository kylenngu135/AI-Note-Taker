const UPLOADS_BASE_URL = "http://localhost:8080/api/uploads";

export function addDeleteButtonListeners() {
    document.getElementById("deleteButton").addEventListener("click", () => {
        deleteRow();
    });
}

async function deleteRow() {
    let id = document.getElementById("idToDelete").value;

    try {
        const response = await fetch(`${UPLOADS_BASE_URL}/${id}`, {
            method: 'DELETE'
        });

        if (response.status === 204) {
            alert(response.status)
        }
    } catch (error) {
        alert('Failed to connect to server: ' + error.message)
    }
}
