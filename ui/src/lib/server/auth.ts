import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';
import type { Cookies } from '@sveltejs/kit';

export interface Session {
	id: string;
	username: string;
	isAdmin: boolean;
	permissions: string[];
}

const AUTH_COOKIE = 'aerendil.auth';
const API_ORIGIN = env.AERENDIL_API_ORIGIN?.trim() || 'http://127.0.0.1:8080';

function parseSession(payload: unknown): Session | null {
	if (typeof payload !== 'object' || payload === null) {
		return null;
	}
	const { user, isAdmin, permissions } = payload as {
		user?: unknown;
		isAdmin?: unknown;
		permissions?: unknown;
	};
	if (typeof user !== 'object' || user === null) {
		return null;
	}
	const { id, username } = user as { id?: unknown; username?: unknown };
	if (
		typeof id !== 'string' ||
		typeof username !== 'string' ||
		typeof isAdmin !== 'boolean' ||
		!Array.isArray(permissions) ||
		!permissions.every((p) => typeof p === 'string')
	) {
		return null;
	}
	return { id, username, isAdmin, permissions };
}

export async function login(username: string, password: string): Promise<{ token: string } | null> {
	const response = await fetch(`${API_ORIGIN}/api/auth/login`, {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: JSON.stringify({ username, password })
	});

	const payload = await response.json().catch(() => null);
	if (!response.ok || typeof payload?.token !== 'string' || !payload.token) {
		return null;
	}

	return { token: payload.token };
}

// getSession always re-fetches /api/auth/me rather than trusting any local
// cache -- the backend resolves permissions fresh from the store on every
// call, so a group/permission change or deactivation is reflected on the
// user's very next request, with no token revocation infrastructure needed.
export async function getSession(cookies: Cookies): Promise<Session | null> {
	const token = cookies.get(AUTH_COOKIE);
	if (!token) {
		return null;
	}

	const response = await fetch(`${API_ORIGIN}/api/auth/me`, {
		headers: { Authorization: `Bearer ${token}` }
	});
	if (!response.ok) {
		return null;
	}

	return parseSession(await response.json().catch(() => null));
}

export function setAuthCookie(cookies: Cookies, token: string) {
	cookies.set(AUTH_COOKIE, token, {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		secure: !dev,
		maxAge: 60 * 60 * 24
	});
}

export function clearAuthCookie(cookies: Cookies) {
	cookies.delete(AUTH_COOKIE, { path: '/' });
}

export function getAuthToken(cookies: Cookies): string | null {
	return cookies.get(AUTH_COOKIE) ?? null;
}
