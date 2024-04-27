<script lang="ts">
	import type { Plugin } from '$lib';
	import Navbar from '$lib/Navbar.svelte';
	import PluginCode from '$lib/PluginCode.svelte';
	import PluginListing from '$lib/PluginListing.svelte';
	import {
		getAIscriptPermissions,
		getAIscriptVersion,
		getCodeWarnings
	} from '$lib/aiScriptCodeParsers';
	import { BASE_DIR } from '$lib/baseDir';
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import { notify } from '$lib/notificationHelper';
	import VersionSelectBar from '$lib/components/VersionSelectBar.svelte';

	let selectedPluginData: Plugin | undefined = undefined;
	let code: string;
	let aiscriptVersion: string = '...';
	let permissions: string[] = [];
	let warnings = '';
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
				notify.error('Server Error');
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

				if (aiscriptVersion === '') {
					aiscriptVersion = getAIscriptVersion(code) ?? '';
				}

				permissions = getAIscriptPermissions(code) ?? [];

				warnings = (getCodeWarnings(code) ?? []).join(', ');
				console.log(permissions);
			} else {
				let err = await response;
				console.error(err);
				notify.error('Server Error');
			}
		}
	};
</script>

<Navbar></Navbar>

<div class="flex flex-col gap-4 justify-center items-center">
	<!-- grid-cols-3 -->
	{#if selectedPluginData}
		<PluginListing
			className="lg:card-side !w-10/12 !h-3/4"
			pluginName={selectedPluginData.name}
			shortDesc={selectedPluginData.summary_long}
			tags={selectedPluginData.tags}
			showExtraOptions={true}
		></PluginListing>
	{/if}

	<VersionSelectBar
		on:change={showCode}
		bind:selectedVersion
		{aiscriptVersion}
		{permissions}
		bind:selectedPluginData
	></VersionSelectBar>

	{#if warnings.length > 0}
		<div
			class="alert alert-warning flex flex-row items-center !w-10/12 !h-3/4 shadow-xl overflow-clip p-4"
		>
			<Icon icon="fluent:warning-28-filled" width={'2em'} height={'2em'} />
			<p>{warnings}</p>
		</div>
	{/if}

	<div class="card !w-10/12 !h-3/4 bg-base-200 shadow-xl overflow-clip">
		<PluginCode bind:code></PluginCode>
	</div>
</div>
