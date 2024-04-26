<script lang="ts">
	import Icon from '@iconify/svelte';

	export let pluginName: string = 'Pronoun Plugin';
	export let shortDesc: string = 'Easily see everyones pronouns when you are looking at a post';
	export let tags: string[] = ['Plugin', 'Quality Of Life'];
	export let imageData: { imageURL: string; alt: string } = {
		imageURL: 'https://woem.men/files/d83c5477-97ed-44a3-9365-36985f0e4146',
		alt: 'Pic of pronoun plugin'
	};
	export let isNew: boolean = false;

	export let showExtraOptions: boolean = false;

	export let className = '';
</script>

<div class="card w-96 bg-base-200 shadow-xl {className}">
	<figure>
		<img src={imageData.imageURL} alt={imageData.alt} />
	</figure>
	<div class="card-body">
		<h2 class="card-title">
			<p>
				{pluginName}
				{#if isNew}
					<div class="badge badge-secondary">NEW</div>
				{/if}
			</p>
			
			{#if showExtraOptions}
				<details class="dropdown dropdown-end">
					<summary class="m-1 btn"><Icon icon="charm:menu-meatball" /></summary>
					<ul class="p-2 shadow menu dropdown-content z-[1] bg-base-100 rounded-box w-52">
						<li><button>Edit</button></li>
					</ul>
				</details>
			{/if}
		</h2>
		<p>{shortDesc}</p>
		<div class="card-actions justify-end">
			{#each tags as tag}
				<div class="badge badge-outline">{tag}</div>
			{/each}
		</div>
	</div>
</div>
