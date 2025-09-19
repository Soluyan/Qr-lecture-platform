<script>
  import { onMount } from 'svelte';
  let sessionId = '';
  let qrCodeUrl = '';
  let questions = [];
  let ws;
  let loading = false;
  let sessionLoading = false;

  async function createSession() {
    sessionLoading = true;
    qrCodeUrl = '';
    questions = [];
    sessionId = '';
    ws && ws.close();

    // Запрашиваем QR-код с backend
    const response = await fetch('http://localhost:8080/create-session');
    const qrBlob = await response.blob();
    qrCodeUrl = URL.createObjectURL(qrBlob);

    sessionLoading = false;
  }

  function connectWebSocket() {
    ws = new WebSocket(`ws://localhost:8080/ws?session=${sessionId}`);
    
    ws.onmessage = (event) => {
      questions = JSON.parse(event.data);
    };
  }

  async function deleteQuestion(id) {
    ws.send(JSON.stringify({ action: 'delete', question_id: id }));
  }
</script>

<div class="teacher-container">
  <div class="qr-section">
    {#if sessionLoading}
      <p>Создание сессии...</p>
    {:else if qrCodeUrl}
      <div class="session-info">
        <strong>Session ID:</strong> <span class="sid">{sessionId ? sessionId : "Не определён"}</span>
      </div>
      <img src={qrCodeUrl} alt="Lecture QR Code" class="qr-code" />
      <p>Session ID: <input bind:value={sessionId} placeholder="Введите Session ID для подключения" /></p>
      <button on:click={connectWebSocket} disabled={!sessionId || ws}>Подключиться к вопросам</button>
    {:else}
      <button on:click={createSession}>Start New Lecture</button>
    {/if}
  </div>

  <div class="questions-section">
    <h2>Student Questions</h2>
    {#if loading}
      <p>Загрузка вопросов...</p>
    {:else if questions.length === 0}
      <p>Вопросов пока нет.</p>
    {:else}
      <div class="questions-list">
        {#each questions as question (question.id)}
          <div class="question-card">
            <div class="question-content">
              <strong>{question.author || 'Anonymous'}:</strong>
              <p>{question.text}</p>
              <small>{new Date(question.createdAt).toLocaleTimeString()}</small>
            </div>
            <button on:click={() => deleteQuestion(question.id)}>×</button>
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