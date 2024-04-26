<script lang="ts">
	import type { Plugin } from '$lib';
	import Navbar from '$lib/Navbar.svelte';
	import PluginCode from '$lib/PluginCode.svelte';
	import PluginListing from '$lib/PluginListing.svelte';
	import { getAIscriptVersion } from '$lib/aiScriptCodeParsers';
	import { BASE_DIR } from '$lib/baseDir';
	import { onMount } from 'svelte';
	import toast from 'svelte-french-toast';

	let selectedPluginData: Plugin | undefined = undefined;
	let code: string;
	let aiscriptVersion: string = "...";
	let selectedVersion: string;
	let pluginId: string = '';

	onMount(async () => {
		let params = new URLSearchParams(location.search);
		pluginId = params.get('id') ?? '';
		if (pluginId) {
			let response = await fetch(`${BASE_DIR}/api/v1/plugins/${pluginId}`, {
				headers: {
					'Content-Type': 'application/json'
				},
				method: 'GET'
			});

			if (response.ok) {
				selectedPluginData = await response.json();
				selectedVersion = selectedPluginData?.current_version ?? '';
				showCode();
			} else {
				let err = await response;
				console.error(err);

				toast.error('Server Error', {
					className: '!btn'
				});
			}
		}
	});

	const showCode = async () => {
		if (pluginId) {
			let response = await fetch(`${BASE_DIR}/api/v1/plugins/${pluginId}/${selectedVersion}`, {
				headers: {
					'Content-Type': 'application/json'
				},
				method: 'GET'
			});

			if (response.ok) {
				let data = await response.json();
				code = data.code;
				aiscriptVersion = data.aiscript_version;

				if (aiscriptVersion === "") {
					aiscriptVersion = getAIscriptVersion(code) ?? "";
				}
				console.log(data);
				//data.aiscript_version
			} else {
				let err = await response;
				console.error(err);

				toast.error('Server Error', {
					className: '!btn'
				});
			}

		}
	};
</script>

<Navbar></Navbar>

<div class="flex flex-col gap-2 justify-center items-center">
	<!-- grid-cols-3 -->
	{#if selectedPluginData}
		<PluginListing
			className="lg:card-side !w-10/12 !h-3/4"
			pluginName={selectedPluginData.name}
			shortDesc={selectedPluginData.summary_long}
			tags={selectedPluginData.tags}
		></PluginListing>
	{/if}

	<div class="card flex flex-row items-center justify-between !w-10/12 !h-3/4 bg-base-100 shadow-xl overflow-clip p-4">
		<select
			class="select select-bordered w-full max-w-40"
			on:change={showCode}
			bind:value={selectedVersion}
		>
			<!-- <option disabled selected>Who shot first?</option> -->
			{#if selectedPluginData}
				{#each (selectedPluginData.all_versions) as version}
					<option selected={version == selectedPluginData.current_version} value={version}>V{version}</option>
				{/each}
			{/if}
		</select>

		<!-- <div class = "align-middle h-full"> -->
		<p class="text-center ">For AIscript @{aiscriptVersion}</p>
		<!-- </div> -->
	</div>

	<div class="card !w-10/12 !h-3/4 bg-base-100 shadow-xl overflow-clip">
		<PluginCode bind:code={code}></PluginCode>
	</div>
</div>
