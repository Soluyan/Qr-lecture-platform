<script>
  import { onMount } from "svelte";

  let name = "";
  let question = "";
  let isAnonymous = false;
  let sessionId = "";
  let submitted = false;
  let loading = false;
  let error = "";
  let apiBaseUrl = "http://localhost:8080";
  let sessionSettings = null;
  let settingsLoading = false;

  onMount(() => {
    const urlParams = new URLSearchParams(window.location.search);
    sessionId = urlParams.get("session") || "";

    if (!sessionId) {
      error = "Идентификатор сеанса не найден. Используйте QR-код из лекции.";
      console.error("No session ID found in URL");
    } else {
      console.log("Session ID detected:", sessionId);
      loadSessionSettings();
    }
  });

  async function loadSessionSettings() {
    if (!sessionId) return;

    settingsLoading = true;
    try {
      const response = await fetch(
        `${apiBaseUrl}/session/settings/get?session=${sessionId}`
      );
      if (response.ok) {
        sessionSettings = await response.json();
        console.log("Session settings loaded:", sessionSettings);
      }
    } catch (error) {
      console.error("Error loading settings:", error);
    } finally {
      settingsLoading = false;
    }
  }

  async function submitQuestion() {
    if (!sessionId) {
      error = "Требуется идентификатор сеанса";
      return;
    }

    // Проверяем настройки сессии
    if (sessionSettings && !sessionSettings.allowAnonymous && isAnonymous) {
      error = "Анонимные вопросы отключены для этой сессии";
      return;
    }
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

    error = "";
    loading = true;

    const author = isAnonymous ? "Аноним" : name.trim() || "Аноним";

    try {
      const response = await fetch(`${apiBaseUrl}/ask?session=${sessionId}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({
          author: author.substring(0, 100), // Ограничиваем длину имени
          text: question.trim(),
        }),
      });

      console.log("Response status:", response.status);

      if (response.ok) {
        submitted = true;
        // Сбрасываем форму
        question = "";
        name = "";
      } else {
        let errorMessage = "Error submitting question";

        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
        } catch (parseError) {
          errorMessage = `Server error: ${response.status} ${response.statusText}`;
        }

        error = errorMessage;

        // Если сессия не найдена
        if (response.status === 404) {
          error = "Сеанс не найден или истёк. Проверьте QR-код.";
        }
      }
    } catch (e) {
      console.error("Network error:", e);
      error = "Ошибка сети. Проверьте подключение и повторите попытку.";
    } finally {
      loading = false;
    }
  }

  // Функция для отправки другого вопроса
  function submitAnother() {
    submitted = false;
    question = "";
    error = "";
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
      <strong>Ошибка:</strong>
      {error}
      {#if error.includes("session")}
        <p>Пожалуйста, обратитесь к лектору за действительным QR-кодом.</p>
      {/if}
    </div>
  {/if}

  {#if !submitted}
    <div class="form-section">
      <h2>Задайте вопрос</h2>
      <p class="session-info" class:hidden={!sessionId}>
        <strong>ID сессии:</strong>
        {sessionId ? sessionId.substring(0, 18) + "..." : "Not detected"}
      </p>

      {#if settingsLoading}
        <p>Загрузка настроек сессии...</p>
      {:else if sessionSettings && !sessionSettings.allowAnonymous}
        <div class="info-message">
          ⓘ Анонимные вопросы отключены для этой лекции
        </div>
      {/if}

      <form on:submit|preventDefault={submitQuestion} class="question-form">
        {#if sessionSettings && sessionSettings.allowAnonymous}
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
        {/if}

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
              required
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
          class:loading
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
    min-height: 100vh;
    padding: 0;
    background: white;
    font-family: "Inter", sans-serif;
    color: #2d3748;
    font-weight: 400;
    font-size: 16px;
    line-height: 1.6;
  }

  .form-section h2 {
    color: #0078cf;
    font-weight: 700;
    text-align: center;
    margin-bottom: 2.5rem;
    position: relative;
  }

  .form-section h2::after {
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

  .form-section {
    padding: 2rem 1rem;
    max-width: 500px;
    margin: 0 auto;
    animation: formEnter 0.6s cubic-bezier(0.23, 1, 0.32, 1);
  }

  @keyframes formEnter {
    from {
      opacity: 0;
      transform: translateY(30px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  @media (min-width: 768px) {
    .form-section {
      padding: 3rem 2rem;
      margin: 2rem auto;
      border: 1px solid #e2e8f0;
      border-radius: 12px;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
      font-size: 1.05rem;
    }

    .student-container {
      background: #f8fafc;
      padding: 1rem;
    }
  }

  @media (max-width: 480px) {
    .form-section {
      padding: 1.5rem 1rem;
      margin: 1rem;
    }

    .form-group {
      padding: 1rem;
    }

    button {
      padding: 1rem 1.5rem;
      font-size: 1.1rem; /* Увеличенный размер для мобильных */
    }
  }

  .form-group {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.25rem;
    margin-bottom: 1rem;
    transition: all 0.3s ease;
    position: relative;
    width: 100%;
    box-sizing: border-box;
  }

  .form-group::after {
    content: "";
    position: absolute;
    bottom: 0;
    left: 0;
    width: 0;
    height: 2px;
    background: #48bb78;
    transition: width 0.3s ease;
  }

  .form-group:focus-within {
    border-color: #48bb78;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(72, 187, 120, 0.15);
  }

  .form-group:focus-within::after {
    width: 100%;
  }

  .form-group label {
    color: #4a5568;
    font-weight: 600;
    margin-bottom: 0.75rem;
    display: block;
    cursor: pointer;
  }

  .form-group textarea {
    border: none;
    padding: 0;
    font-size: 1rem;
    background: transparent;
    width: 100%;
    resize: vertical;
    min-height: 100px;
    max-height: 300px;
    line-height: 1.5;
    font-family: inherit;
  }

  .info-message {
    background: #e3f2fd;
    border: 1px solid #2196f3;
    color: #1976d2;
    padding: 0.75rem;
    border-radius: 4px;
    margin-bottom: 1rem;
    font-size: 0.9rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: bold;
  }

  textarea {
    word-wrap: break-word;
    overflow-wrap: break-word;
    white-space: pre-wrap;
  }

  input,
  textarea {
    border: none;
    padding: 0;
    font-size: 1rem;
    background: transparent;
    width: 100%;
    resize: none;
  }

  input:focus,
  textarea:focus {
    outline: none;
    box-shadow: none;
    transform: none;
  }

  input:invalid,
  textarea:invalid {
    border-color: #e53e3e;
  }

  input:valid,
  textarea:valid {
    border-color: #48bb78;
  }

  button {
    background: #48bb78;
    color: white;
    border: none;
    padding: 0.875rem 1.5rem;
    border-radius: 6px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  button:hover:not(:disabled) {
    background: #38a169;
  }

  button:focus {
    outline: 3px solid #48bb78;
    outline-offset: 2px;
    box-shadow: 0 0 0 4px rgba(72, 187, 120, 0.2);
  }

  button:disabled {
    background: #a0aec0;
    cursor: not-allowed;
    transform: none !important;
  }

  button:disabled:hover {
    background: #a0aec0;
  }

  .success-message {
    text-align: center;
    color: #48bb78;
    animation: successEnter 0.8s cubic-bezier(0.23, 1, 0.32, 1);
  }

  @keyframes successEnter {
    0% {
      opacity: 0;
      transform: scale(0.8) rotateX(-90deg);
    }
    50% {
      transform: scale(1.05) rotateX(0);
    }
    100% {
      opacity: 1;
      transform: scale(1) rotateX(0);
    }
  }

  .success-icon {
    animation: iconBounce 0.6s ease;
  }

  @keyframes iconBounce {
    0%,
    20%,
    50%,
    80%,
    100% {
      transform: scale(1);
    }
    40% {
      transform: scale(1.3);
    }
    60% {
      transform: scale(1.1);
    }
  }

  .loading {
    position: relative;
    color: transparent;
  }

  .loading::after {
    content: "";
    width: 20px;
    height: 20px;
    border: 2px solid transparent;
    border-top: 2px solid white;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    display: inline-block;
    margin-left: 10px;
  }

  .loading::before {
    content: "";
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 20px;
    height: 20px;
    border: 2px solid transparent;
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  .char-counter {
    text-align: right;
    font-size: 0.8rem;
    color: #a0aec0;
    margin-top: 0.5rem;
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
