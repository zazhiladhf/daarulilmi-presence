<script>
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  /** @type {any[]} */
  let siswaList = [];
  let isLoading = true;
  let errorMessage = '';

  onMount(async () => {
    if (browser) {
      const token = localStorage.getItem('jwt_token');
      try {
        const apiUrl = import.meta.env.VITE_API_BASE_URL;
        const response = await fetch('${apiUrl}/api/siswa', {
          headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!response.ok) {
          throw new Error('Gagal mengambil data siswa dari server.');
        }
        const data = await response.json();
        siswaList = data;
      } catch (/** @type {any} */ error) {
        errorMessage = error.message;
        console.error(error);
      } finally {
        isLoading = false;
      }
    }
  });

  /**
	 * @param {any} nisn
	 * @param {any} nama
	 */
  async function handleDelete(nisn, nama) {
    if (!confirm(`Apakah Anda yakin ingin menghapus data siswa "${nama}"?`)) {
        return;
    }
    
    try {
        const token = localStorage.getItem('jwt_token');
        const apiUrl = import.meta.env.VITE_API_BASE_URL;
        const response = await fetch(`${apiUrl}/api/siswa/${nisn}`, {
            method: 'DELETE',
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!response.ok) throw new Error('Gagal menghapus data.');
        
        // Hapus siswa dari daftar di frontend agar UI update
        siswaList = siswaList.filter(s => s.NISN !== nisn);

    } catch (/**@type {any}*/error) {
        alert(error.message);
    }
  }
</script>

<svelte:head>
    <title>Manajemen Siswa</title>
</svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Manajemen Siswa</h1>
    <div class="btn-toolbar mb-2 mb-md-0">
        <a href="/dashboard/siswa/tambah" class="btn btn-sm btn-primary">
            <i class="bi bi-plus-circle"></i> Tambah Siswa Baru
        </a>
    </div>
</div>

{#if errorMessage}
    <div class="alert alert-danger">{errorMessage}</div>
{/if}

<!-- <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">Manajemen Siswa</h1>
    <div class="btn-toolbar mb-2 mb-md-0">
        <a href="/dashboard/siswa/tambah" class="btn btn-sm btn-primary">
            <i class="bi bi-plus-circle"></i> Tambah Siswa Baru
        </a>
    </div>
</div>

{#if errorMessage}
    <div class="alert alert-danger">{errorMessage}</div>
{/if} -->

<div class="card shadow-sm">
    <div class="card-body">
        <div class="table-responsive">
            <table class="table table-striped table-hover">
                <thead>
                    <tr>
                        <th>NISN</th><th>Nama Lengkap</th><th>Kelas</th><th>Kontak Ortu</th><th>Aksi</th>
                    </tr>
                </thead>
                <tbody>
                    {#if isLoading}
                        <tr><td colspan="5" class="text-center">Memuat data siswa...</td></tr>
                    {:else if siswaList.length > 0}
                        {#each siswaList as siswa}
                            <tr>
                                <td>{siswa.nisn}</td>
                                <td>{siswa.namaLengkap}</td>
                                <td>{siswa.kelas}</td>
                                <td>{siswa.nomorTeleponOrtu}</td>
                                <td>
                                    <!-- svelte-ignore a11y_consider_explicit_label -->
                                    <a href="/dashboard/siswa/edit/{siswa.nisn}" class="btn btn-sm btn-warning"><i class="bi bi-pencil-square"></i></a>
                                    <!-- svelte-ignore a11y_consider_explicit_label -->
                                    <button class="btn btn-sm btn-danger" on:click={() => handleDelete(siswa.nisn, siswa.namaLengkap)}><i class="bi bi-trash"></i></button>
                                </td>
                            </tr>
                        {/each}
                    {:else}
                        <tr><td colspan="5" class="text-center">Belum ada data siswa.</td></tr>
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>