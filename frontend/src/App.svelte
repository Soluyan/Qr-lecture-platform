<script>
  import TeacherView from "./TeacherView.svelte";
  import StudentView from "./StudentView.svelte";

  let isTeacherView = true;

  // Проверяем, есть ли параметры в URL
  if (typeof window !== "undefined") {
    const url = new URL(window.location.href);
    // Если путь содержит "student" или есть параметр session - это студент
    isTeacherView =
      !url.pathname.includes("student") && !url.searchParams.has("session");
  }
</script>

<main>
  {#if isTeacherView}
    <TeacherView />
  {:else}
    <StudentView />
  {/if}
</main>
