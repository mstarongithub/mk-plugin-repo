<script lang="ts">
	import type { Plugin } from '$lib';
	import Navbar from '$lib/Navbar.svelte';
	import PluginCode from '$lib/PluginCode.svelte';
	import PluginListing from '$lib/PluginListing.svelte';
	import { BASE_DIR } from '$lib/baseDir';
	import { onMount } from 'svelte';
	import toast from 'svelte-french-toast';


	let selectedPluginData : Plugin | undefined = undefined;

	onMount(async () => {
		let params = new URLSearchParams(location.search);
		let pluginId = params.get('id');
		if (pluginId) {
			let response = await fetch(`${BASE_DIR}/api/v1/plugins/${pluginId}`, {
				// body: JSON.stringify({

				// }),
				headers: {
					'Content-Type': 'application/json'
				},
				method: 'GET'
			});

			if (response.ok) {
				selectedPluginData = await response.json();
				console.log(selectedPluginData)

			} else {
				let err = await response;
				console.error(err);

				toast.error('Server Error', {
					className: '!btn'
				});
			}
		}
	});
</script>

<Navbar></Navbar>

<div class="flex flex-col gap-2 justify-center items-center">
	<!-- grid-cols-3 -->
	{#if selectedPluginData}
		<PluginListing className="lg:card-side !w-10/12 !h-3/4" id={-1} pluginName={selectedPluginData.name} shortDesc={selectedPluginData.summary_long} tags={selectedPluginData.tags}></PluginListing>
	{/if}
	<div class="card !w-10/12 !h-3/4 bg-base-100 shadow-xl overflow-clip">
		<PluginCode></PluginCode>
	</div>
</div>
