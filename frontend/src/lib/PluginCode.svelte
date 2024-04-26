<script lang="ts">
	import Highlight, { LineNumbers } from 'svelte-highlight';
	import typescript from 'svelte-highlight/languages/typescript';
	import { a11yDark, darktooth, githubDark, icyDark } from 'svelte-highlight/styles';
	import github from 'svelte-highlight/styles/github';
	import { notify } from './notificationHelper';

	export let code = `//loading...`;

	let modal: HTMLDialogElement;

	const copy = async () => {
		if (!navigator.clipboard) {
			notify.error('Copy failed.');
		}
		try {
			await navigator.clipboard.writeText(code);
			
			notify.success('Copied Plugin to clipboad');
		} catch (error) {
			notify.error('Copy failed.');
		}
	};
</script>

<svelte:head>
	{@html githubDark}
</svelte:head>

<div>
	<Highlight language={typescript} bind:code={code} let:highlighted>
		<LineNumbers {highlighted} hideBorder wrapLines />
	</Highlight>
	<div class="z-40 absolute top-0 right-0 m-2">
		<button
			type="submit"
			class="btn"
			on:click={() => {
				modal.showModal();
			}}>How To Install</button
		>
		<button type="submit" class="btn" on:click={copy}>Copy</button>
	</div>
</div>

<dialog id="my_modal_1" class="modal" bind:this={modal}>
	<div class="modal-box">
		<h3 class="font-bold text-lg">Installing a plugin</h3>
		<p class="py-4">
			To install a plugin go to your <mark class="text-accent bg-transparent">user settings</mark>.
			In the settings you should see a section for
			<mark class="text-accent bg-transparent">Plugins</mark>. Click that and click
			<mark class="text-accent bg-transparent">Install plugins</mark>.
			<mark class="text-accent bg-transparent">Paste in the code here and click install</mark>. It
			<mark class="text-accent bg-transparent">may ask for some permissions, accept these</mark> and the plugin should work.
		</p>
		<div class="modal-action">
			<form method="dialog">
				<!-- if there is a button in form, it will close the modal -->
				<button class="btn">Close</button>
			</form>
		</div>
	</div>
</dialog>
