<script lang="ts">
	import { goto } from '$app/navigation';
	import Navbar from '$lib/Navbar.svelte';
	import {
		getAIscriptVersion,
		getPluginDesc,
		getPluginName,
		getPluginVersion
	} from '$lib/aiScriptCodeParsers';
	import { BASE_DIR } from '$lib/baseDir';
	import EditBasicPluginData from '$lib/components/EditBasicPluginData.svelte';
	import SettingsField from '$lib/components/SettingsField.svelte';
	import TagField from '$lib/components/TagField.svelte';
	import { notify } from '$lib/notificationHelper';

	interface PluginData {
		name: string;
		summary_short: string;
		summary_long: string;
		initial_version: string;
		tags: string[];
		type: 'plugin' | 'widget'; // Type should be either "plugin" or "widget"
		code: string;
		aiscript_version: string;
	}

	let newPluginData: PluginData = {
		name: '',
		summary_short: '',
		summary_long: '',
		initial_version: '',
		tags: [],
		type: 'plugin',
		code: '',
		aiscript_version: ''
	};

	const ERRORS = {
		INVALID_CODE: () => {
			return notify.error('Please insert valid AIscript code');
		},
		MISSING_FIELDS: () => {
			return notify.error('Please fill out all the fields');
		}
	};

	let codeEdited = () => {
		const { code, name, summary_short, summary_long } = newPluginData;

		if (name.length == 0) {
			newPluginData.name = getPluginName(code) ?? '';
		}

		if (summary_short.length == 0) {
			const extractedDesc = getPluginDesc(code) ?? '';
			if (extractedDesc.length < 128) {
				newPluginData.summary_short = extractedDesc;
			}
		}
	};

	function parseVersionFromCode(code: string) {
		const aiscriptVersion = getAIscriptVersion(code);
		if (aiscriptVersion == null) return ERRORS.INVALID_CODE();
		newPluginData.aiscript_version = aiscriptVersion;

		const pluginVersion = getPluginVersion(code);
		if (aiscriptVersion == null) return ERRORS.INVALID_CODE();
		newPluginData.initial_version = pluginVersion as string;
	}

	let submit = async () => {
		const { code } = newPluginData;
		if (code.trim().length < 1) return ERRORS.INVALID_CODE();

		const codeParseErrors = parseVersionFromCode(code.trim());

		if (codeParseErrors != undefined) {
			return codeParseErrors;
		}

		for (let key in newPluginData) {
			if (newPluginData.hasOwnProperty(key)) {
				const value = newPluginData[key as keyof PluginData];

				if (typeof value === 'string' && !value.trim()) {
					console.log(`Field '${key}' is blank or empty.`);
					return ERRORS.MISSING_FIELDS();
				} else if (Array.isArray(value) && value.length === 0) {
					if (key !== 'tags') {
						console.log(`Array field '${key}' is empty.`);
						return ERRORS.MISSING_FIELDS();
					}
				}
			}
		}

		let response = await fetch(`${BASE_DIR}/api/v1/plugins`, {
			body: JSON.stringify(newPluginData),
			headers: {
				'Content-Type': 'application/json'
			},
			method: 'POST'
		});

		if (response.ok) {
			notify.success('Plugin submitted');

			goto('/');
		} else {
			let err = await response;
			console.error(err);

			notify.error('Server Error');
		}
	};
</script>

<svelte:head>
	<title>Submit new plugin - *Key Plugin Repo</title>
</svelte:head>

<Navbar></Navbar>

<div class="flex flex-col gap-2 justify-center items-center">
	<div class="card flex min-w-96 w-10/12 h-3/4 m-10 p-3 bg-base-200 shadow-xl lg:card-side">
		<div class="card-body p-10 grow !flex !justify-center items-center">
			<h1 class="text-4xl font-semibold capitalize text-center">Submit a new plugin</h1>
			<div class="w-fit">
				<EditBasicPluginData
					bind:name={newPluginData.name}
					bind:summary_short={newPluginData.summary_short}
					bind:summary_long={newPluginData.summary_long}
					bind:tags={newPluginData.tags}
					bind:type={newPluginData.type}
				></EditBasicPluginData>

				<SettingsField title="Code" subtitle="The full code that runs this {newPluginData.type}">
					<textarea class="textarea" placeholder="Code Here" bind:value={newPluginData.code}
					></textarea>
				</SettingsField>
			</div>

			<div class="card-actions justify-center">
				<button class="btn btn-primary" on:click={submit}>Submit Plugin</button>
			</div>
		</div>
	</div>
</div>
