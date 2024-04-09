<script lang="ts">
	import Navbar from "$lib/Navbar.svelte";
</script>

<Navbar></Navbar>

<div class="flex justify-center items-center">
  <!-- grid-cols-3 -->
  <div class="flex lg:w-1/3">
    <div class="card items-center w-full space-y-3">
      <input type="text" placeholder="Name" class="input input-bordered w-full max-w-xs" />

      <div class="dropdown">
        <div tabindex="0" role="button" class="btn m-1">What type of script is it?</div>
        <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
          <li><a>Plugin</a></li>
          <li><a>Widget</a></li>
        </ul>
      </div>

      <label class="form-control w-full max-w-xs">
        <div class="label">
          <span class="label-text">Paste the code here</span>
        </div>
        <textarea class="textarea textarea-primary" placeholder="Code"></textarea>
      </label>

      <label class="form-control w-full max-w-xs">
        <div class="label">
          <span class="label-text">Describe what the plugin does</span>
        </div>
        <textarea class="textarea textarea-primary" placeholder="Description"></textarea>
      </label>

      <label class="form-control w-full max-w-xs">
        <div class="label">
          <span class="label-text">Upload a cover image</span>
        </div>
        <input type="file" class="file-input file-input-bordered w-full max-w-xs" />
      </label>
      
      <button class="btn btn-primary">Submit Plugin</button>
    </div>
  </div>
</div>
