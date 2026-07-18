<script lang="ts">
	import { page } from '$app/state';

	const role = page.url.searchParams.get('role') ?? 'user';
	const isAdmin = role === 'admin';
</script>

<svelte:head>
	<title>Dashboard • Toggly</title>
</svelte:head>

<div class="page-shell">
	<div class="card">
		<p class="eyebrow">{isAdmin ? 'Administrator view' : 'Member view'}</p>
		<h1>Welcome, {role}</h1>
		<p class="subtext">You’ve been redirected here after a successful sign-in.</p>

		<div class="stat-grid">
			<div class="stat">
				<span class="label">Access level</span>
				<strong>{isAdmin ? 'Full administration' : 'Standard access'}</strong>
			</div>
			<div class="stat">
				<span class="label">Status</span>
				<strong>{isAdmin ? 'Ready to manage' : 'Ready to work'}</strong>
			</div>
		</div>
	</div>
</div>

<style>
	.page-shell {
		display: grid;
		place-items: center;
		min-height: 100vh;
		padding: 2rem;
		background: linear-gradient(135deg, oklch(0.96 0.015 300) 0%, oklch(0.99 0.008 240) 100%);
	}

	.card {
		width: min(100%, 34rem);
		padding: 2rem;
		border-radius: 1.5rem;
		background: white;
		box-shadow: 0 24px 60px oklch(0.25 0.04 300 / 0.14);
		border: 1px solid oklch(0.92 0.02 300);
	}

	.eyebrow {
		margin: 0 0 0.4rem;
		font-size: 0.8rem;
		font-weight: 700;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: oklch(0.62 0.18 300);
	}

	h1 {
		margin: 0;
		font-size: 2rem;
		color: oklch(0.28 0.08 300);
	}

	.subtext {
		margin: 0.6rem 0 1.2rem;
		color: oklch(0.45 0.04 240);
	}

	.stat-grid {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
	}

	.stat {
		padding: 1rem;
		border-radius: 1rem;
		background: oklch(0.98 0.008 300);
		border: 1px solid oklch(0.9 0.01 300);
		display: grid;
		gap: 0.4rem;
	}

	.label {
		font-size: 0.8rem;
		font-weight: 700;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: oklch(0.55 0.12 240);
	}
</style>
