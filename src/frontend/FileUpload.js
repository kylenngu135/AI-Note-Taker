import { loadUploads } from './notes.js';

const UPLOADS_BASE_URL = "http://localhost:8080/api/uploads";

export function addUploadButtonListeners() {
    document.getElementById("uploadButton").addEventListener("click", uploadFile);
}

async function uploadFile() {
    let file = document.getElementById('fileToUpload').files[0];

    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch(`${UPLOADS_BASE_URL}`, {
            method: 'POST',
            body: formData,
            credentials: "include"
        });

        const data = await response.json();
        console.log(data);

        if (data.error) {
            alert('Error: ' + data.error);
        } else {
            alert('uploaded');
            loadUploads();
        }
    } catch (error) {
        alert('Failed to connect to server: ' + error.message)
    }
}
