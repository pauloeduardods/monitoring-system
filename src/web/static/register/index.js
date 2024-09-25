document
  .getElementById("registerForm")
  .addEventListener("submit", async function (event) {
    event.preventDefault();

    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;
    const messageDiv = document.getElementById("message");
    messageDiv.innerHTML = "";

    const token = localStorage.getItem("authToken");
    console.log(token);

    const headers = {
      "Content-Type": "application/json",
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const response = await fetch(
      `${window.location.origin}/api/v1/auth/register`,
      {
        method: "POST",
        headers: headers,
        body: JSON.stringify({ username, password }),
      }
    );

    const data = await response.json();

    if (response.status === 200) {
      messageDiv.innerHTML = `<div class="alert alert-success">${data.message}</div>`;
      setTimeout(() => {
        window.location.href = "/web/login";
      }, 2000);
    } else if (response.status === 409) {
      messageDiv.innerHTML = `<div class="alert alert-danger">${data.message}</div>`;
    } else {
      messageDiv.innerHTML = `<div class="alert alert-danger">An error occurred. Please try again.</div>`;
    }
  });
