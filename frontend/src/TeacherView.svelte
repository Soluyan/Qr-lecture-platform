<script>
  import { onMount } from "svelte";
  let sessionId = "";
  let qrCodeUrl = "";
  let questions = [];
  let ws = null;
  let loading = false;
  let sessionLoading = false;

  async function createSession() {
    sessionLoading = true;
    qrCodeUrl = "";
    questions = [];
    sessionId = "";
    ws && ws.close();

    try {
      // Запрашиваем сессию с backend
      const response = await fetch("http://localhost:8080/create-session");

      // Проверяем статус ответа
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();

      // Проверяем наличие необходимых полей
      if (!data.sessionId || !data.qrCode) {
        throw new Error("Invalid response format from server");
      }

      // Получаем sessionId и QR-код в base64 из ответа
      sessionId = data.sessionId;
      qrCodeUrl = `data:image/png;base64,${data.qrCode}`;

      console.log("Session created successfully:", sessionId);
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
      console.error("No sessionId provided");
      alert("Session ID не найден");
      return;
    }

    // Закрываем существующее соединение
    if (ws) {
      ws.close();
      ws = null;
    }

    try {
      ws = new WebSocket(`ws://localhost:8080/ws?session=${sessionId}`);

      ws.onopen = () => {
        console.log("WebSocket connected for session:", sessionId);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          // Убеждаемся, что data - это массив
          questions = Array.isArray(data) ? data : [];
        } catch (error) {
          console.error("Error parsing questions:", error);
          questions = [];
        }
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        alert("Ошибка подключения к WebSocket");
      };

      ws.onclose = () => {
        console.log("WebSocket disconnected");
        ws = null;
      };
    } catch (error) {
      console.error("WebSocket creation error:", error);
      alert("Ошибка создания WebSocket соединения");
    }
    onMount(() => {
      console.log(
        "TeacherView mounted, current ws state:",
        ws ? ws.readyState : "null"
      );
    });
  }

  function deleteQuestion(id) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ action: "delete", question_id: id }));
    }
  }

  // Автоматически подключаемся к WebSocket после создания сессии
  $: if (sessionId && sessionId !== '' && (!ws || ws.readyState !== WebSocket.OPEN)) {
    console.log('Attempting to connect WebSocket for session:', sessionId);
    // Небольшая задержка перед подключением
    setTimeout(connectWebSocket, 500);
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
        {#if ws && ws.readyState === WebSocket.OPEN}
          <p class="connected">✅ Подключено к вопросам</p>
        {:else if ws && ws.readyState === WebSocket.CONNECTING}
          <p class="connecting">⏳ Подключение...</p>
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
      <!-- Добавляем проверку на null -->
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
              class="delete-btn">×</button
            >
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
