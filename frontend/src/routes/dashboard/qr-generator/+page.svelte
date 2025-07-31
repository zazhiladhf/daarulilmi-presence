<script>
    import { browser } from '$app/environment';
    import { onMount } from 'svelte';

    let token = '';
    let errorMessage = '';

    onMount(() => {
        if (browser) {
            token = localStorage.getItem('jwt_token') || '';
        }
    });

    // @ts-ignore
    async function fetchQRCode(type) {
        const qrImg = document.getElementById('qr-code-img');
        const qrLoading = document.getElementById('qr-loading');
        const qrTitle = document.getElementById('qr-title');
        
        if(!qrImg || !qrLoading || !qrTitle) return;

        qrImg.style.display = 'none';
        qrLoading.style.display = 'block';
        qrTitle.style.display = 'block';
        qrTitle.innerText = `Memuat QR Code Absen ${type}...`;
        errorMessage = ''; // Hapus pesan error lama

        try {
            const apiUrl = import.meta.env.VITE_API_BASE_URL;
            const response = await fetch(`${apiUrl}/api/qr/generate?type=${type}`, {
                method: 'GET',
                headers: { 'Authorization': 'Bearer ' + token },
                cache: 'no-store'
            });

            if (!response.ok) {
                // Jika server merespons dengan error (seperti 400 Bad Request)
                const errorData = await response.json();
                throw new Error(errorData.message || 'Gagal mengambil data dari server.');
            }

            const blob = await response.blob();
            // @ts-ignore
            qrImg.src = URL.createObjectURL(blob);
            qrImg.style.display = 'block';
            qrLoading.style.display = 'none';
            qrTitle.innerText = `Scan QR Code Absen ${type}`;

        } catch (/**@type {any}*/error) {
            console.error('Error fetching QR code:', error);
            qrLoading.style.display = 'none';
            qrTitle.style.display = 'none';
            errorMessage = error.message;
        }
    }
</script>

<svelte:head>
    <title>QR Code Generator</title>
</svelte:head>

<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2">QR Code Generator</h1>
</div>

{#if errorMessage}
    <div class="alert alert-danger">{errorMessage}</div>
{/if}

<div class="card mt-3 shadow-sm">
    <div class="card-body">
        <div class="row align-items-center">
            <div class="col-md-7">
                <p class="mb-2">Klik tombol untuk menampilkan QR code absensi masuk atau pulang. QR code hanya valid selama 60 detik.</p>
                <button class="btn btn-success mb-2" on:click={() => fetchQRCode('masuk')}>
                    <i class="bi bi-box-arrow-in-right"></i> Tampilkan QR Masuk
                </button>
                <button class="btn btn-danger mb-2" on:click={() => fetchQRCode('pulang')}>
                    <i class="bi bi-box-arrow-left"></i> Tampilkan QR Pulang
                </button>
            </div>
            <div class="col-md-5 text-center">
                <h5 id="qr-title" class="d-none">Scan QR Code Ini</h5>
                <img id="qr-code-img" src="" alt="QR Code akan muncul di sini" class="img-fluid border rounded" style="max-width: 256px; display: none;">
                <p id="qr-loading" style="display: none;">Memuat QR Code...</p>
            </div>
        </div>
    </div>
</div>