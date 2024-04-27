<script lang="ts">
	import type { Plugin } from '$lib';
	import Navbar from '$lib/Navbar.svelte';
	import { getAIscriptVersion, getPluginVersion } from '$lib/aiScriptCodeParsers';
	import { BASE_DIR } from '$lib/baseDir';
	import PluginCode from '$lib/components/PluginCode.svelte';
	import SettingsField from '$lib/components/SettingsField.svelte';
	import TagField from '$lib/components/TagField.svelte';
	import { notify } from '$lib/notificationHelper';
	import Icon from '@iconify/svelte';
	import { onMount } from 'svelte';

	let tab = 0;

	let TAB_IDS = {
		BASIC: 0,
		UPDATE_PLUGIN: 1,
		MANAGE_HISTORY: 2
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
			</ul>
		</figure>
		{#if tab === TAB_IDS.BASIC}
			<div class="card-body p-10 grow">
				<!-- <h2 class="card-title">Edit Plugin Page</h2> -->

				<SettingsField
					title="Name of the {metadata.type}"
					subtitle="This is one of the first things people will see"
				>
					<input
						type="text"
						placeholder="Amazing {metadata.type}"
						class="input input-bordered w-full max-w-xs capitalize"
						bind:value={metadata.name}
					/>
				</SettingsField>

				<SettingsField title="Type" subtitle="Is this a plugin or widget?">
					<!-- <input type="text" placeholder="Plugin" class="input input-bordered w-full max-w-xs" /> -->
					<select class="select select-bordered w-full max-w-xs" bind:value={metadata.type}>
						<option disabled selected value="">What type of script is it?</option>
						<option value="plugin">Plugin</option>
						<option value="widget">Widget</option>
					</select>
				</SettingsField>

				<SettingsField title="Short Description" subtitle="In one sentance explain what it does.">
					<!-- <input type="text" placeholder="Plugin" class="input input-bordered w-full max-w-xs" /> -->
					<textarea
						class="textarea max-w-xs"
						placeholder="Short Description"
						maxlength="128"
						bind:value={metadata.summary_short}
					></textarea>
				</SettingsField>

				<SettingsField
					title="Long Description"
					subtitle="Explain how to use this {metadata.type} and all its features."
				>
					<!-- <input type="text" placeholder="Plugin" class="input input-bordered w-full max-w-xs" /> -->
					<textarea
						class="textarea max-w-xs"
						placeholder="Long Description"
						bind:value={metadata.summary_long}
					></textarea>
				</SettingsField>

				<SettingsField title="Tags" subtitle="Select some tags that represent your {metadata.type}">
					<TagField className="!max-w-xs" bind:tags={metadata.tags}></TagField>
				</SettingsField>

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
						placeholder="Long Description"
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
						<!-- head -->
						<thead>
							<tr>
								<!-- <th></th> -->
								<th>Version</th>
								<!-- <th>Job</th>
                          <th>Favorite Color</th> -->
							</tr>
						</thead>
						<tbody>
							<!-- row 1 -->
							{#each versionHistory.reverse() as versionHistoryEntry}
								<!-- viewUpdateHistoryCode -->
								<tr class="content-between w-full">
									<th>{versionHistoryEntry}</th>
									<td class="flex justify-end">
										<button class="btn btn-primary" on:click={()=>{
                                            notify.error('Not implemented yet');
                                        }}> View Code </button>

										<button
											class="btn"
											on:click={() => {
												deleteSelectedVersion(versionHistoryEntry);
											}}
										>
											<Icon icon="mdi:delete-outline" width={'2em'} height={'2em'} />
										</button></td
									>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	</div>
</div>
