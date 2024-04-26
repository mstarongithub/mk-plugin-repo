<script lang="ts">
	import { goto } from '$app/navigation';
	import Navbar from '$lib/Navbar.svelte';
	import { BASE_DIR } from '$lib/baseDir';
	import TagField from '$lib/components/TagField.svelte';
	import toast from 'svelte-french-toast';

	interface PluginData {
		name: string;
		summary_short: string;
		summary_long: string;
		initial_version: string;
		tags: string[];
		type: 'plugin' | 'widget' | undefined; // Type should be either "plugin" or "widget"
		code: string;
		aiscript_version: string;
	}

	let newPluginData: PluginData = {
		name: '',
		summary_short: '',
		summary_long: '',
		initial_version: '',
		tags: [],
		type: undefined,
		code: '',
		aiscript_version: ''
	};

	const ERRORS = {
		INVALID_CODE: () => {
			return toast.error('Please insert valid AIscript code', {
				className: '!btn'
			});
		},
		MISSING_FIELDS: () => {
			return toast.error('Please fill out all the fields', {
				className: '!btn'
			});
		}
	};

	function getPluginName(str: string): string | null {
		const regex = /###\s*{\s*.*name:\s*"([^"]*)".*\s*}/s;
		const match = str.match(regex);
		if (match && match.length > 1) {
			return match[1].trim();
		}
		return null;
	}

	function getPluginDesc(str: string): string | null {
		const regex = /###\s*{\s*.*description:\s*"([^"]*)".*\s*}/s;
		const match = str.match(regex);
		if (match && match.length > 1) {
			return match[1].trim();
		}
		return null;
	}

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

	function getAIscriptVersion(str: string): string | null {
		const regex = /^\/\/\/ @ (.*)/m;
		const match = str.match(regex);
		if (match && match.length > 1) {
			return match[1].trim();
		}
		return null;
	}

	function getPluginVersion(str: string): string | null {
		const regex = /###\s*{\s*.*version:\s*"([^"]*)".*\s*}/s;
		const match = str.match(regex);
		if (match && match.length > 1) {
			return match[1].trim();
		}
		return null;
	}

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

			toast.success("Plugin submitted", {
				className: '!btn'
			});

			goto("/");
		} else {
			let err = await response;
			console.error(err);
			
			toast.error("Server Error", {
				className: '!btn'
			});
		}
	};
</script>

<Navbar></Navbar>

<div class="flex justify-center items-center">
	<!-- grid-cols-3 -->
	<div class="flex lg:w-1/3">
		<div class="card items-center w-full space-y-3">
			<input
				type="text"
				placeholder="Name"
				class="input input-bordered w-full max-w-xs"
				bind:value={newPluginData.name}
			/>

			<select class="select select-bordered w-full max-w-xs" bind:value={newPluginData.type}>
				<option disabled selected value="">What type of script is it?</option>
				<option value="plugin">Plugin</option>
				<option value="widget">Widget</option>
			</select>

			<label class="form-control w-full max-w-xs">
				<div class="label">
					<span class="label-text">Paste the code here</span>
				</div>
				<textarea
					class="textarea textarea-primary"
					placeholder="Code"
					bind:value={newPluginData.code}
					on:change={codeEdited}
				></textarea>
			</label>

			<label class="form-control w-full max-w-xs">
				<div class="label">
					<span class="label-text">A short sentence on what the plugin does</span>
				</div>
				<textarea
					class="textarea textarea-primary"
					placeholder="Short Description"
					bind:value={newPluginData.summary_short}
				></textarea>
			</label>

			<label class="form-control w-full max-w-xs">
				<div class="label">
					<span class="label-text">Describe what the plugin does in detail</span>
				</div>
				<textarea
					class="textarea textarea-primary"
					placeholder="Long Description"
					bind:value={newPluginData.summary_long}
				></textarea>
			</label>

			<!-- svelte-ignore a11y-label-has-associated-control -->
			<label class="form-control w-full max-w-xs">
				<div class="label">
					<span class="label-text">Tags</span>
				</div>
				<TagField bind:tags={newPluginData.tags}></TagField>
			</label>

			<label class="form-control w-full max-w-xs">
				<div class="label">
					<span class="label-text">Upload a cover image</span>
				</div>
				<input type="file" class="file-input file-input-bordered w-full max-w-xs" />
			</label>

			<button class="btn btn-primary" on:click={submit}>Submit Plugin</button>
		</div>
	</div>
</div>
