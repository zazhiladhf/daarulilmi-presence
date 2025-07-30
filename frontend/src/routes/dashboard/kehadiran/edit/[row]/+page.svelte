<script>
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  // @ts-ignore
  import { goto, invalidateAll } from '$app/navigation';
  import { onMount } from 'svelte';

  const rowNumber = $page.params.row;
  let token = '';

  /** @type {any} */
  let kehadiran = null;
  let tanggal = '';
  let waktu = '';
  
  let isLoading = true;
  let isSaving = false;
  let errorMessage = '';
  let successMessage = '';

  onMount(async () => {
    if (browser) {
      token = localStorage.getItem('jwt_token') || '';
      try {
        const response = await fetch(`http://localhost:1412/api/absensi/log/${rowNumber}`, {
          headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!response.ok) throw new Error('Gagal mengambil data absensi.');
        kehadiran = await response.json();

        // --- PERBAIKAN DI SINI ---
        // Gunakan nama field JSON (huruf kecil), dan tambahkan pengecekan
        if (kehadiran && kehadiran.timestamp) {
            const [tgl, wkt] = kehadiran.timestamp.split(' ');
            tanggal = tgl;
            waktu = wkt ? wkt.substring(0, 5) : '07:00';
        }

      } catch (/**@type {any}*/error) {
        errorMessage = error.message;
      } finally {
        isLoading = false;
      }
    }
  });

  async function handleUpdate() {
    isSaving = true;
    errorMessage = '';
    successMessage = '';
    
    const updatedTimestamp = `${tanggal} ${waktu}:00`;

    try {
      const response = await fetch(`http://localhost:1412/api/absensi/log/${rowNumber}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + token
        },
        body: JSON.stringify({
            NISN: kehadiran.username,
            Status: kehadiran.status,
            Timestamp: updatedTimestamp
        })
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.message);
      
      successMessage = data.message;
      
      // --- PERBAIKAN KUNCI DI SINI ---
      // 1. Beritahu SvelteKit untuk membuang cache
      // @ts-ignore
      await invalidateAll();
      
      // 2. Arahkan kembali setelah cache dibersihkan
      setTimeout(() => {
        goto('/dashboard/kehadiran');
      }, 1000); // Beri jeda 1 detik agar pesan sukses terbaca

    } catch (/**@type {any}*/error) {
      errorMessage = error.message;
    } finally {
      isSaving = false;
    }
  }

</script>

<svelte:head><title>Edit Kehadiran</title></svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Edit Kehadiran Manual</h1>
    <a href="/dashboard/kehadiran" class="btn btn-sm btn-outline-secondary"><i class="bi bi-arrow-left"></i> Kembali</a>
</div>

{#if isLoading}
    <p>Memuat data...</p>
{:else if kehadiran}
    <div class="row">
        <div class="col-lg-8">
            <div class="card shadow-sm">
                <div class="card-body">
                    {#if errorMessage}<div class="alert alert-danger">{errorMessage}</div>{/if}
                    {#if successMessage}<div class="alert alert-success">{successMessage}</div>{/if}
                    
                    <form on:submit|preventDefault={handleUpdate}>
                        <div class="mb-3">
                            <label for="nisn" class="form-label">NISN Siswa</label>
                            <input type="text" class="form-control" id="nisn" bind:value={kehadiran.username} required>
                        </div>
                        <div class="row">
                            <div class="col-md-6 mb-3">
                                <label for="tanggal" class="form-label">Tanggal</label>
                                <input type="date" class="form-control" id="tanggal" bind:value={tanggal} required>
                            </div>
                            <div class="col-md-6 mb-3">
                                <label for="waktu" class="form-label">Waktu (Jam:Menit)</label>
                                <input type="time" class="form-control" id="waktu" bind:value={waktu} required>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label for="status" class="form-label">Status Kehadiran</label>
                            <select class="form-select" id="status" bind:value={kehadiran.status} required>
                                <option value="Hadir">Hadir</option>
                                <option value="Sakit">Sakit</option>
                                <option value="Izin">Izin</option>
                                <option value="Alpa">Alpa</option>
                            </select>
                        </div>
                        <button type="submit" class="btn btn-primary" disabled={isSaving}>
                            {#if isSaving}
                                <span class="spinner-border spinner-border-sm"></span> Menyimpan...
                            {:else}
                                <i class="bi bi-save"></i> Simpan Perubahan
                            {/if}
                        </button>
                    </form>
                </div>
            </div>
        </div>
    </div>
{:else}
     <div class="alert alert-danger">{errorMessage || 'Data absensi tidak ditemukan.'}</div>
{/if}