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
    <h2>Вопросы</h2>
    {#if !questions || questions.length === 0}
      <p>Вопросов пока нет.</p>
    {:else}
      <div class="questions-list">
        {#each questions.slice().reverse() as question (question.id)}
          <div class="question-card">
            <div class="question-content">
              <div class="question-header">
                <span class="question-author"
                  >{question.author || "Anonymous"}</span
                >
                <span class="question-time"
                  >{new Date(question.createdAt).toLocaleTimeString()}</span
                >
              </div>
              <p class="question-text">{question.text}</p>
            </div>
            <button
              on:click={() => deleteQuestion(question.id)}
              class="delete-btn"
              aria-label="Удалить вопрос"
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
  @import url("https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap");

  .teacher-container {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background: white;
    font-family: "Inter", sans-serif;
    font-size: 16px;
    line-height: 1.5;
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

  .qr-code {
    width: 256px;
    height: 256px;
    margin-bottom: 1rem;
    filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.1));
    transition: filter 0.3s ease;
    animation: fadeInScale 0.6s ease-out;
  }

  .qr-code:hover {
    filter: drop-shadow(0 8px 16px rgba(0, 0, 0, 0.15));
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

  @media (max-width: 767px) {
    .question-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.25rem;
    }

    .question-card {
      padding: 0.75rem;
      gap: 0.75rem;
    }

    .question-text {
      font-size: 0.9rem;
      line-height: 1.4;
    }
  }

  .questions-section {
    flex: 1;
    padding: 1.5rem;
    overflow-y: auto;
    min-height: 0;
  }

  .questions-section h2 {
    color: #0078cf;
    font-weight: 700;
    text-align: center;
    margin-bottom: 2.5rem;
    position: relative;
  }

  .questions-section::-webkit-scrollbar {
    width: 8px;
  }

  .questions-section::-webkit-scrollbar-track {
    background: #f1f1f1;
    border-radius: 4px;
  }

  .questions-section::-webkit-scrollbar-thumb {
    background: #cbd5e0;
    border-radius: 4px;
  }

  .questions-section::-webkit-scrollbar-thumb:hover {
    background: #a0aec0;
  }

  @media (min-width: 1200px) {
    .qr-section {
      width: 380px;
      padding: 2rem;
    }

    .questions-section {
      padding: 2rem 3rem;
    }
  }

  .question-text {
    color: #2d3748;
    font-size: 0.95rem;
    line-height: 1.5;
    word-break: break-word;
    overflow-wrap: break-word;
    hyphens: auto;
    white-space: pre-wrap;
    max-width: 100%;
    margin: 0;
  }

  .questions-section h2::after {
    content: "";
    position: absolute;
    bottom: -0.5rem;
    left: 50%;
    transform: translateX(-50%);
    width: 60px;
    height: 3px;
    background: linear-gradient(135deg, #0078cf 0%, #48bb78 100%);
    border-radius: 2px;
  }

  .session-info .sid {
    color: #0078cf;
    font-weight: 600;
  }

  button:not(.delete-btn) {
    background: #0078cf;
    color: white;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  button:not(.delete-btn):focus {
    outline: 3px solid #0078cf;
    outline-offset: 2px;
    box-shadow: 0 0 0 4px rgba(0, 120, 207, 0.2);
  }

  button:not(.delete-btn):hover {
    background: #0066b3;
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(0, 120, 207, 0.3);
  }

  .delete-btn {
    background: #e53e3e;
    color: white;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;
  }

  .delete-btn:hover {
    background: #c53030;
    transform: scale(1.1);
  }

  .delete-btn:focus {
    outline: 3px solid #e53e3e;
    outline-offset: 2px;
  }

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

  @keyframes fadeInScale {
    from {
      opacity: 0;
      transform: scale(0.8);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  .questions-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    width: 100%;
  }

  .question-card {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    animation: slideIn 0.4s ease-out;
    transform-origin: top;
    width: 100%;
    box-sizing: border-box;
  }

  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateY(-10px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  .question-card:hover {
    border-color: #0078cf;
    background: #f8fafc;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 120, 207, 0.15);
  }

  .question-card.removing {
    animation: slideOut 0.3s ease-in forwards;
  }

  @keyframes slideOut {
    to {
      opacity: 0;
      transform: translateX(100px) scale(0.8);
      height: 0;
      padding: 0;
      margin: 0;
      border: none;
    }
  }

  .question-card.new-question {
    border-left: 4px solid #48bb78;
    animation: pulse 2s ease-in-out;
  }

  @keyframes pulse {
    0% {
      box-shadow: 0 0 0 0 rgba(72, 187, 120, 0.4);
    }
    70% {
      box-shadow: 0 0 0 10px rgba(72, 187, 120, 0);
    }
    100% {
      box-shadow: 0 0 0 0 rgba(72, 187, 120, 0);
    }
  }

  .question-content {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .question-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .question-author {
    color: #1a365d;
    font-weight: 600;
    font-size: 0.9rem;
    flex-shrink: 0;
  }

  .question-time {
    color: #a0aec0;
    font-size: 0.8rem;
    flex-shrink: 0;
  }

  .question-text {
    color: #2d3748;
    font-size: 0.95rem;
    line-height: 1.5;
    word-break: break-word;
    overflow-wrap: break-word;
    hyphens: auto;
    white-space: pre-wrap;
    max-width: 100%;
    margin: 0;
  }

  .session-loading {
    position: relative;
    color: #718096;
  }

  .session-loading::after {
    content: "⠋";
    animation: dots 1.5s steps(5, end) infinite;
    margin-left: 0.5rem;
    display: inline-block;
  }

  @keyframes dots {
    0%,
    20% {
      content: "⠋";
    }
    40% {
      content: "⠙";
    }
    60% {
      content: "⠹";
    }
    80% {
      content: "⠸";
    }
    100% {
      content: "⠼";
    }
  }

  .connection-status {
    transition: all 0.3s ease;
  }

  .connection-status.connecting {
    animation: pulseStatus 1.5s infinite;
  }

  @keyframes pulseStatus {
    0% {
      opacity: 1;
    }
    50% {
      opacity: 0.7;
    }
    100% {
      opacity: 1;
    }
  }

  .connection-status .connected {
    color: #38a169;
  }
  .connection-status .connecting {
    color: #d69e2e;
  }
  .connection-status .error {
    color: #e53e3e;
  }
  .disconnected {
    color: red;
  }
</style>
