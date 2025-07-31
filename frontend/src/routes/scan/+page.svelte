<script>
	import { browser } from '$app/environment';
	import { onMount, onDestroy } from 'svelte';
	import { Html5Qrcode } from 'html5-qrcode';

	// @ts-ignore
	let html5QrCode;
	let scanResult = { type: '', message: '' }; // type: 'success' | 'error' | 'info'

	// Fungsi yang dijalankan saat QR code berhasil dipindai
	// @ts-ignore
	async function onScanSuccess(decodedText) {
		// Hentikan scanner agar tidak memindai berulang kali
		// @ts-ignore
		if (html5QrCode && html5QrCode.isScanning) {
			try {
				await html5QrCode.stop();
			} catch (err) {
				console.error('Gagal menghentikan scanner.', err);
			}
		}

		scanResult = { type: 'info', message: 'Memproses absensi...' };

		const token = localStorage.getItem('jwt_token');

		try {
			const apiUrl = import.meta.env.VITE_API_BASE_URL;
			const response = await fetch('${apiUrl}/api/absensi/scan', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': 'Bearer ' + token
				},
				body: JSON.stringify({ qr_data: decodedText })
			});

			const data = await response.json();
			if (!response.ok) throw new Error(data.message);

			scanResult = { type: 'success', message: data.message };
		} catch (/** @type {any} */ error) {
			scanResult = { type: 'error', message: error.message };
		}
	}

	onMount(() => {
		if (browser) {
			const token = localStorage.getItem('jwt_token');
			if (!token) {
				window.location.href = '/'; // Arahkan ke login jika belum login
				return;
			}

			// Inisialisasi scanner
			html5QrCode = new Html5Qrcode('reader');
			const config = { fps: 10, qrbox: { width: 250, height: 250 } };

			// Mulai scanner
			html5QrCode.start({ facingMode: 'environment' }, config, onScanSuccess, undefined)
                // @ts-ignore
                .catch(err => {
                    scanResult = { type: 'error', message: 'Gagal memulai kamera. Pastikan Anda memberikan izin.' };
                });
		}
	});

	// Pastikan kamera berhenti saat pengguna meninggalkan halaman
	onDestroy(() => {
		// @ts-ignore
		if (browser && html5QrCode && html5QrCode.isScanning) {
			// @ts-ignore
			html5QrCode.stop().catch(err => {
                console.error("Gagal membersihkan scanner saat keluar.", err)
            });
		}
	});
</script>

<svelte:head>
	<title>Scan Presensi</title>
</svelte:head>

<div class="d-flex flex-column align-items-center justify-content-center vh-100 bg-dark text-white">
	<div class="text-center p-3">
		<h2 class="mb-3">Scan QR Code Presensi</h2>
		
		<div id="reader" style="width: 300px; border: 2px solid #555; border-radius: 8px;"></div>

		<div class="mt-4 fs-5" style="min-height: 50px;">
			{#if scanResult.message}
				<div class:alert-success={scanResult.type === 'success'}
					 class:alert-danger={scanResult.type === 'error'}
					 class:alert-info={scanResult.type === 'info'}
					 class="alert"
					 role="alert">
					{#if scanResult.type === 'success'} <i class="bi bi-check-circle-fill"></i>
					{:else if scanResult.type === 'error'} <i class="bi bi-exclamation-triangle-fill"></i>
					{:else} <span class="spinner-border spinner-border-sm"></span>
					{/if}
					{scanResult.message}
				</div>
			{/if}
		</div>
	</div>
</div>