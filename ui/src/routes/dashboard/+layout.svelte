<script lang="ts">
	import { page } from '$app/state';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import { toast } from '$lib/toast.svelte';
	import { m } from '$lib/paraglide/messages.js';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	$effect(() => {
		if (page.data.flashReason === 'route-not-found') {
			toast.show(m.toast_route_not_found());
		}
	});
</script>

<div class="flex h-screen overflow-hidden bg-gradient-to-br from-brand-50 to-accent-50">
	<Sidebar
		flags={data.flags}
		username={data.username}
		isAdmin={data.isAdmin}
		permissions={data.permissions}
	/>
	<main class="min-w-0 flex-1 overflow-y-auto">
		{@render children()}
	</main>
</div>

<Toast />
