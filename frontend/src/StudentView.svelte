<script>
  import { onMount } from 'svelte';

  let name = '';
  let question = '';
  let isAnonymous = false;
  let sessionId = '';
  let submitted = false;
  let loading = false;
  let error = '';
  let apiBaseUrl = 'http://localhost:8080';

  onMount(() => {
    const urlParams = new URLSearchParams(window.location.search);
    sessionId = urlParams.get('session') || '';
    
    if (!sessionId) {
      error = "Идентификатор сеанса не найден. Используйте QR-код из лекции.";
      console.error('No session ID found in URL');
    } else {
      console.log('Session ID detected:', sessionId);
    }
  });

  async function submitQuestion() {
    if (!sessionId) {
      error = "Требуется идентификатор сеанса";
      return;
    }
    
    // Валидация
    if (!question.trim()) {
      error = "Пожалуйста, введите ваш вопрос";
      return;
    }
    
    if (question.length > 500) {
      error = "Вопрос слишком длинный (максимум 500 символов)";
      return;
    }

    error = '';
    loading = true;
    
    const author = isAnonymous ? 'Аноним' : (name.trim() || 'Аноним');

    try {
      const response = await fetch(`${apiBaseUrl}/ask?session=${sessionId}`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({ 
          author: author.substring(0, 100), // Ограничиваем длину имени
          text: question.trim() 
        })
      });

      console.log('Response status:', response.status);
      
      if (response.ok) {
        submitted = true;
        // Сбрасываем форму
        question = '';
        name = '';
      } else {
        let errorMessage = 'Error submitting question';
        
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
        } catch (parseError) {
          errorMessage = `Server error: ${response.status} ${response.statusText}`;
        }
        
        error = errorMessage;
        
        // Если сессия не найдена, показываем особое сообщение
        if (response.status === 404) {
          error = "Сеанс не найден или истёк. Проверьте QR-код.";
        }
      }
    } catch (e) {
      console.error('Network error:', e);
      error = 'Ошибка сети. Проверьте подключение и повторите попытку.';
    } finally {
      loading = false;
    }
  }

  // Функция для отправки другого вопроса
  function submitAnother() {
    submitted = false;
    question = '';
    error = '';
  }

  // Автоматически фокусироваться на поле вопроса при загрузке
  let questionTextarea;
  onMount(() => {
    if (questionTextarea) {
      questionTextarea.focus();
    }
  });
</script>

<div class="student-container">
  {#if error}
    <div class="error-message" role="alert">
      <strong>Error:</strong> {error}
      {#if error.includes('session')}
        <p>Пожалуйста, обратитесь к лектору за действительным QR-кодом.</p>
      {/if}
    </div>
  {/if}
  
  {#if !submitted}
    <div class="form-section">
      <h2>Задайте вопрос</h2>
      <p class="session-info" class:hidden={!sessionId}>
        <strong>ID сессии:</strong> {sessionId ? sessionId.substring(0, 18) + '...' : 'Not detected'}
      </p>

      <form on:submit|preventDefault={submitQuestion} class="question-form">
        <div class="form-group">
          <label class="checkbox-label">
            <input 
              type="checkbox" 
              bind:checked={isAnonymous}
              aria-label="Ask anonymously"
            />
            <span class="checkmark"></span>
            Анонимно
          </label>
        </div>

        {#if !isAnonymous}
          <div class="form-group">
            <label for="name">Ваше Имя:</label>
            <input 
              id="name"
              type="text" 
              bind:value={name} 
              placeholder="Введите ваше имя"
              maxlength="100"
              disabled={loading}
            />
          </div>
        {/if}

        <div class="form-group">
          <label for="question">Ваш Вопрос:</label>
          <textarea 
            id="question"
            bind:value={question} 
            bind:this={questionTextarea}
            rows="4" 
            maxlength="500" 
            placeholder="Введите ваш вопрос здесь..."
            disabled={loading}
            required
          />
          <div class="char-counter">
            {500 - question.length} символов осталось
          </div>
        </div>

        <button 
          type="submit" 
          disabled={loading || !question.trim()}
          class:loading={loading}
        >
          {#if loading}
            <span class="spinner"></span>
            Sending...
          {:else}
            Отправить Вопрос
          {/if}
        </button>
      </form>
    </div>
  {:else}
    <div class="success-message" role="status">
      <div class="success-icon">✓</div>
      <h2>Спасибо!</h2>
      <p>Ваш вопрос был передан.</p>
      <button on:click={submitAnother} class="another-question">
        Задать еще один вопрос
      </button>
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