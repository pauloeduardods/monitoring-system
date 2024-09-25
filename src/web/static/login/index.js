document
  .getElementById("loginForm")
  .addEventListener("submit", async function (event) {
    event.preventDefault();

    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;
    const messageDiv = document.getElementById("message");
    messageDiv.innerHTML = ""; // Clear previous messages

    const response = await fetch(
      `${window.location.origin}/api/v1/auth/login`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password }),
      }
    );

    const data = await response.json();

    if (response.status === 200) {
      localStorage.setItem("authToken", data.token);
      window.location.href = "/web/home";
    } else if (response.status === 401) {
      messageDiv.innerHTML = `<div class="alert alert-danger">${data.message}</div>`;
    } else {
      messageDiv.innerHTML = `<div class="alert alert-danger">An error occurred. Please try again.</div>`;
    }
  });
