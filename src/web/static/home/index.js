document.addEventListener("DOMContentLoaded", async function () {
  const messageDiv = document.getElementById("message");
  const token = localStorage.getItem("authToken");

  if (!token) {
    messageDiv.textContent = "Unauthorized: No token found";

    setTimeout(() => {
      window.location.href = "/web/login";
    }, 3000);
    return;
  }

  let previousUrls = [];

  const fetchCameraDetails = async () => {
    try {
      const response = await fetch(`/api/v1/monitoring/camera/details`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error("Failed to fetch camera details");
      }

      const cameraDetails = await response.json();
      return cameraDetails;
    } catch (error) {
      console.error("Error fetching camera details:", error);
      messageDiv.textContent = "Error fetching camera details.";
      return [];
    }
  };

  const createVideoElements = (cameras) => {
    const container = document.querySelector(".card");
    cameras.forEach((camera, index) => {
      const div = document.createElement("div");
      div.classList.add("mb-3");
      div.innerHTML = `
        <img
          id="videoStream${camera.ID}"
          width="640"
          height="480"
          alt="Camera Stream ${camera.ID}"
          class="img-fluid videoStream mb-3"
        />
      `;
      container.appendChild(div);
      previousUrls.push(null);
    });
  };

  const connectWebSocket = (cameraIndex) => {
    const ws = new WebSocket(
      `ws://${
        window.location.host
      }/api/v1/ws/video/${cameraIndex}?token=${encodeURIComponent(token)}`
    );
    ws.binaryType = "arraybuffer";

    ws.onmessage = function (event) {
      const arrayBuffer = event.data;
      const blob = new Blob([arrayBuffer], { type: "image/jpeg" });
      const imageUrl = URL.createObjectURL(blob);

      if (previousUrls[cameraIndex]) {
        URL.revokeObjectURL(previousUrls[cameraIndex]);
      }

      const videoStream = document.getElementById(`videoStream${cameraIndex}`);
      if (videoStream) {
        videoStream.src = imageUrl;
        previousUrls[cameraIndex] = imageUrl;
      }
    };

    ws.onclose = function (event) {
      console.log(`WebSocket for camera ${cameraIndex} closed: `, event);

      if (event.code === 1006) {
        messageDiv.textContent = `Connection closed abnormally. Please check your credentials or token.`;

        if (cameraIndex === 0 || cameraIndex === 1) {
          messageDiv.textContent = `Error: Camera ${cameraIndex} not found or inaccessible.`;
        } else {
          localStorage.removeItem("authToken");
          setTimeout(() => {
            window.location.href = "/web/login";
          }, 3000);
        }
      }
    };

    ws.onerror = function (event) {
      messageDiv.textContent = `WebSocket error for camera ${cameraIndex}: ${event.message}`;

      if (event.message.includes("401") || event.message.includes("403")) {
        localStorage.removeItem("authToken");
        setTimeout(() => {
          window.location.href = "/web/login";
        }, 3000);
      }
    };
  };

  const cameraDetails = await fetchCameraDetails();
  if (cameraDetails.length > 0) {
    createVideoElements(cameraDetails);
    cameraDetails.forEach((camera) => {
      connectWebSocket(camera.ID);
    });
  } else {
    messageDiv.textContent = "No cameras found.";
  }
});
