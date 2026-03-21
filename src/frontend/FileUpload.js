const UPLOADS_BASE_URL = "http://localhost:8080/api/uploads";

export function addUploadButtonListeners() {
    document.getElementById("uploadDocButton").addEventListener("click", () => {
        uploadFile("documents");
    });
    document.getElementById("uploadVideoButton").addEventListener("click", () => {
        uploadFile("videos");
    });
    document.getElementById("uploadAudioButton").addEventListener("click", () => {
        uploadFile("audios");
    });
}

async function uploadFile(type) {
    let file = document.getElementById('fileToUpload').files[0];

    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch(`${UPLOADS_BASE_URL}/${type}`, {
            method: 'POST',
            body: formData
        });

        const data = await response.json();
        console.log(data);

        if (data.error) {
            alert('Error: ' + data.error);
        } else {
            alert('uploaded')
        }
    } catch (error) {
        alert('Failed to connect to server: ' + error.message)
    }
}
