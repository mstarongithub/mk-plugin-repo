<script>
	import Navbar from '$lib/Navbar.svelte';
	import PluginCode from '$lib/PluginCode.svelte';
	import PluginListing from '$lib/PluginListing.svelte';
    
</script>

<Navbar></Navbar>

<div class="flex flex-col gap-2 justify-center items-center">
	<!-- grid-cols-3 -->
		<PluginListing className="lg:card-side !w-10/12 !h-3/4"></PluginListing>

        <div class = "card !w-10/12 !h-3/4 bg-base-100 shadow-xl overflow-clip">
            <PluginCode></PluginCode>
        </div>
</div>
