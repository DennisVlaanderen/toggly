<script lang="ts">
	import { toast } from '$lib/toast.svelte';
	import { m } from '$lib/paraglide/messages.js';
</script>

{#if toast.message}
	<div
		class="fixed inset-x-0 bottom-6 z-40 flex justify-center px-4"
		role="status"
		aria-live="polite"
	>
		<div
			class="flex max-w-md items-center gap-3 rounded-2xl border border-info-200 bg-info-50 px-4 py-3 text-sm font-semibold text-info-800 shadow-xl"
		>
			<span class="icon-[lucide--info] size-4 shrink-0" aria-hidden="true"></span>
			<p class="flex-1">{toast.message}</p>
			<button
				type="button"
				class="icon-[lucide--x] size-4 shrink-0 cursor-pointer text-info-600 hover:text-info-800"
				aria-label={m.toast_dismiss()}
				onclick={() => toast.dismiss()}
			></button>
		</div>
	</div>
{/if}
