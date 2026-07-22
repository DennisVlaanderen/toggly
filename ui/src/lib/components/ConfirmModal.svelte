<script lang="ts">
	interface Props {
		title: string;
		description?: string;
		confirmLabel: string;
		cancelLabel: string;
		variant?: 'danger' | 'default';
		onconfirm: () => void;
		oncancel?: () => void;
	}

	let {
		title,
		description,
		confirmLabel,
		cancelLabel,
		variant = 'default',
		onconfirm,
		oncancel
	}: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let confirmed = false;

	export function show() {
		confirmed = false;
		dialogEl?.showModal();
	}

	function requestConfirm() {
		confirmed = true;
		dialogEl?.close();
	}

	function requestCancel() {
		dialogEl?.close();
	}

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === dialogEl) {
			requestCancel();
		}
	}

	// `close()` fires the native `close` event for every dismissal path --
	// the Confirm click above, the Cancel click, and the Escape key -- so
	// `confirmed` is the single source of truth for which one happened,
	// rather than wiring up separate handlers that could double-fire.
	function handleClose() {
		if (confirmed) {
			onconfirm();
		} else {
			oncancel?.();
		}
	}
</script>

<dialog
	bind:this={dialogEl}
	onclose={handleClose}
	onclick={handleBackdropClick}
	class="m-auto rounded-3xl border border-brand-100 bg-white p-0 shadow-xl backdrop:bg-brand-900/50"
>
	<div class="grid gap-4 p-6 sm:w-96">
		<div class="grid gap-1.5">
			<h2 class="text-lg font-bold text-brand-900">{title}</h2>
			{#if description}
				<p class="text-sm text-accent-900/70">{description}</p>
			{/if}
		</div>
		<div class="flex justify-end gap-3">
			<button
				type="button"
				class="cursor-pointer rounded-full px-4 py-2 text-sm font-semibold text-brand-800 hover:bg-accent-50"
				onclick={requestCancel}
			>
				{cancelLabel}
			</button>
			<button
				type="button"
				class="cursor-pointer rounded-full px-4 py-2 text-sm font-bold text-white {variant ===
				'danger'
					? 'bg-error-600 hover:bg-error-700'
					: 'bg-brand-600 hover:bg-brand-700'}"
				onclick={requestConfirm}
			>
				{confirmLabel}
			</button>
		</div>
	</div>
</dialog>
