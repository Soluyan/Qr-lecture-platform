import App from './App.svelte';

// Просто создаем приложение без явного указания target
const app = new App({
  target: document.getElementById('app')
});

export default app;