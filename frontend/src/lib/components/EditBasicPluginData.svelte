<script lang="ts">

	import SettingsField from "./SettingsField.svelte";
	import TagField from "./TagField.svelte";

    // export let metadata : {
	// 	name: string;
	// 	summary_short: string;
	// 	summary_long: string;
	// 	tags: string[];
	// 	type: 'plugin' | 'widget';
	// } 

	export let name : string;
	export let summary_short: string;
	export let summary_long: string;
	export let tags: string[];
	export let type : 'plugin' | 'widget'
</script>
<SettingsField
	title="Name of the {type}"
	subtitle="This is one of the first things people will see"
>
	<input
		type="text"
		placeholder="Amazing {type}"
		class="input input-bordered w-full max-w-xs capitalize"
		bind:value={name}
	/>
</SettingsField>

<SettingsField title="Type" subtitle="Is this a plugin or widget?">
	<!-- <input type="text" placeholder="Plugin" class="input input-bordered w-full max-w-xs" /> -->
	<select class="select select-bordered w-full max-w-xs" bind:value={type}>
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
		bind:value={summary_short}
	></textarea>
</SettingsField>

<SettingsField
	title="Long Description"
	subtitle="Explain how to use this {type} and all its features."
>
	<!-- <input type="text" placeholder="Plugin" class="input input-bordered w-full max-w-xs" /> -->
	<textarea
		class="textarea max-w-xs"
		placeholder="Long Description"
		bind:value={summary_long}
	></textarea>
</SettingsField>

<SettingsField title="Tags" subtitle="Select some tags that represent your {type}">
	<TagField className="!max-w-xs" bind:tags={tags}></TagField>
</SettingsField>
