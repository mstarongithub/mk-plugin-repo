<script lang="ts">
    export let className : string = ''
    export let title : string;
    export let subtitle : string = '';

</script>


<div class="card flex bg-base-200 lg:card-side {className}">
    <figure class="flex flex-col p-10 md:w-1/3">
        <h3 class="text-xl font-semibold">{title}</h3>
        <p class="text-neutral-500 font-sans text-wrap">{subtitle}</p>
    </figure>
    <div class="card-body p-10 grow">
        <slot></slot>
    </div>
</div>