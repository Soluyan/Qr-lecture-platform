<script>
  import { onMount } from 'svelte';
  let sessionId = '';
  let qrCodeUrl = '';
  let questions = [];
  let ws;

  async function createSession() {
    const response = await fetch('http://localhost:8080/create-session');
    const qrBlob = await response.blob();
    qrCodeUrl = URL.createObjectURL(qrBlob);
    sessionId = new URL(response.url).searchParams.get('session');
    connectWebSocket();
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
    {#if qrCodeUrl}
      <img src={qrCodeUrl} alt="Lecture QR Code" class="qr-code" />
      <p>Session ID: {sessionId}</p>
    {:else}
      <button on:click={createSession}>Start New Lecture</button>
    {/if}
  </div>

  <div class="questions-section">
    <h2>Student Questions</h2>
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