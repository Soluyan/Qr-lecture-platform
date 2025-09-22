<script>
  import { onMount } from "svelte";
  let sessionId = "";
  let qrCodeUrl = "";
  let questions = [];
  let ws = null;
  let loading = false;
  let sessionLoading = false;
  let connectionStatus = "disconnected";

  onMount(() => {
    console.log("TeacherView mounted");
    return () => {
      if (ws) {
        ws.close();
      }
    };
  });

  async function createSession() {
    sessionLoading = true;
    qrCodeUrl = "";
    questions = [];
    sessionId = "";
    connectionStatus = "disconnected";
    
    if (ws) {
      ws.close();
      ws = null;
    }

    try {
      const response = await fetch("http://localhost:8080/create-session");

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();

      if (!data.sessionId || !data.qrCode) {
        throw new Error("Invalid response format from server");
      }

      sessionId = data.sessionId;
      qrCodeUrl = `data:image/png;base64,${data.qrCode}`;

      console.log("Session created successfully:", sessionId);
      
      // Подключаем WebSocket
      connectWebSocket();
      
    } catch (error) {
      console.error("Error creating session:", error);
      alert(`Ошибка при создании сессии: ${error.message}`);
    } finally {
      sessionLoading = false;
    }
  }

  function connectWebSocket() {
    console.log("connectWebSocket called with sessionId:", sessionId);

    if (!sessionId) {
      alert("Session ID не найден");
      return;
    }

    if (ws) {
      ws.close();
      ws = null;
    }

    connectionStatus = "connecting"; // Устанавливаем статус подключения
    
    try {
      ws = new WebSocket(`ws://localhost:8080/ws?session=${sessionId}`);

      ws.onopen = () => {
        console.log("WebSocket connected for session:", sessionId);
        connectionStatus = "connected";
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          questions = Array.isArray(data) ? data : [];
          console.log("Received questions:", questions.length);
        } catch (error) {
          console.error("Error parsing questions:", error);
          questions = [];
        }
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        connectionStatus = "error";
      };

      ws.onclose = (event) => {
        console.log("WebSocket disconnected:", event.code, event.reason);
        connectionStatus = "disconnected";
        ws = null;
      };
    } catch (error) {
      console.error("WebSocket creation error:", error);
      connectionStatus = "error";
      alert("Ошибка создания WebSocket соединения");
    }
  }

  function deleteQuestion(id) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ action: "delete", question_id: id }));
    }
  }
</script>

<div class="teacher-container">
  <div class="qr-section">
    {#if sessionLoading}
      <p>Создание сессии...</p>
    {:else if qrCodeUrl}
      <div class="session-info">
        <strong>Session ID:</strong> <span class="sid">{sessionId}</span>
      </div>
      <img src={qrCodeUrl} alt="Lecture QR Code" class="qr-code" />
      <div class="connection-status">
        {#if connectionStatus === "connected"}
          <p class="connected">✅ Подключено к вопросам</p>
        {:else if connectionStatus === "connecting"}
          <p class="connecting">⏳ Подключение...</p>
        {:else if connectionStatus === "error"}
          <p class="error">❌ Ошибка подключения</p>
          <button on:click={connectWebSocket}>Повторить подключение</button>
        {:else}
          <p class="disconnected">❌ Не подключено</p>
          <button on:click={connectWebSocket}>Подключиться к вопросам</button>
        {/if}
      </div>
    {:else}
      <button on:click={createSession}>Start New Lecture</button>
    {/if}
  </div>

  <div class="questions-section">
    <h2>Student Questions</h2>
    {#if !questions || questions.length === 0}
      <p>Вопросов пока нет.</p>
    {:else}
      <div class="questions-list">
        {#each questions as question (question.id)}
          <div class="question-card">
            <div class="question-content">
              <strong>{question.author || "Anonymous"}:</strong>
              <p>{question.text}</p>
              <small>{new Date(question.createdAt).toLocaleTimeString()}</small>
            </div>
            <button on:click={() => deleteQuestion(question.id)} class="delete-btn">
              ×
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .teacher-container {
    display: flex;
    height: 100vh;
  }

  .qr-section {
    width: 30%;
    padding: 2rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    border-right: 1px solid #eee;
  }

  .qr-code {
    width: 256px;
    height: 256px;
    margin-bottom: 1rem;
  }

  .questions-section {
    width: 70%;
    padding: 2rem;
  }

  .questions-list {
    margin-top: 1rem;
  }

  .question-card {
    display: flex;
    justify-content: space-between;
    padding: 1rem;
    margin-bottom: 1rem;
    border: 1px solid #ddd;
    border-radius: 4px;
  }

  .question-content {
    flex-grow: 1;
  }
</style>
