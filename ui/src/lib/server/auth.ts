import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';
import type { Cookies } from '@sveltejs/kit';

export type UserRole = 'admin' | 'user';

export interface Session {
	username: string;
	role: UserRole;
}

const AUTH_COOKIE = 'toggly.auth';
const API_ORIGIN = env.TOGGLY_API_ORIGIN?.trim() || 'http://127.0.0.1:8080';

function parseUser(payload: unknown): Session | null {
	if (typeof payload !== 'object' || payload === null) {
		return null;
	}
	const user = (payload as { user?: unknown }).user;
	if (typeof user !== 'object' || user === null) {
		return null;
	}
	const { username, role } = user as { username?: unknown; role?: unknown };
	if (typeof username !== 'string' || (role !== 'admin' && role !== 'user')) {
		return null;
	}
	return { username, role };
}

export async function login(username: string, password: string): Promise<{ token: string; user: Session } | null> {
	const response = await fetch(`${API_ORIGIN}/api/auth/login`, {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: JSON.stringify({ username, password })
	});

	const payload = await response.json().catch(() => null);
	if (!response.ok || typeof payload?.token !== 'string' || !payload.token) {
		return null;
	}

	const user = parseUser(payload);
	if (!user) {
		return null;
	}

	return { token: payload.token, user };
}

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

	return parseUser(await response.json().catch(() => null));
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
