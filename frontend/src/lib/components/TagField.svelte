<script lang="ts">
	let cachedTags = ['Fun', 'Tool'] as string[];
	export let tags: string[] = [];

	let hasFocus = false;
	let inputEl: HTMLInputElement;
	let tempSearchQuery = '';

	function addSearchTag(tag: string) {
		tags = [...tags, tag];

		tempSearchQuery = '';
	}

	function deleteSearchTag(index: number) {
		tags.splice(index, 1);
		//This tells svelte to update the UI allowing it to be reactive
		tags = tags;
	}

	function getTagsToShow() {
		return cachedTags.filter(
			(a) => a.toLowerCase().includes(tempSearchQuery.toLowerCase()) && !tags.includes(a)
		);
	}

	function keyPressed(ev: KeyboardEvent) {
		if (ev.code === 'Enter') {
			addSearchTag(inputEl.value);
		}
	}
</script>

<div>
	<label class="flex gap-2 grow input input-bordered items-center">
		<input
			class="grow"
			placeholder={'Search Tags'}
			type="text"
			bind:this={inputEl}
			bind:value={tempSearchQuery}
			on:keypress={keyPressed}
			on:focusin={() => {
				hasFocus = true;
			}}
			on:focusout={() => {
				hasFocus = false;
			}}
		/>
	</label>

	<div class="flex">
		{#each tags as tag, i}
			<a
				class="badge hover:animate-pulse
        hover:line-through"
                aria-hidden="false"
				href="#tag"
				on:click={() => deleteSearchTag(i)}
			>
				{tag}
			</a>
		{/each}
	</div>

	<!-- Autocomplete help -->
	{#if tempSearchQuery.trim().length > 0 && getTagsToShow().length > 0}
		<ul
			class="
        bg-base-200
        grid
        grid-cols-3
        max-h-96
        max-w-80
        menu
        mt-1
        overflow-scroll
        rounded-box
      "
		>
			{#each getTagsToShow() as tag}
				<li>
					<button class="max-w-32" on:click={() => addSearchTag(tag)}>
						{tag}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
