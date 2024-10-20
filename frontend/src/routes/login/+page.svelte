<script lang="ts">
	import { BASE_DIR } from '$lib/baseDir';
	import Navbar from '$lib/Navbar.svelte';
	import { notify } from '$lib/notificationHelper';
	import { startAuthentication } from '@simplewebauthn/browser';

	let username = '';

	async function login() {
		try {
			// const response = await fetch(`${BASE_DIR}/webauthn/passkey/registerBegin`, {
			// 	method: 'POST',
			// 	headers: { 'Content-Type': 'application/json' },
			// 	body: JSON.stringify({ username })
			// });

			const response = await fetch(`${BASE_DIR}/webauthn/passkey/loginBegin`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ username: username })
			});
			// Check if the login options are ok.
			if (!response.ok) {
				const msg = await response.json();
				throw new Error('Failed to get login options from server: ' + msg);
			}
			// Convert the login options to JSON.
			const options = await response.json();

			// This triggers the browser to display the passkey / WebAuthn modal (e.g. Face ID, Touch ID, Windows Hello).
			// A new assertionResponse is created. This also means that the challenge has been signed.
			const assertionResponse = await startAuthentication({
				optionsJSON: options.publicKey
			});

			// Send assertionResponse back to server for verification.
			const verificationResponse = await fetch(`${BASE_DIR}/webauthn/passkey/loginFinish`, {
				method: 'POST',
				credentials: 'same-origin',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(assertionResponse)
			});

			const msg = await verificationResponse.json();
			if (verificationResponse.ok) {
				notify.success(msg);
				localStorage.setItem('username', username);
			} else {
				notify.error(msg);
			}
		} catch (error) {
			console.error(error);
			if (error instanceof Error) {
				notify.error(error.message);
			} else {
				notify.error('An unknown error occurred');
			}
		}
	}
</script>

<Navbar></Navbar>

<!-- <div class="flex items-center"> -->
<div class="flex flex-col justify-center items-center">
	<div class="card w-96 bg-base-100 shadow-xl m-10 space-y-3">
		<div class="card-body">
			<h1 class="card-title">Login</h1>
			<label class="input input-bordered flex items-center gap-2">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 16 16"
					fill="currentColor"
					class="w-4 h-4 opacity-70"
					><path
						d="M8 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM12.735 14c.618 0 1.093-.561.872-1.139a6.002 6.002 0 0 0-11.215 0c-.22.578.254 1.139.872 1.139h9.47Z"
					/></svg
				>
				<input type="text" class="grow" placeholder="Username" bind:value={username} />
			</label>

			<button class="btn btn-primary" on:click={login}>Sign In</button>
		</div>
	</div>
	<p>Dont have an account? <a href="/register" class="link link-accent">Sign up</a></p>
</div>
