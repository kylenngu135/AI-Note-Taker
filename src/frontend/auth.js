const AUTH_URL = "https://ai-note-taker-j35g.onrender.com/api/auth"

function addButtonListeners() {
    document.getElementById("registerButton").addEventListener("click", () => {
        register();
    });
    document.getElementById("loginButton").addEventListener("click", () => {
        login();
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

    if (response.status === 201) {
        window.location.href = "/";
    }
}

async function login() {
    const data = {
        email: document.getElementById("loginEmailField").value,
        password: document.getElementById("loginPasswordField").value
    }
    
    const response = await fetch(`${AUTH_URL}/login`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(data)
    })

    if (response.status === 201) {
        window.location.href = "/";
    }
}

addButtonListeners();
