<script>
  import { browser } from '$app/environment';
  import { onMount, onDestroy } from 'svelte';
  import { Chart, registerables } from 'chart.js/auto';
  import { Calendar } from '@fullcalendar/core';
  import dayGridPlugin from '@fullcalendar/daygrid';

  /** @type {any} */
  let portalData = null;
  let isLoading = true;
  let errorMessage = '';

  // Variabel untuk menampung referensi elemen DOM
  // @ts-ignore
  let calendarEl;
  // @ts-ignore
  let chartEl;

  // Variabel untuk instance library
  // @ts-ignore
  let calendarInstance;
  // @ts-ignore
  let chartInstance;
  
  // @ts-ignore
  let currentDate = new Date();

  onMount(async () => {
    if (browser) {
      const token = localStorage.getItem('jwt_token');
      if (!token) {
        window.location.href = '/';
        return;
      }
      
      isLoading = true;
      try {
        const year = new Date().getFullYear();
        const month = (new Date().getMonth() + 1).toString().padStart(2, '0');
        const response = await fetch(`http://localhost:1412/api/portal/dashboard-data/${year}/${month}`, {
          headers: { 'Authorization': 'Bearer ' + token },
          cache: 'no-store'
        });
        if (!response.ok) throw new Error('Gagal memuat data dari server.');
        portalData = await response.json();
      } catch (error) {
        console.error("Fetch Error:", error);
        // @ts-ignore
        errorMessage = error.message;
      } finally {
        isLoading = false;
      }
    }
  });

  // Reactive statement: Ini akan berjalan setiap kali portalData, calendarEl, atau chartEl berubah
 // @ts-ignore
   $: if (portalData && calendarEl && chartEl) {
      // Render Kalender
      if (portalData.events) {
          // @ts-ignore
          if (calendarInstance) calendarInstance.destroy();
          calendarInstance = new Calendar(calendarEl, {
              plugins: [ dayGridPlugin ],
              initialView: 'dayGridMonth',
              events: portalData.events,
              locale: 'id',
              height: 'auto',
          });
          calendarInstance.render();
      }

      // Render Chart
      if (portalData.statistikBulan) {
          // @ts-ignore
          if (chartInstance) chartInstance.destroy();
          Chart.register(...registerables);
          chartInstance = new Chart(chartEl, {
              type: 'bar',
              data: {
                  labels: ['Hadir', 'Izin', 'Sakit', 'Alpa'],
                  datasets: [{ 
                      data: [portalData.statistikBulan.totalHadir, portalData.statistikBulan.totalIzin, portalData.statistikBulan.totalSakit, portalData.statistikBulan.totalAlpa],
                      backgroundColor: ['#198754', '#ffc107', '#fd7e14', '#dc3545'] 
                  }]
              },
              options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false } } }
          });
      }
  }

  onDestroy(() => {
    // @ts-ignore
    if (chartInstance) chartInstance.destroy();
    // @ts-ignore
    if (calendarInstance) calendarInstance.destroy();
  });
</script>

<svelte:head>
  <title>Portal Wali Murid</title>
</svelte:head>

<div class="container py-4">
    {#if isLoading}
        <p class="text-center mt-5">Memuat data portal...</p>
    {:else if portalData && portalData.siswa}
        <div class="mb-4">
            <h3 class="h4">Assalamu'alaikum, {portalData.siswa.namaOrangTua || 'Bapak/Ibu Wali Murid'}!</h3>
            <p class="text-muted">Berikut adalah data kehadiran ananda {portalData.siswa.namaLengkap} (NISN: {portalData.siswa.nisn}).</p>
        </div>
        <div class="row">
            <div class="col-12 col-lg-8 mb-4">
                <div class="card shadow-sm h-100">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <h5 class="mb-0">Kalender Kehadiran</h5>
                        <a href="https://forms.gle/YyiQwU9K31mmcHtAA" target="_blank" class="btn btn-primary btn-sm"><i class="bi bi-pencil-square"></i> Ajukan Izin</a>
                    </div>
                    <div class="card-body">
                        <div bind:this={calendarEl} id="calendar"></div>
                    </div>
                </div>
            </div>
            <div class="col-12 col-lg-4 mb-4">
                <div class="card shadow-sm h-100">
                    <div class="card-header fw-bold">Statistik Bulan Ini ({currentDate.toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })})</div>
                    <div class="card-body d-flex align-items-center justify-content-center">
                        <canvas bind:this={chartEl} id="myChart"></canvas>
                    </div>
                </div>
            </div>
        </div>
    {:else}
        <div class="alert alert-danger">{errorMessage || 'Gagal memuat data portal. Silakan coba lagi.'}</div>
    {/if}
</div>

<style>
    :global(.fc-event-title) { white-space: normal !important; font-size: 0.8em !important; }
</style>