<script lang="ts">
	import type { Plugin } from "$lib";

    export let permissions: string[];
    export let selectedPluginData: Plugin | undefined;
    export let aiscriptVersion: string;

    export let selectedVersion: string;
</script>

<div
class="card flex flex-row items-center justify-between !w-10/12 !h-3/4 bg-base-200 shadow-xl overflow-clip p-4"
>
<div class="flex flex-row items-center">
    <select
        class="select select-bordered w-full max-w-40"
        bind:value={selectedVersion}
        on:change
        >
        {#if selectedPluginData}
            {#each selectedPluginData.all_versions as version}
                <option selected={version == selectedPluginData.current_version} value={version}
                    >V{version}</option
                >
            {/each}
        {/if}
    </select>
    {#if permissions.length > 0}
        <div class="pl-8 flex flex-row items-center gap-3">
            Permissions:
            {#each permissions as permission}
                <div class="badge badge-outline">{permission}</div>
            {/each}
        </div>
    {/if}
</div>

<!-- <div class = "align-middle h-full"> -->
<p class="text-center">For AIscript @{aiscriptVersion}</p>
<!-- </div> -->
</div>