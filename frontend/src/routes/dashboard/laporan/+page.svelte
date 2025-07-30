<script>
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';

  let token = '';
  /** @type {any[]} */
  let rekapList = [];
  let isLoading = false;
  let errorMessage = '';

  // Set tanggal default: awal dan akhir bulan ini
  const today = new Date();
  const y = today.getFullYear();
  const m = today.getMonth();
  const firstDay = new Date(y, m, 1).toISOString().split('T')[0];
  const lastDay = new Date(y, m + 1, 0).toISOString().split('T')[0];

  let startDate = firstDay;
  let endDate = lastDay;

  async function fetchData() {
    if (!token || !startDate || !endDate) return;
    isLoading = true;
    errorMessage = '';
    try {
      const response = await fetch(`http://localhost:1412/api/rekap?mulai=${startDate}&selesai=${endDate}`, {
        headers: { 'Authorization': 'Bearer ' + token },
        cache: 'no-store'
      });
      if (!response.ok) throw new Error('Gagal memuat data rekap.');
      rekapList = await response.json();
    } catch (/** @type {any} */ error) {
      errorMessage = error.message;
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    if (browser) {
      token = localStorage.getItem('jwt_token') || '';
      fetchData();
    }
  });
  
  // Pemicu reaktif untuk mengambil data ulang saat tanggal berubah
  $: if(startDate || endDate) {
      if(browser && token) fetchData();
  }
</script>

<svelte:head><title>Laporan Kehadiran</title></svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Laporan & Rekapitulasi Kehadiran</h1>
</div>

<div class="card shadow-sm">
    <div class="card-body">
        <div class="row g-3 align-items-center mb-3">
            <div class="col-auto"><label for="startDate" class="col-form-label">Dari Tanggal:</label></div>
            <div class="col-auto"><input type="date" class="form-control" id="startDate" bind:value={startDate}></div>
            <div class="col-auto"><label for="endDate" class="col-form-label">Sampai Tanggal:</label></div>
            <div class="col-auto"><input type="date" class="form-control" id="endDate" bind:value={endDate}></div>
        </div>

        <div class="table-responsive">
            <table class="table table-bordered table-hover">
                <thead class="table-light">
                    <tr>
                        <th>NISN</th>
                        <th>Nama Lengkap</th>
                        <th class="text-center">Hadir</th>
                        <th class="text-center">Izin</th>
                        <th class="text-center">Sakit</th>
                        <th class="text-center">Alpa</th>
                    </tr>
                </thead>
                <tbody>
                    {#if isLoading}
                        <tr><td colspan="6" class="text-center">Memuat data...</td></tr>
                    {:else if rekapList.length > 0}
                        {#each rekapList as rekap}
                            <tr>
                                <td>{rekap.nisn}</td>
                                <td>{rekap.namaLengkap}</td>
                                <td class="text-center">{rekap.hadir}</td>
                                <td class="text-center">{rekap.izin}</td>
                                <td class="text-center">{rekap.sakit}</td>
                                <td class="text-center">{rekap.alpa}</td>
                            </tr>
                        {/each}
                    {:else}
                        <tr><td colspan="6" class="text-center">Tidak ada data untuk rentang tanggal yang dipilih.</td></tr>
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>