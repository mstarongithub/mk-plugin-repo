<script>
	import Navbar from "$lib/Navbar.svelte";
import PluginListing from "$lib/PluginListing.svelte";
import ThemeSwitcher from "$lib/ThemeSwitcher.svelte";

</script>
<!-- <nav class="bg-white border-gray-200 dark:bg-gray-900">
    
</nav> -->

<Navbar></Navbar>

<!-- <h1 class="text-3xl text-red-600 font-bold underline">
    Plugins here
</h1> -->

<div class="flex justify-center items-center">
  <!-- grid-cols-3 -->
  <div class="w-10/12 h-screen flex flex-wrap gap-2">
    <PluginListing></PluginListing>
    <PluginListing></PluginListing>
    <PluginListing></PluginListing>
    <PluginListing></PluginListing>
    <PluginListing></PluginListing>
    <PluginListing></PluginListing>
    <PluginListing></PluginListing>
  </div>
</div>
