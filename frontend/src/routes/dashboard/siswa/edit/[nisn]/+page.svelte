<script>
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

  const nisn = $page.params.nisn;
  let token = '';
  /** @type {any} */
  let siswa = null;
  
  let isLoading = true;
  let isSaving = false;
  let errorMessage = '';
  let successMessage = '';

  // 1. Mengambil data siswa yang akan diedit saat halaman dimuat
  onMount(async () => {
    if (browser) {
      token = localStorage.getItem('jwt_token') || '';
      try {
        const apiUrl = import.meta.env.VITE_API_BASE_URL;
        const response = await fetch(`${apiUrl}/api/siswa/${nisn}`, {
          headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!response.ok) throw new Error('Gagal mengambil data siswa.');
        siswa = await response.json();
      } catch (/**@type {any}*/error) {
        errorMessage = error.message;
      } finally {
        isLoading = false;
      }
    }
  });

  // 2. Fungsi untuk menangani submit form (sebelumnya kosong)
  async function handleUpdate() {
    isSaving = true;
    errorMessage = '';
    successMessage = '';

    try {
      const apiUrl = import.meta.env.VITE_API_BASE_URL;
      const response = await fetch(`${apiUrl}/api/siswa/${nisn}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + token
        },
        body: JSON.stringify(siswa)
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.message);
      
      successMessage = data.message;
      setTimeout(() => {
        goto('/dashboard/siswa');
      }, 2000);

    } catch (/**@type {any}*/error) {
      errorMessage = error.message;
    } finally {
      isSaving = false;
    }
  }
</script>

<svelte:head>
  <title>Edit Siswa {nisn}</title>
</svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Edit Data Siswa</h1>
    <a href="/dashboard/siswa" class="btn btn-sm btn-outline-secondary">
        <i class="bi bi-arrow-left"></i> Kembali
    </a>
</div>

{#if isLoading}
    <p>Memuat data siswa...</p>
{:else if siswa}
    <div class="row">
        <div class="col-lg-8">
            <div class="card shadow-sm">
                <div class="card-body">
                    {#if errorMessage}
                        <div class="alert alert-danger">{errorMessage}</div>
                    {/if}
                    {#if successMessage}
                        <div class="alert alert-success">{successMessage}. Anda akan diarahkan kembali...</div>
                    {/if}
                    
                    <form on:submit|preventDefault={handleUpdate}>
                        <div class="mb-3">
                            <label for="nisn" class="form-label">NISN</label>
                            <input type="text" class="form-control" id="nisn" bind:value={siswa.NISN} required>
                        </div>
                        <div class="mb-3">
                            <label for="nama_lengkap" class="form-label">Nama Lengkap</label>
                            <input type="text" class="form-control" id="nama_lengkap" bind:value={siswa.NamaLengkap} required>
                        </div>
                        <div class="mb-3">
                            <label for="kelas" class="form-label">Kelas</label>
                            <input type="text" class="form-control" id="kelas" bind:value={siswa.Kelas} required>
                        </div>
                        <div class="mb-3">
                            <label for="kontak_ortu" class="form-label">Kontak Orang Tua</label>
                            <input type="text" class="form-control" id="kontak_ortu" bind:value={siswa.NomorTeleponOrtu} required>
                        </div>
                        <div class="mb-3">
                            <label for="email_ortu" class="form-label">Email Orang Tua</label>
                            <input type="email" class="form-control" id="email_ortu" bind:value={siswa.EmailOrtu} required>
                        </div>
                        <button type="submit" class="btn btn-primary" disabled={isSaving}>
                            {#if isSaving}
                                <span class="spinner-border spinner-border-sm"></span>
                                Menyimpan...
                            {:else}
                                Simpan Perubahan
                            {/if}
                        </button>
                    </form>
                </div>
            </div>
        </div>
    </div>
{:else}
    <div class="alert alert-danger">{errorMessage || 'Data siswa tidak ditemukan.'}</div>
{/if}