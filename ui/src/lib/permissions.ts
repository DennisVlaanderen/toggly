// A structural (not imported) type -- lib/server/auth's Session satisfies
// this, but this module intentionally doesn't import from lib/server so it
// stays safe to use from client-reachable code (e.g. Sidebar.svelte)
// without tripping SvelteKit's server-only import boundary.
export interface PermissionAware {
	isAdmin: boolean;
	permissions: string[];
}

// The Admin bypass is unconditional, mirroring the backend's
// auth.Service.Resolve: an admin is never blocked by a missing permission
// string, regardless of what a group is or isn't configured with.
export function hasPermission(session: PermissionAware, perm: string): boolean {
	return session.isAdmin || session.permissions.includes(perm);
}
