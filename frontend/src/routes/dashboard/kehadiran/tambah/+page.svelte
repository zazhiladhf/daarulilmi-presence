<script>
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  let token = '';
  /** @type {any[]} */
  let siswaList = [];
  /** @type {Object<string, string>} */
  let attendanceStatus = {};

  let tanggal = new Date().toISOString().split('T')[0];
  let isLoading = true;
  let isSaving = false;
  let errorMessage = '';
  let successMessage = '';

  onMount(async () => {
    if (browser) {
      token = localStorage.getItem('jwt_token') || '';
      isLoading = true;
      errorMessage = '';
      try {
        console.log("Mencoba mengambil daftar siswa..."); // Log Debug 1
        const apiUrl = import.meta.env.VITE_API_BASE_URL;
        const response = await fetch(`${apiUrl}/api/siswa`, {
          headers: { 'Authorization': 'Bearer ' + token }
        });

        if (!response.ok) {
          throw new Error('Gagal mengambil daftar siswa dari server.');
        }
        
        const data = await response.json();
        console.log("Data siswa diterima dari backend:", data); // Log Debug 2: Lihat apa isinya

        // Filter data yang tidak valid di frontend sebagai lapisan pengaman terakhir
        // @ts-ignore
        siswaList = data.filter(siswa => siswa && siswa.nisn);
        
        // Inisialisasi status default
        siswaList.forEach(siswa => {
            attendanceStatus[siswa.nisn] = 'Hadir';
        });

      } catch (/** @type {any} */ error) {
        console.error("Fetch Error:", error);
        errorMessage = error.message;
      } finally {
        isLoading = false;
      }
    }
  });

  async function handleSubmit() {
    isSaving = true;
    errorMessage = '';
    successMessage = '';

    // Ubah data dari form menjadi array yang akan dikirim ke API
    const payload = Object.entries(attendanceStatus).map(([nisn, status]) => {
        return {
            NISN: nisn,
            Status: status,
            // Gabungkan tanggal yang dipilih dengan jam default, misal jam 7 pagi
            Timestamp: `${tanggal} 07:00:00` 
        };
    });

    try {
      const apiUrl = import.meta.env.VITE_API_BASE_URL;
      const response = await fetch(`${apiUrl}/api/absensi/manual/batch`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + token
        },
        body: JSON.stringify(payload)
      });

      const data = await response.json();
      if (!response.ok) throw new Error(data.message);
      
      successMessage = data.message;
      // Tunggu 2 detik lalu arahkan kembali ke halaman rekap
      setTimeout(() => {
        goto('/dashboard/kehadiran');
      }, 2000);

    } catch (/** @type {any} */ error) {
      errorMessage = error.message;
    } finally {
      isSaving = false;
    }
  }
</script>

<svelte:head>
  <title>Catat Kehadiran Massal</title>
</svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Catat Kehadiran Massal</h1>
    <a href="/dashboard/kehadiran" class="btn btn-sm btn-outline-secondary"><i class="bi bi-arrow-left"></i> Kembali</a>
</div>

<div class="card shadow-sm">
    <div class="card-body">
        {#if errorMessage}<div class="alert alert-danger">{errorMessage}</div>{/if}
        {#if successMessage}<div class="alert alert-success">{successMessage}</div>{/if}
        
        <form on:submit|preventDefault={handleSubmit}>
            <div class="row mb-3">
                <div class="col-md-4">
                    <label for="tanggal" class="form-label fw-bold">Pilih Tanggal Absensi:</label>
                    <input type="date" class="form-control" id="tanggal" bind:value={tanggal} required>
                </div>
            </div>

            {#if isLoading}
                <p>Memuat daftar siswa...</p>
            {:else if siswaList.length > 0}
            <div class="table-responsive">
                <table class="table">
                    <thead><tr><th>Nama Siswa</th><th>Status Kehadiran</th></tr></thead>
                    <tbody>
                        {#each siswaList as siswa (siswa.nisn)}
                            <tr>
                                <td class="align-middle">{siswa.namaLengkap} ({siswa.nisn})</td>
                                <td>
                                    <select class="form-select" bind:value={attendanceStatus[siswa.nisn]}>
                                        <option value="Hadir">Hadir</option>
                                        <option value="Sakit">Sakit</option>
                                        <option value="Izin">Izin</option>
                                        <option value="Alpa">Alpa</option>
                                    </select>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
            <button type="submit" class="btn btn-primary" disabled={isSaving}>
                {#if isSaving}
                    <span class="spinner-border spinner-border-sm"></span> Menyimpan...
                {:else}
                    <i class="bi bi-save"></i> Simpan Semua Kehadiran
                {/if}
            </button>
            {:else}
                <div class="alert alert-warning">Tidak ada data siswa yang bisa ditampilkan. Pastikan data di sheet 'DataSiswa' sudah benar.</div>
            {/if}
        </form>
    </div>
</div>