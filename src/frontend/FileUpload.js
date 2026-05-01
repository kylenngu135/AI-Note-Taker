import { loadUploads, loadNotesById } from './notes.js';

const UPLOADS_BASE_URL = "http://localhost:8080/api/uploads";

export function addUploadButtonListeners() {
    document.getElementById("uploadButton").addEventListener("click", uploadFile);
}

async function uploadFile() {
    let file = document.getElementById('fileToUpload').files[0];
    if (!file) {
        alert('Please select a file to upload.');
        return;
    }

    const uploadButton = document.getElementById('uploadButton');
    const welcomeView = document.getElementById('welcomeView');
    const processingView = document.getElementById('processingView');
    const messageBar = document.getElementById('messageBar');

    // Disable button and show processing state
    uploadButton.disabled = true;
    uploadButton.value = 'Processing…';
    welcomeView.classList.add('hidden');
    processingView.classList.remove('hidden');
    messageBar.classList.add('hidden');

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
            showWelcomeView();
        } else {
            // Reload uploads list and display the new note
            const uploads = await loadUploads();
            if (uploads && uploads.length > 0) {
                const latestUpload = uploads[0];
                await loadNotesById(latestUpload.id, latestUpload.filename);
            } else {
                showWelcomeView();
            }
        }
    } catch (error) {
        alert('Failed to connect to server: ' + error.message);
        showWelcomeView();
    } finally {
        uploadButton.disabled = false;
        uploadButton.value = 'Upload';
    }
}

function showWelcomeView() {
    document.getElementById('welcomeView').classList.remove('hidden');
    document.getElementById('processingView').classList.add('hidden');
    document.getElementById('messageBar').classList.add('hidden');
}
