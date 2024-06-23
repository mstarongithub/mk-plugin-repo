<script lang="ts">
	import type { Plugin } from '$lib';
	import Navbar from '$lib/Navbar.svelte';
	import { getAIscriptVersion, getPluginVersion } from '$lib/aiScriptCodeParsers';
	import { BASE_DIR } from '$lib/baseDir';
	import EditBasicPluginData from '$lib/components/EditBasicPluginData.svelte';
	import PluginCode from '$lib/components/PluginCode.svelte';
	import SettingsField from '$lib/components/SettingsField.svelte';
	import TagField from '$lib/components/TagField.svelte';
	import { notify } from '$lib/notificationHelper';
	import Icon from '@iconify/svelte';
	import { plugin } from 'postcss';
	import { onMount } from 'svelte';

	let tab = 0;

	let TAB_IDS = {
		BASIC: 0,
		UPDATE_PLUGIN: 1,
		MANAGE_HISTORY: 2,
		DELETE: 3
	};

	let pluginId: string = '';

	let metadata: {
		name: string;
		summary_short: string;
		summary_long: string;
		tags: string[];
		type: 'plugin' | 'widget';
	} = {
		name: '',
		summary_short: '',
		summary_long: '',
		tags: [],
		type: 'plugin'
	};

	let newCode: {
		code: string;
		aiscript_version: string;
		version_name: string;
	} = {
		code: '',
		aiscript_version: '',
		version_name: ''
	};

	let versionHistory: string[] = [];

	//   - `code`: `string` - The full code of this version
	//   - `aiscript_version`: `string` - The version of AIScript this plugin is intended for
	//   - `version_name`: `string` - The name of the version

	const updateMetadata = async () => {
		let response = await fetch(`${BASE_DIR}/api/v1/plugins/${pluginId}`, {
			body: JSON.stringify(metadata),
			headers: {
				'Content-Type': 'application/json'
			},
			method: 'PUT'
		});

		if (response.ok) {
			notify.success('Saved');
		} else {
			let err = await response;
			console.error(err);
			notify.error('Server Error');
		}
	};

	const publishCodeVersion = async () => {
		let response = await fetch(`${BASE_DIR}/api/v1/plugins/${pluginId}`, {
			body: JSON.stringify(newCode),
			headers: {
				'Content-Type': 'application/json'
			},
			method: 'POST'
		});

		if (response.ok) {
			notify.success('Saved');
		} else {
			let err = await response;
			console.log(newCode);
			console.error(err);
			notify.error('Server Error');
		}
	};

	const updateVersionInfo = () => {
		newCode.aiscript_version = getAIscriptVersion(newCode.code) ?? '';
		newCode.version_name = getPluginVersion(newCode.code) ?? '';
	};

	async function loadSelectedVersion(version: string) {
		let response = await fetch(`${BASE_DIR}/api/v1/plugins/${pluginId}/${version}`, {
			headers: {
				'Content-Type': 'application/json'
			},
			method: 'GET'
		});
		if (response.ok) {
			return await response.json();
		} else {
			let err = await response;
			console.error(err);
			notify.error('Server Error');
		}
	}

	async function deleteSelectedVersion(version: string) {
		let response = await fetch(`${BASE_DIR}/api/v1/plugins/${pluginId}/${version}`, {
			headers: {
				'Content-Type': 'application/json'
			},
			method: 'DELETE'
		});
		if (response.ok) {
			notify.success(`Deleted version ${version}`);
			return await response.json();
		} else {
			let err = await response;
			console.error(err);
			notify.error('Server Error');
		}
	}

	async function deletePlugin() {
		let response = await fetch(`${BASE_DIR}/api/v1/plugins/${pluginId}`, {
			headers: {
				'Content-Type': 'application/json'
			},
			method: 'DELETE'
		});
		if (response.ok) {
			notify.success(`Deleted Plugin`);

			window.setTimeout(()=>{
				window.location.href = '/';
			}, 1000);
			return await response.json();
		} else {
			let err = await response;
			console.error(err);
			notify.error('Server Error');
		}
	}

	let selectedVersion = '';

	function expandSelectedVersion(version: string) {
		selectedVersion = version;
	}

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
				const selectedPluginData: Plugin = await response.json();
				metadata.name = selectedPluginData.name;
				metadata.summary_short = selectedPluginData.summary_short;
				metadata.summary_long = selectedPluginData.summary_long;
				metadata.tags = selectedPluginData.tags;
				metadata.type = selectedPluginData.type;

				versionHistory = selectedPluginData.all_versions;
				newCode.version_name = selectedPluginData.current_version;

				newCode = {
					...newCode,
					...(await loadSelectedVersion(selectedPluginData.current_version))
				};

				updateVersionInfo();
			} else {
				let err = await response;
				console.error(err);
				notify.error('Server Error');
			}
		}
	});
