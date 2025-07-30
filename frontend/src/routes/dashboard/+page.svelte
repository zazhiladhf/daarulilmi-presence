<script>
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';

  /** @type {any} */
  let dashboardData = null;
  let errorMessage = '';
  let isLoading = true;
  let token = '';

  async function fetchData() {
    isLoading = true;
    errorMessage = '';
    
    // Pastikan token ada sebelum fetch
    if (!token) {
        isLoading = false;
        errorMessage = 'Sesi Anda telah berakhir. Silakan login kembali.';
        return;
    }

    try {
      const response = await fetch(`http://localhost:1412/api/dashboard-data?t=${new Date().getTime()}`, {
        headers: { 'Authorization': 'Bearer ' + token },
        cache: 'no-store'
      });
      
      // TAMBAHAN: Cek jika token ditolak (401 Unauthorized)
      if (response.status === 401) {
          localStorage.removeItem('jwt_token'); // Hapus token yang tidak valid
          window.location.href = '/'; // Paksa kembali ke login
          throw new Error('Sesi tidak valid. Harap login kembali.');
      }

      if (!response.ok) {
          throw new Error('Gagal memuat data dashboard dari server.');
      }
      
      dashboardData = await response.json();

    } catch (/** @type {any} */ error) {
      errorMessage = error.message;
      console.error("Fetch error:", error); // Tampilkan error di console
    } finally {
      isLoading = false;
    }
  }
  
  /**
	 * @param {any} nisn
	 * @param {string} status
	 */
  async function handleManualAttendance(nisn, status) {
        if (!confirm(`Anda yakin ingin mengubah status NISN ${nisn} menjadi "${status}"?`)) return;
        
        try {
            const response = await fetch('http://localhost:1412/api/absensi/manual', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token },
                body: JSON.stringify({ NISN: nisn, Status: status })
            });
            const data = await response.json();
            if (!response.ok) throw new Error(data.message);
            
            alert(data.message);
            
            // --- GANTI BAGIAN INI ---
            // Hapus: await invalidateAll(); fetchData();
            // Ganti dengan:
            window.location.reload();
            // --- AKHIR PERUBAHAN ---

        } catch (/** @type {any} */ error) {
            alert(`Error: ${error.message}`);
        }
    }

  onMount(() => {
    if (browser) {
      token = localStorage.getItem('jwt_token') || '';
      fetchData(); // Panggil fetchData

      // Logika untuk jam tetap di sini, karena tidak bergantung pada data fetch
      const dateTimeElement = document.getElementById('live-datetime');
      if (dateTimeElement) {
        const timer = setInterval(() => {
            const now = new Date();
            const date = now.toLocaleDateString('id-ID', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' });
            const time = now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
            dateTimeElement.innerText = `Waktu Saat Ini: ${date} pukul ${time}`;
        }, 1000);

        return () => clearInterval(timer); // Cleanup
      }
    }
  });
</script>

<svelte:head>
  <title>Dashboard Kehadiran</title>
</svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
  <div>
    {#if dashboardData}
      <h1 class="h2">Selamat Datang, {dashboardData.namaLengkapUser}!</h1>
    {:else}
      <h1 class="h2">Memuat...</h1>
    {/if}
    <p class="text-muted" id="live-datetime">Memuat waktu...</p>
  </div>
  <div class="btn-toolbar mb-2 mb-md-0">
    <button class="btn btn-sm btn-outline-secondary" on:click={fetchData} disabled={isLoading}>
      <i class="bi bi-arrow-clockwise"></i> Refresh
    </button>
    <a href="/dashboard/kehadiran/tambah" class="btn btn-sm btn-outline-primary ms-2">
        <i class="bi bi-calendar-plus"></i> Catat Absensi Lalu
    </a>
  </div>
</div>

{#if isLoading}
  <p>Memuat data dashboard...</p>
{:else if errorMessage}
  <div class="alert alert-danger">{errorMessage}</div>
{:else if dashboardData}
  {#if dashboardData.isHoliday}
    <div class="alert alert-info text-center">
      <h4 class="alert-heading"><i class="bi bi-calendar-heart"></i> Hari Libur</h4>
      <p>{dashboardData.holidayDescription}. Tidak ada pencatatan absensi hari ini.</p>
    </div>
  {:else}
    <div class="row">
        <div class="col-lg-3 mb-3"><div class="card bg-primary text-white"><div class="card-body"><h5 class="card-title">Total Siswa</h5><p class="card-text fs-2 fw-bold">{dashboardData.totalSiswa}</p></div></div></div>
        <div class="col-lg-3 mb-3"><div class="card bg-success text-white"><div class="card-body"><h5 class="card-title">Hadir</h5><p class="card-text fs-2 fw-bold">{dashboardData.totalHadir}</p></div></div></div>
        <div class="col-lg-3 mb-3"><div class="card bg-warning text-dark"><div class="card-body"><h5 class="card-title">Izin/Sakit</h5><p class="card-text fs-2 fw-bold">{dashboardData.totalIzin}</p></div></div></div>
        <div class="col-lg-3 mb-3"><div class="card bg-secondary text-white"><div class="card-body"><h5 class="card-title">Belum Ada Kabar</h5><p class="card-text fs-2 fw-bold">{dashboardData.totalBelumAdaKabar}</p></div></div></div>
    </div>
    
    <div class="card mt-3 shadow-sm">
      <div class="card-header"><h5><i class="bi bi-list-task"></i> Status Kehadiran Seluruh Siswa Hari Ini</h5></div>
      <div class="card-body">
        <div class="table-responsive">
          <table class="table table-striped table-hover">
            <thead>
              <tr><th>NISN</th><th>Nama Lengkap</th><th>Status</th><th>Keterangan</th><th class="text-center">Aksi Manual</th></tr>
            </thead>
            <tbody>
              {#each dashboardData.daftarStatusSiswa as siswa}
                <tr>
                  <td>{siswa.nisn}</td><td>{siswa.namaLengkap}</td>
                  <td>
                    {#if siswa.status === 'Hadir'} <span class="badge bg-success">{siswa.status}</span>
                    {:else if siswa.status === 'Sakit' || siswa.status === 'Izin'} <span class="badge bg-warning text-dark">{siswa.status}</span>
                    {:else if siswa.status === 'Alpa'} <span class="badge bg-danger">{siswa.status}</span>
                    {:else} <span class="badge bg-secondary">{siswa.status}</span> {/if}
                  </td>
                  <td>{siswa.keterangan}</td>
                  <td class="text-center">
                    <div class="btn-group btn-group-sm" role="group">
                      <button class="btn btn-outline-success" title="Set Hadir" disabled={siswa.status === 'Hadir'} on:click={() => handleManualAttendance(siswa.nisn, 'Hadir')}>H</button>
                      <button class="btn btn-outline-warning" title="Set Sakit" disabled={siswa.status === 'Sakit'} on:click={() => handleManualAttendance(siswa.nisn, 'Sakit')}>S</button>
                      <button class="btn btn-outline-info" title="Set Izin" disabled={siswa.status === 'Izin'} on:click={() => handleManualAttendance(siswa.nisn, 'Izin')}>I</button>
                      <button class="btn btn-outline-danger" title="Set Alpa" disabled={siswa.status === 'Alpa'} on:click={() => handleManualAttendance(siswa.nisn, 'Alpa')}>A</button>
                    </div>
                  </td>
                </tr>
              {:else}
                <tr><td colspan="5" class="text-center">Data siswa tidak ditemukan.</td></tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  {/if}
{/if}