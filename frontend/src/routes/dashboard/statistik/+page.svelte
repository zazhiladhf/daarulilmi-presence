<script>
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';
  import { Chart, registerables } from 'chart.js/auto';

  let token = '';
  // @ts-ignore
  let chartInstance = null;
  
  // Opsi untuk bulan dan tahun
  const months = [
    { value: 7, name: 'Juli' }, { value: 8, name: 'Agustus' }, // ... tambahkan bulan lain
  ];
  const years = [2025, 2026];

  let selectedMonth = new Date().getMonth() + 1;
  let selectedYear = new Date().getFullYear();
  
  async function fetchData() {
    if (!token) return;
    try {
      const response = await fetch(`http://localhost:1412/api/statistik/bulanan/${selectedYear}/${selectedMonth}`, {
        headers: { 'Authorization': 'Bearer ' + token },
        cache: 'no-store'
      });
      if (!response.ok) throw new Error('Gagal memuat data statistik.');
      const data = await response.json();
      updateChart(data);
    } catch (/**@type{any}*/error) {
      alert(error.message);
    }
  }

  // @ts-ignore
  function updateChart(data) {
    const chartData = {
      labels: ['Hadir', 'Izin', 'Sakit', 'Alpa'],
      datasets: [{
        label: 'Jumlah Kehadiran',
        data: [data.totalHadir, data.totalIzin, data.totalSakit, data.totalAlpa],
        backgroundColor: ['#198754', '#ffc107', '#fd7e14', '#dc3545'],
      }]
    };

    // @ts-ignore
    if (chartInstance) {
      chartInstance.data = chartData;
      chartInstance.update();
    } else {
      const ctx = document.getElementById('myChart');
      if (ctx) {
        Chart.register(...registerables);
        // @ts-ignore
        chartInstance = new Chart(ctx, {
          type: 'bar', // Tipe grafik: bar, pie, line, dll.
          data: chartData,
          options: { responsive: true, plugins: { legend: { display: false } } }
        });
      }
    }
  }

  onMount(() => {
    if (browser) {
      token = localStorage.getItem('jwt_token') || '';
      fetchData();
    }
  });

  // Pemicu reaktif
 // @ts-ignore
   $: if (selectedMonth || selectedYear) {
      if (browser) fetchData();
  }
</script>

<svelte:head><title>Statistik Kehadiran</title></svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Statistik Kehadiran</h1>
</div>

<div class="card shadow-sm">
    <div class="card-body">
        <div class="row">
            <div class="col-md-3">
                <label for="year" class="form-label">Tahun</label>
                <select class="form-select" id="year" bind:value={selectedYear}>
                    {#each years as year}<option value={year}>{year}</option>{/each}
                </select>
            </div>
            <div class="col-md-3">
                <label for="month" class="form-label">Bulan</label>
                <select class="form-select" id="month" bind:value={selectedMonth}>
                    {#each months as month}<option value={month.value}>{month.name}</option>{/each}
                </select>
            </div>
        </div>
        <hr>
        <div style="max-height: 400px;">
            <canvas id="myChart"></canvas>
        </div>
    </div>
</div>