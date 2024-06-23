<script lang="ts">
	import { browser } from '$app/environment';
	import ThemeSwitcher from './ThemeSwitcher.svelte';
	import logo from './favicon.svg';
	import Shortcuts from 'shortcuts';
	let inputEl: HTMLInputElement;
	let smallDevice = false;

	let loggedIn = false;

	if (browser) {
		const shortcuts = new Shortcuts({
			capture: true, // Handle events during the capturing phase
			target: document, // Listening for events on the document object
			shouldHandleEvent() {
				return true; // Handle all possible events
			}
		});

		shortcuts.add([
			{
				shortcut: 'CmdOrCtrl+K',
				handler: () => {
					inputEl.select();
					return true; // Returning true if we don't want other handlers for the same shortcut to be called later
				}
			}
		]);

		shortcuts.start();

		// invoke this function as soon as window is available
		const attachListener = () => {
			// attach a media query listener to the window
			const mediaQuery = window.matchMedia('(width <= 640px)');
			smallDevice = mediaQuery.matches;

			// every time the media query matches or unmatches
			mediaQuery.addEventListener('change', ({ matches }) => {
				// set the state of our variable
				smallDevice = matches;
			});
		};
		attachListener();
	}
</script>

<div class="navbar bg-base-100 sticky top-0 z-50">
	<div class="flex-1">
		<!-- {#if !smallDevice}
			<a class="btn btn-ghost text-xl">MK-Plugin-Repo</a>
		{:else}
			<a class="btn btn-ghost text-xl">MKP</a>
		{/if} -->
		<a class="btn btn-ghost text-xl" href="/"
			><img src={logo} alt="The website logo" class="max-h-9" /></a
		>
	</div>
	<div class="flex-none gap-2">
		<div class="form-control">
			<!-- <input type="text" placeholder="Search" class="input input-bordered w-24 md:w-auto" /> -->
			<label class="input input-bordered flex items-center gap-2">
				<input type="text" class="grow" placeholder="Search" bind:this={inputEl} />
				<!-- Dont show shortcuts on mobile -->
				{#if !smallDevice}
					<kbd class="kbd kbd-sm">⌘</kbd>
					<kbd class="kbd kbd-sm">K</kbd>
				{/if}
			</label>
		</div>
		{#if loggedIn}
			<div class="dropdown dropdown-end">
				<div tabindex="0" role="button" class="btn btn-ghost btn-circle avatar">
					<div class="w-10 rounded-full">
						<img
							alt="Tailwind CSS Navbar component"
							src="https://woem.men/files/thumbnail-f2f87598-0eca-4abf-b1b2-800bdde1d8bf"
						/>
					</div>
				</div>
				<ul
					tabindex="0"
					class="mt-3 z-[1] p-2 shadow menu menu-sm dropdown-content bg-base-100 rounded-box w-52"
				>
					<li class="pl-2 rounded-full"><ThemeSwitcher></ThemeSwitcher></li>

					<li><a href="/submit">Submit Plugin/Widget</a></li>
					<!-- <li>
						<a class="justify-between">
							Profile
							<span class="badge">New</span>
						</a>
					</li> -->
					<!-- <li><a>Settings</a></li> -->
					<li><a>Logout</a></li>
				</ul>
			</div>
		{:else}
			<a href="/signup" class="btn btn-outline">Sign up</a>

			<a href="/login" class="btn btn-outline">Login</a>
		{/if}
	</div>
</div>
