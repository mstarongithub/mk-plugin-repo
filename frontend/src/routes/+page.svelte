<script lang="ts">
	import type { Plugin } from '$lib';
	import Navbar from '$lib/Navbar.svelte';
	import PluginListing from '$lib/PluginListing.svelte';
	import ThemeSwitcher from '$lib/ThemeSwitcher.svelte';
	import { BASE_DIR } from '$lib/baseDir';
	import toast from 'svelte-french-toast';


	let plugins: Plugin[] = [];

	async function fetchListings() {
		let response = await fetch(`${BASE_DIR}/api/v1/plugins`, {
			headers: {
				'Content-Type': 'application/json'
			},
			method: 'GET'
		});

		if (response.ok) {
			const newLoadedPlugins : Plugin[] = await response.json();
      
    		plugins = [...plugins, ...newLoadedPlugins];

			console.log(newLoadedPlugins);
		} else {
			let err = await response;
			console.error(err);

			toast.error('Unable to load plugins', {
				className: '!btn'
			});
		}
	}
	fetchListings();
</script>

<Navbar></Navbar>

<!-- <h1 class="text-3xl text-red-600 font-bold underline">
    Plugins here
</h1> -->

<div class="flex justify-center items-center">
	<!-- grid-cols-3 -->
	<div class="w-10/12 h-screen flex flex-wrap gap-2">
    {#each plugins as plugin}
  		<PluginListing pluginName={plugin.name} shortDesc={plugin.summary_short} tags={plugin.tags} id={plugin.id}></PluginListing>
    {/each}
	</div>
</div>