</script>

<svelte:head>
	<title>Edit {metadata.name} - *Key Plugin Repo</title>
</svelte:head>

<Navbar></Navbar>

<div class="flex flex-col gap-2 justify-center items-center">
	<div class="card flex min-w-96 w-10/12 h-3/4 m-10 p-3 bg-base-200 shadow-xl lg:card-side">
		<figure>
			<ul class="menu menu-md bg-base-200 w-56 rounded-box">
				<li>
					<button
						class="capitalize"
						on:click={() => {
							tab = TAB_IDS.BASIC;
						}}>Page Settings</button
					>
				</li>
				<li>
					<button
						class="capitalize"
						on:click={() => {
							tab = TAB_IDS.UPDATE_PLUGIN;
						}}>Update {metadata.type}</button
					>
				</li>
				<li>
					<button
						class="capitalize"
						on:click={() => {
							tab = TAB_IDS.MANAGE_HISTORY;
						}}>Manage Update History</button
					>
				</li>
				<li>
					<button
						class="capitalize"
						on:click={() => {
							tab = TAB_IDS.DELETE;
						}}>Danger Zone</button
					>
				</li>
			</ul>
		</figure>
		{#if tab === TAB_IDS.BASIC}
			<div class="card-body p-10 grow">
				<!-- <h2 class="card-title">Edit Plugin Page</h2> -->

				<EditBasicPluginData bind:name={metadata.name} bind:summary_short={metadata.summary_short} bind:summary_long={metadata.summary_long} bind:tags={metadata.tags} bind:type={metadata.type}></EditBasicPluginData>

				<div class="card-actions justify-center">
					<button class="btn btn-primary" on:click={updateMetadata}>Save</button>
				</div>
			</div>
		{:else if tab === TAB_IDS.UPDATE_PLUGIN}
			<div class="card-body p-10 grow">
				<h2 class="card-title capitalize">Update {metadata.type}</h2>

				<SettingsField title="Code" subtitle="The full code that runs this {metadata.type}">
					<!-- <input type="text" placeholder="Plugin" class="input input-bordered w-full max-w-xs" /> -->
					<textarea
						class="textarea"
						placeholder="Code Here"
						bind:value={newCode.code}
						on:change={updateVersionInfo}
					></textarea>
				</SettingsField>

				<SettingsField
					title="Info"
					subtitle={`New plugin version: ${newCode.version_name}\nAIscript version: ${newCode.aiscript_version}`}
				></SettingsField>

				<div class="card-actions justify-center">
					<button class="btn btn-primary" on:click={publishCodeVersion}>Save</button>
				</div>
			</div>
		{:else if tab === TAB_IDS.MANAGE_HISTORY}
			<div class="card-body p-10 grow">
				<div class="overflow-x-auto">
					<table class="table table-zebra">
						<thead>
							<tr>
								<th>Version</th>
							</tr>
						</thead>
						<tbody>
							<!-- row 1 -->
							{#each versionHistory.reverse() as versionHistoryEntry}
								<!-- viewUpdateHistoryCode -->
								<tr class="w-full">
									<div class="content-between w-full flex items-center">
										<th class="grow">{versionHistoryEntry}</th>
										<td class="flex justify-end">
											<button
												class="btn btn-primary"
												on:click={() => {
													window.open(`/plugin?id=${pluginId}&v=${versionHistoryEntry}`, '_blank');
													// console.log(versionHistoryEntry)
													// expandSelectedVersion(versionHistoryEntry);
												}}
											>
												View Code
											</button>

											<button
												class="btn"
												on:click={() => {
													deleteSelectedVersion(versionHistoryEntry);
												}}
											>
												<Icon icon="mdi:delete-outline" width={'2em'} height={'2em'} />
											</button></td
										>
									</div>
									<!-- {#if "1.0.1" == versionHistoryEntry}
										<PluginCode showActions={false} code={selectedVersion}></PluginCode>
									{/if} -->
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{:else if tab === TAB_IDS.DELETE}
			<div class="card-body p-10 grow">
				<h2 class="card-title capitalize">Delete {metadata.type}</h2>

				<SettingsField title="Delete" subtitle="This action cannot be undone">
					<!-- <input type="text" placeholder="Plugin" class="input input-bordered w-full max-w-xs" /> -->
					<button class="btn btn-error" on:click={deletePlugin}>Delete</button>
					
				</SettingsField>
			</div>
		{/if}
	</div>
</div>
