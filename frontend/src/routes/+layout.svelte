<script>
    import "../app.css";

    import { onMount } from 'svelte';
	import toast, { Toaster } from 'svelte-french-toast';

</script>
  
<Toaster />
<slot />