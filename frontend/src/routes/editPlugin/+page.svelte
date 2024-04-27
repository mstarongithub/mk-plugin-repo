<script lang="ts">
	import Navbar from '$lib/Navbar.svelte';
	import SettingsField from '$lib/components/SettingsField.svelte';
	import TagField from '$lib/components/TagField.svelte';

	let tab = 0;

	let TAB_IDS = {
		BASIC: 0,
		UPDATE_PLUGIN: 1
	};
</script>

<Navbar></Navbar>

<div class="flex flex-col gap-2 justify-center items-center">
	<!-- <div
		class="card flex flex-row items-center justify-between !w-10/12 !h-3/4 bg-base-100 shadow-xl overflow-clip p-4"
	>
		<div class="card join join-vertical">
			<button class="btn btn-wide join-item">Basic Settings</button>

			<button class="btn btn-wide join-item">Update Plugin</button>
		</div>
		<div class="card items-center w-full space-y-3">
            <div class="card">
                <button class="btn">Basic Settings</button>
                
                <button class="btn">Update Plugin</button>
            </div>
        </div>
	</div> -->

	<div class="card flex min-w-96 w-10/12 h-3/4 m-10 p-3 bg-base-200 shadow-xl lg:card-side">
		<figure>
			<div class="card join join-vertical">
				<button
					class="btn btn-wide join-item"
					on:click={() => {
						tab = TAB_IDS.BASIC;
					}}>Plugin Page Settings</button
				>

				<button
					class="btn btn-wide join-item"
					on:click={() => {
						tab = TAB_IDS.UPDATE_PLUGIN;
					}}>Update Plugin</button
				>
			</div>
		</figure>
		{#if tab === TAB_IDS.BASIC}
			<div class="card-body p-10 grow">
				<!-- <h2 class="card-title">Edit Plugin Page</h2> -->

				<SettingsField
					title="Name of the plugin"
					subtitle="This is one of the first things people will see"
				>
					<input type="text" placeholder="Plugin" class="input input-bordered w-full max-w-xs" />
				</SettingsField>

				<SettingsField title="Short Description" subtitle="In one sentance explain what it does.">
					<!-- <input type="text" placeholder="Plugin" class="input input-bordered w-full max-w-xs" /> -->
					<textarea class="textarea max-w-xs" placeholder="Short Description"></textarea>
				</SettingsField>

				<SettingsField title="Tags" subtitle="Select some tags that represent your plugin">
					<TagField className="!max-w-xs"></TagField>
				</SettingsField>

				<div class="card-actions justify-center">
					<button class="btn btn-primary">Save</button>
				</div>
			</div>
		{:else if tab === TAB_IDS.UPDATE_PLUGIN}
			<div class="card-body p-10 grow">
				<h2 class="card-title">Update plugin</h2>

				<div class="card-actions justify-center">
					<button class="btn btn-primary">Save</button>
				</div>
			</div>
		{/if}
	</div>
</div>
