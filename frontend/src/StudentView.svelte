<script>
  import { onMount } from 'svelte';
  let name = '';
  let question = '';
  let isAnonymous = false;
  let sessionId = '';
  let submitted = false;

  onMount(() => {
    const params = new URLSearchParams(window.location.search);
    sessionId = params.get('session');
  });

  async function submitQuestion() {
    const author = isAnonymous ? 'Anonymous' : name;
    
    const response = await fetch(`http://localhost:8080/ask?session=${sessionId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ author, text: question })
    });

    if (response.ok) {
      submitted = true;
    }
  }
</script>

<div class="student-container">
  {#if !submitted}
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

      <button type="submit">Submit Question</button>
    </form>
  {:else}
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
</style>