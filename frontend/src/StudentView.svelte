<script>
  import { onMount } from 'svelte';
  let name = '';
  let question = '';
  let isAnonymous = false;
  let sessionId = '';
  let submitted = false;
  let loading = false;
  let error = '';

  onMount(() => {
    const params = new URLSearchParams(window.location.search);
    sessionId = params.get('session');
    if (!sessionId) {
      error = "Session ID not found. Перейдите по QR-коду с лекции.";
    }
  });

  async function submitQuestion() {
    if (!sessionId) return;
    error = '';
    loading = true;
    const author = isAnonymous ? 'Anonymous' : name;
    try {
      const response = await fetch(`http://localhost:8080/ask?session=${sessionId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ author, text: question })
      });

      if (response.ok) {
        submitted = true;
      } else {
        const { message } = await response.json().catch(() => ({}));
        error = message || 'Ошибка отправки. Возможно, сессия устарела или не найдена.';
      }
    } catch (e) {
      error = 'Ошибка сети. Попробуйте еще раз.';
    }
    loading = false;
  }
</script>

<div class="student-container">
  {#if error}
    <div class="error-message">{error}</div>
  {/if}
  {#if !submitted && !error}
    <h2>Ask a Question</h2>
    <form on:submit|preventDefault={submitQuestion}>
      <div class="form-group">
        <label>
          <input type="checkbox" bind:checked={isAnonymous} />
          Ask anonymously
        </label>
      </div>

      {#if !isAnonymous}
        <div class="form-group">
          <label>Your Name:</label>
          <input type="text" bind:value={name} required />
        </div>
      {/if}

      <div class="form-group">
        <label>Your Question:</label>
        <textarea bind:value={question} rows="4" maxlength="500" required />
        <small>{500 - question.length} characters remaining</small>
      </div>

      <button type="submit" disabled={loading}>
        {#if loading}Отправка...{/if}
        {#if !loading}Submit Question{/if}
      </button>
    </form>
  {:else if submitted}
    <div class="success-message">
      <h2>Thank you!</h2>
      <p>Your question has been submitted.</p>
    </div>
  {/if}
</div>

<style>
  .student-container {
    max-width: 600px;
    margin: 2rem auto;
    padding: 2rem;
    border: 1px solid #eee;
    border-radius: 8px;
  }
  
  .form-group {
    margin-bottom: 1rem;
  }
  
  label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: bold;
  }
  
  input, textarea {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
  }
  
  button {
    background: #4CAF50;
    color: white;
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }
  
  .success-message {
    text-align: center;
    color: #4CAF50;
  }

    .error-message {
    color: #c0392b;
    background: #ffecec;
    padding: 1rem;
    margin-bottom: 1rem;
    border-radius: 4px;
    text-align: center;
    font-weight: bold;
  }
</style>