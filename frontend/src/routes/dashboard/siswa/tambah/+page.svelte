<script>
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation'; // Impor 'goto' untuk redirect
  import { onMount } from 'svelte';

  let token = '';
  let siswa = {
    NISN: '',
    NamaLengkap: '',
    Kelas: '',
    NomorTeleponOrtu: '',
    EmailOrtu: ''
  };

  let isLoading = false;
  let errorMessage = '';
  let successMessage = '';

  onMount(() => {
    if (browser) {
      token = localStorage.getItem('jwt_token') || '';
    }
  });

  async function handleSubmit() {
    isLoading = true;
    errorMessage = '';
    successMessage = '';

    try {
      const apiUrl = import.meta.env.VITE_API_BASE_URL;
      const response = await fetch('${apiUrl}/api/admin/siswa/tambah', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json', // Kirim sebagai JSON
          'Authorization': 'Bearer ' + token
        },
        body: JSON.stringify(siswa) // Ubah objek siswa menjadi string JSON
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Gagal menambahkan siswa.');
      }

      successMessage = data.message;
      // Redirect kembali ke halaman manajemen siswa setelah 2 detik
      setTimeout(() => {
        goto('/dashboard/siswa');
      }, 2000);

    } catch (/** @type {any} */ error) {
      errorMessage = error.message;
    } finally {
      isLoading = false;
    }
  }
</script>

<svelte:head>
  <title>Tambah Siswa Baru</title>
</svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Tambah Siswa Baru</h1>
    <a href="/dashboard/siswa" class="btn btn-sm btn-outline-secondary">
        <i class="bi bi-arrow-left"></i> Kembali ke Daftar Siswa
    </a>
</div>

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

                <form on:submit|preventDefault={handleSubmit}>
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
                    <button type="submit" class="btn btn-primary" disabled={isLoading}>
                        {#if isLoading}
                            <span class="spinner-border spinner-border-sm"></span>
                            Menyimpan...
                        {:else}
                            Simpan Data Siswa
                        {/if}
                    </button>
                </form>
            </div>
        </div>
    </div>
</div>