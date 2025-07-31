<script>
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  /** @type {any} */
  let rekapData = null;
  let isLoading = true;
  let errorMessage = '';
  let token = '';
  let selectedDate = new Date().toISOString().split('T')[0];

  async function fetchRekap() {
    if (!token || !selectedDate) return;

    isLoading = true;
    errorMessage = '';
    try {
      const apiUrl = import.meta.env.VITE_API_BASE_URL;
      const response = await fetch(`${apiUrl}/api/absensi/rekap/${selectedDate}?t=${new Date().getTime()}`, {
        headers: { 'Authorization': 'Bearer ' + token },
        cache: 'no-store'
      });
      if (!response.ok) throw new Error('Gagal memuat data rekap.');
      rekapData = await response.json();
    } catch (/** @type {any} */ error) {
      errorMessage = error.message;
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    if (browser) {
      token = localStorage.getItem('jwt_token') || '';
    }
  });

  // --- INI PERBAIKAN KUNCINYA ---
  // Baris ini secara eksplisit memberitahu Svelte untuk menjalankan fetchRekap()
  // setiap kali nilai 'selectedDate' berubah (setelah token ada).
  $: if (browser && token && selectedDate) {
      fetchRekap();
  }
</script>

<svelte:head>
    <title>Manajemen Kehadiran</title>
</svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Rekap & Manajemen Kehadiran</h1>
    <div class="btn-toolbar mb-2 mb-md-0">
        <a href="/dashboard/kehadiran/tambah" class="btn btn-sm btn-primary">
            <i class="bi bi-plus-circle"></i> Catat Absensi Massal
        </a>
    </div>
</div>

<div class="row">
    <div class="col-md-4 mb-3">
        <label for="tanggal" class="form-label fw-bold">Pilih Tanggal:</label>
        <input type="date" class="form-control" id="tanggal" bind:value={selectedDate}>
    </div>
</div>


{#if errorMessage}
    <div class="alert alert-danger">{errorMessage}</div>
{/if}

{#if isLoading}
    <p>Memuat data untuk tanggal {selectedDate}...</p>
{:else if rekapData}
    {#if rekapData.isHoliday}
        <div class="alert alert-info text-center mt-3">
            <h4 class="alert-heading"><i class="bi bi-calendar-heart"></i> Hari Libur</h4>
            <p>Tanggal {selectedDate} adalah {rekapData.holidayDescription}. Tidak ada data absensi.</p>
        </div>
    {:else}
        <div class="card shadow-sm">
            <div class="card-body">
                <div class="table-responsive">
                    <table class="table table-striped table-hover">
                        <thead>
                            <tr>
                                <th>NISN</th>
                                <th>Nama Lengkap</th>
                                <th>Status</th>
                                <th>Keterangan</th>
                                <th class="text-center">Aksi</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each rekapData.daftarStatusSiswa as siswa}
                                <tr>
                                    <td>{siswa.nisn}</td>
                                    <td>{siswa.namaLengkap}</td>
                                    <td>
                                        {#if siswa.status === 'Hadir'} <span class="badge bg-success">{siswa.status}</span>
                                        {:else if siswa.status === 'Sakit' || siswa.status === 'Izin'} <span class="badge bg-warning text-dark">{siswa.status}</span>
                                        {:else if siswa.status === 'Alpa'} <span class="badge bg-danger">{siswa.status}</span>
                                        {:else} <span class="badge bg-secondary">{siswa.status}</span> {/if}
                                    </td>
                                    <td>{siswa.keterangan}</td>
                                    <td class="text-center">
                                        <!-- svelte-ignore a11y_consider_explicit_label -->
                                        <button class="btn btn-sm btn-warning" disabled={!siswa.rowNumber} on:click={() => goto(`/dashboard/kehadiran/edit/${siswa.rowNumber}`)}><i class="bi bi-pencil-square"></i></button>
                                        <!-- svelte-ignore a11y_consider_explicit_label -->
                                        <button class="btn btn-sm btn-danger" disabled={!siswa.rowNumber}><i class="bi bi-trash"></i></button>
                                    </td>
                                </tr>
                            {:else}
                                <tr><td colspan="5" class="text-center">Tidak ada data untuk tanggal yang dipilih.</td></tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    {/if}
{/if}