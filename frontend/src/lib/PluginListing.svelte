<script lang="ts">
	export let pluginName: string = 'Pronoun Plugin';
	export let shortDesc: string = 'Easily see everyones pronouns when you are looking at a post';
	export let tags: string[] = ['Plugin', 'Quality Of Life'];
	export let imageData: { imageURL: string; alt: string } = {
		imageURL: 'https://woem.men/files/d83c5477-97ed-44a3-9365-36985f0e4146',
		alt: 'Pic of pronoun plugin'
	};
	export let isNew: boolean = false;
	export let id: number = -1;

	export let className = '';
</script>

{#if id != -1}
	<a href="/plugin?id={id}">
		<div class="card w-96 bg-base-100 shadow-xl {className}">
			<figure>
				<img src={imageData.imageURL} alt={imageData.alt} />
			</figure>
			<div class="card-body">
				<h2 class="card-title">
					{pluginName}
					{#if isNew}
						<div class="badge badge-secondary">NEW</div>
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
	</a>
{:else}
<!-- Its the same as above but without the <a> tag -->
	<div class="card w-96 bg-base-100 shadow-xl {className}">
		<figure>
			<img src={imageData.imageURL} alt={imageData.alt} />
		</figure>
		<div class="card-body">
			<h2 class="card-title">
				{pluginName}
				{#if isNew}
					<div class="badge badge-secondary">NEW</div>
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
{/if}
