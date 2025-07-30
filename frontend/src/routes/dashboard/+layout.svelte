<!-- file: src/routes/dashboard/+layout.svelte (Final Berdasarkan Versi Anda) -->
<script>
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import '$lib/app.css';

  let namaLengkap = '...';
  let userInitials = '';
  let schoolName = 'SMA Islam Daarul Ilmi';
  let isSidebarVisible = false;

  onMount(async () => {
    if (browser) {
      const token = localStorage.getItem('jwt_token');
      if (!token) { goto('/'); return; }
      
      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        const username = payload.username;
        userInitials = username.charAt(0).toUpperCase();

        const response = await fetch(`http://localhost:1412/api/user/profile`, {
          headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!response.ok) throw new Error('Gagal mengambil profil.');
        
        const userData = await response.json();
        namaLengkap = userData.namaLengkap || username;

      } catch (e) {
        console.error("Gagal memproses token atau profil:", e);
        localStorage.removeItem('jwt_token');
        goto('/');
      }
    }
  });

  function toggleSidebar() {
    isSidebarVisible = !isSidebarVisible;
  }
  
  function handleMenuClick() {
    if (window.innerWidth < 992) {
      isSidebarVisible = false;
    }
  }

  function logout() {
    if (browser) {
      localStorage.removeItem('jwt_token');
      goto('/');
    }
  }
</script>

<svelte:head>
  <title>{$page.route.id?.split('/').pop()?.replace(/\[|\]/g, '')?.replace(/-/g, ' ')?.replace(/\b\w/g, l => l.toUpperCase()) || 'Dashboard'} - {schoolName}</title>
</svelte:head>

<div class="admin-layout">
  <nav class="sidebar" class:show={isSidebarVisible}>
    <div class="sidebar-header text-center">
      <a href="/dashboard" class="sidebar-brand text-dark text-decoration-none">
        <i class="bi bi-mortarboard-fill fs-3 align-middle"></i>
        <span class="fs-5 fw-bold align-middle ms-2">{schoolName}</span>
      </a>
    </div>
    <ul class="nav nav-pills flex-column mb-auto">
        <li class="nav-item"><a href="/dashboard" class="nav-link" class:active={$page.url.pathname === '/dashboard'} on:click={handleMenuClick}><i class="bi bi-speedometer2"></i> Dashboard</a></li>
        <li><a href="/dashboard/siswa" class="nav-link" class:active={$page.url.pathname.startsWith('/dashboard/siswa')} on:click={handleMenuClick}><i class="bi bi-people"></i> Manajemen Siswa</a></li>
        <li><a href="/dashboard/izin" class="nav-link" class:active={$page.url.pathname === '/dashboard/izin'} on:click={handleMenuClick}><i class="bi bi-journal-check"></i> Kelola Izin</a></li>
        <li><a href="/dashboard/kehadiran" class="nav-link" class:active={$page.url.pathname.startsWith('/dashboard/kehadiran')} on:click={handleMenuClick}><i class="bi bi-calendar-event"></i> Kehadiran Manual</a></li>
        <li><a href="/dashboard/laporan" class="nav-link" class:active={$page.url.pathname === '/dashboard/laporan'} on:click={handleMenuClick}><i class="bi bi-file-earmark-bar-graph"></i> Laporan</a></li>
        <li><a href="/dashboard/statistik" class="nav-link" class:active={$page.url.pathname === '/dashboard/statistik'} on:click={handleMenuClick}><i class="bi bi-bar-chart-line"></i> Statistik</a></li>
        <li><a href="/dashboard/qr-generator" class="nav-link" class:active={$page.url.pathname === '/dashboard/qr-generator'} on:click={handleMenuClick}><i class="bi bi-qr-code"></i> QR Generator</a></li>
    </ul>
    <hr>
    <div class="sidebar-footer dropdown">
      <!-- svelte-ignore a11y_invalid_attribute -->
      <a href="javascript:void(0);" class="d-flex align-items-center link-dark text-decoration-none dropdown-toggle" data-bs-toggle="dropdown" aria-expanded="false">
          <i class="bi bi-person-circle fs-2 me-2"></i>
          <span class="fw-semibold">{namaLengkap}</span>
      </a>
      <ul class="dropdown-menu text-small shadow">
          <!-- svelte-ignore a11y_invalid_attribute -->
          <li><a class="dropdown-item" href="#">Pengaturan</a></li>
          <li><hr class="dropdown-divider"></li>
          <li><button class="dropdown-item" on:click={logout}>Logout</button></li>
      </ul>
    </div>
  </nav>

  <div class="main-content">
    <header class="main-header">
      <!-- svelte-ignore a11y_consider_explicit_label -->
      <button class="btn btn-light d-lg-none" on:click={toggleSidebar}>
        <i class="bi bi-list"></i>
      </button>
      <div class="fw-bold d-lg-none mx-auto">{schoolName}</div>
      <div class="ms-auto dropdown">
        <!-- svelte-ignore a11y_invalid_attribute -->
        <a href="javascript:void(0);" class="d-flex align-items-center link-dark text-decoration-none dropdown-toggle" data-bs-toggle="dropdown" aria-expanded="false">
          <strong class="me-2 d-none d-sm-block">{namaLengkap}</strong>
          <div class="bg-primary text-white rounded-circle d-flex align-items-center justify-content-center" style="width: 40px; height: 40px;">
              <b>{userInitials}</b>
          </div>
        </a>
        <ul class="dropdown-menu dropdown-menu-end text-small shadow">
          <!-- svelte-ignore a11y_invalid_attribute -->
          <li><a class="dropdown-item" href="#">Pengaturan</a></li>
          <li><hr class="dropdown-divider"></li>
          <li><button class="dropdown-item" on:click={logout}>Logout</button></li>
        </ul>
      </div>
    </header>
    
    <div class="page-content">
      <slot />
    </div>

    <footer class="main-footer">
        Made with <i class="bi bi-heart-fill text-danger"></i> by <a href="https://zazhil.my.id" target="_blank" class="fw-bold text-decoration-none">Zazhil Adhafi</a>
    </footer>
  </div>

  {#if isSidebarVisible}
    <div class="sidebar-overlay d-lg-none" on:click={toggleSidebar} role="button" tabindex="0" on:keypress on:keydown={()=>{}} ></div>
  {/if}
</div>