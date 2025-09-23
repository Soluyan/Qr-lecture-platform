<script>
  import { onMount } from "svelte";

  let sessionId = "";
  let qrCodeUrl = "";
  let questions = [];
  let ws = null;
  let loading = false;
  let sessionLoading = false;
  let connectionStatus = "disconnected";
  let allowAnonymous = true;
  let settingsLoading = false;

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

      // Загружаем настройки сессии
      await loadSessionSettings();

      // Подключаем WebSocket
      connectWebSocket();
    } catch (error) {
      console.error("Error creating session:", error);
      alert(`Ошибка при создании сессии: ${error.message}`);
    } finally {
      sessionLoading = false;
    }
  }

  async function loadSessionSettings() {
    if (!sessionId) return;

    try {
      const response = await fetch(
        `http://localhost:8080/session/settings/get?session=${sessionId}`
      );
      if (response.ok) {
        const settings = await response.json();
        allowAnonymous = settings.allowAnonymous;
      }
    } catch (error) {
      console.error("Error loading settings:", error);
    }
  }

  async function updateSessionSettings() {
    if (!sessionId) return;

    settingsLoading = true;
    try {
      const response = await fetch(
        `http://localhost:8080/session/settings?session=${sessionId}`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            allowAnonymous: allowAnonymous,
          }),
        }
      );

      if (response.ok) {
        console.log("Settings updated successfully");
      } else {
        throw new Error("Failed to update settings");
      }
    } catch (error) {
      console.error("Error updating settings:", error);
      alert("Ошибка при обновлении настроек");
    } finally {
      settingsLoading = false;
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

  // Обновляем настройки при изменении чекбокса
  $: if (sessionId && allowAnonymous !== undefined) {
    updateSessionSettings();
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

      <!-- Добавляем настройки сессии -->
      <div class="session-settings">
        <h3>Настройки сессии</h3>
        <label class="setting-option">
          <input
            type="checkbox"
            bind:checked={allowAnonymous}
            disabled={settingsLoading}
          />
          <span>Разрешить анонимные вопросы</span>
        </label>
        {#if settingsLoading}
          <span class="settings-saving">Сохранение...</span>
        {/if}
      </div>

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
            <button
              on:click={() => deleteQuestion(question.id)}
              class="delete-btn"
            >
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
    flex-direction: column;
    height: 100vh;
    background: white;
    font-family: "Inter", sans-serif;
  }

  @media (min-width: 768px) {
    .teacher-container {
      flex-direction: row;
    }
  }

  .qr-section {
    width: 100%;
    padding: 1.5rem;
    background: #f8fafc;
    border-bottom: 2px solid #e2e8f0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }

  @media (min-width: 768px) {
    .qr-section {
      width: 320px;
      border-right: 2px solid #e2e8f0;
      border-bottom: none;
      height: 100vh;
      position: sticky;
      top: 0;
    }
  }

  .questions-section {
    flex: 1;
    padding: 1.5rem;
    overflow-y: auto;
    min-height: 0;
  }

  /* 
  Old
  */
  .session-settings {
    background: #f5f5f5;
    padding: 1rem;
    border-radius: 8px;
    width: 100%;
  }

  .session-settings h3 {
    margin: 0 0 1rem 0;
    font-size: 1.1rem;
  }

  .setting-option {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
  }

  .settings-saving {
    font-size: 0.8rem;
    color: #666;
    margin-left: 1.5rem;
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

  .connected {
    color: green;
  }
  .connecting {
    color: orange;
  }
  .error {
    color: darkred;
  }
  .disconnected {
    color: red;
  }
</style>
