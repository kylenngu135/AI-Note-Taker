const UPLOADS_BASE_URL = "http://localhost:8080/api/uploads";

export function addDisplayButtonListeners() {
    document.getElementById("printRows").addEventListener("click", () => {
        displayRows();
    });
}

/*
export function loadUploads() {
    try {
        const response = await fetch(`${UPLOADS_BASE_URL}`, {
            method: 'GET'
        });

        const data = await response.json();

        if (data.error) {
            alert('Error: ' + data.error);
        } else {
            alert('uploaded')
        }
    } catch (error) {
        alert('Failed to connect to server: ' + error.message)
    }
}
*/

async function displayRows() {
    try {
        const response = await fetch(UPLOADS_BASE_URL, {
            method: 'GET'
        });

        const data = await response.json();

        if (data.error) {
            alert('Error: ' + data.error);
        } else {
            alert('uploaded')
        }
    } catch (error) {
        alert('Failed to connect to server: ' + error.message)
    }
}

