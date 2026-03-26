const AUTH_URL = "http://localhost:8080/api/auth"

function addButtonListeners() {
    document.getElementById("registerButton").addEventListener("click", () => {
        register();
    });
}

async function register() {
    const data = {
        email: document.getElementById("registerEmailField").value,
        password: document.getElementById("registerPasswordField").value
    }
    
    const response = await fetch(`${AUTH_URL}/register`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(data)
    })

    console.log(response.json())
}

addButtonListeners();
