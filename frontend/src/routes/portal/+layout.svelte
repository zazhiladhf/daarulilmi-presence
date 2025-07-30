<script>
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import '$lib/app.css';

  let userInitials = '';
  let schoolName = 'SMA Islam Daarul Ilmi';

  onMount(() => {
    if (browser) {
      const token = localStorage.getItem('jwt_token');
      if (!token) {
        goto('/'); // Jika tidak ada token, usir ke halaman login
        return;
      }
      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        userInitials = payload.username.charAt(0).toUpperCase();
      } catch (e) {
        localStorage.removeItem('jwt_token'); // Hapus token rusak
        goto('/');
      }
    }
  });

  function logout() {
    if (browser) {
      localStorage.removeItem('jwt_token');
      goto('/');
    }
  }
</script>

<div class="d-flex flex-column" style="min-height: 100vh;">
  <header class="navbar navbar-dark bg-success sticky-top shadow-sm p-2">
    <div class="container-fluid">
        <span class="navbar-brand mb-0 h1">{schoolName}</span>
        <div class="dropdown">
            <button class="btn btn-light rounded-circle fw-bold" data-bs-toggle="dropdown" style="width: 40px; height: 40px;">
                {userInitials}
            </button>
            <ul class="dropdown-menu dropdown-menu-end">
                <li><button class="dropdown-item" on:click={logout}>Logout</button></li>
            </ul>
        </div>
    </div>
  </header>

  <main class="flex-grow-1">
    <slot />
  </main>

  <footer class="footer py-3 bg-white border-top fixed-bottom">
    <div class="container text-center">
      <span class="text-muted">
        Made with <i class="bi bi-heart-fill text-danger"></i> by <a href="https://zazhil.my.id" target="_blank" class="fw-bold text-decoration-none">Zazhil Adhafi</a>
      </span>
    </div>
  </footer>
</div>