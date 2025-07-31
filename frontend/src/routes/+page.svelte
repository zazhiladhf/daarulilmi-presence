<script>
    import { browser } from '$app/environment';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';

    let username = '';
    let password = '';
    let errorMessage = '';
    let isLoading = false;

    // Cek jika pengguna sudah login saat halaman dimuat
    onMount(() => {
        if (browser) {
            const token = localStorage.getItem('jwt_token');
            if (token) {
                // Jika sudah ada token, langsung arahkan berdasarkan peran
                redirectByRole(token);
            }
        }
    });

    // --- FUNGSI BARU UNTUK MENGARAHKAN BERDASARKAN PERAN ---
    // @ts-ignore
    function redirectByRole(token) {
        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            const role = payload.role;

            if (role === 'admin') {
                goto('/dashboard');
            } else if (role === 'siswa') {
                goto('/scan');
            } else if (role === 'ortu') {
                goto('/portal');
            } else {
                // Fallback jika peran tidak dikenal
                errorMessage = 'Peran pengguna tidak dikenali.';
                localStorage.removeItem('jwt_token'); // Hapus token yang salah
            }
        } catch (e) {
            console.error("Gagal mendekode token:", e);
            errorMessage = "Token tidak valid, silakan login kembali.";
            localStorage.removeItem('jwt_token');
        }
    }

    async function handleLogin() {
        isLoading = true;
        errorMessage = '';
        const apiUrl = import.meta.env.VITE_API_BASE_URL;
        try {
            const response = await fetch('${apiUrl}/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: new URLSearchParams({
                    'username': username,
                    'password': password
                })
            });
            
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.message || 'Login gagal.');
            }

            if (data.token) {
                localStorage.setItem('jwt_token', data.token);
                // Panggil fungsi redirect setelah berhasil login
                redirectByRole(data.token);
            } else {
                throw new Error('Token tidak diterima dari server.');
            }

        } catch (/** @type {any} */ error) {
            errorMessage = error.message;
        } finally {
            isLoading = false;
        }
    }
</script>

<svelte:head>
  <title>Login - DI Smart Presence</title>
</svelte:head>

<div class="bg-light">
    <div class="container">
        <div class="row justify-content-center align-items-center" style="min-height: 100vh;">
            <div class="col-md-5 col-lg-4">
                <div class="card shadow-sm">
                    <div class="card-body p-4">
                        <h3 class="card-title text-center mb-4">Smart Presence Login</h3>
                        
                        {#if errorMessage}
                          <div class="alert alert-danger" role="alert">
                            {errorMessage}
                          </div>
                        {/if}

                        <form on:submit|preventDefault={handleLogin}>
                            <div class="mb-3">
                                <label for="username" class="form-label">Username</label>
                                <input type="text" class="form-control" id="username" bind:value={username} required>
                            </div>
                            <div class="mb-3">
                                <label for="password" class="form-label">Password</label>
                                <input type="password" class="form-control" id="password" bind:value={password} required>
                            </div>
                            <div class="d-grid mt-4">
                                <button type="submit" class="btn btn-primary" disabled={isLoading}>
                                    {#if isLoading}
                                        <span class="spinner-border spinner-border-sm"></span> Loading...
                                    {:else}
                                        Masuk
                                    {/if}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>