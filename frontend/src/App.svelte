<script>
  import svelteLogo from './assets/svelte.svg'
  import viteLogo from '/vite.svg'
  import Counter from './lib/Counter.svelte'

  import TeacherView from './TeacherView.svelte';
  import StudentView from './StudentView.svelte';
  
  const isTeacherView = !window.location.pathname.includes('/ask');
</script>

<main>
  {#if isTeacherView}
    <TeacherView />
  {:else}
    <StudentView />
  {/if}
</main>

<style>
  main {
    font-family: Arial, sans-serif;
    line-height: 1.6;
  }
</style>
