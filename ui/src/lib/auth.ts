export type UserRole = 'admin' | 'user';

export type AuthResult =
	| { ok: true; role: UserRole; token: string }
	| { ok: false; role: null; message: string };

function parseJwtPayload(token: string): Record<string, unknown> | null {
	try {
		const payload = token.split('.')[1];
		if (!payload) {
			return null;
		}
		const normalized = payload.replace(/-/g, '+').replace(/_/g, '/');
		const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=');
		return JSON.parse(atob(padded)) as Record<string, unknown>;
	} catch {
		return null;
	}
}

function getCookie(name: string): string | null {
	if (typeof document === 'undefined') {
		return null;
	}

	const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
	return match ? decodeURIComponent(match[1]) : null;
}

function setCookie(name: string, value: string, days = 7) {
	if (typeof document === 'undefined') {
		return;
	}

	const expiry = new Date(Date.now() + days * 24 * 60 * 60 * 1000).toUTCString();
	document.cookie = `${name}=${encodeURIComponent(value)}; path=/; expires=${expiry}; SameSite=Lax`;
}

export function clearAuthCookie() {
	if (typeof document === 'undefined') {
		return;
	}
	document.cookie = 'toggly.auth=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
}

export async function authenticateUser(username: string, password: string): Promise<AuthResult> {
	const normalizedUsername = username.trim();
	const normalizedPassword = password.trim();

	if (!normalizedUsername || !normalizedPassword) {
		return {
			ok: false,
			role: null,
			message: 'Invalid username or password.'
		};
	}

	const response = await fetch('/api/auth/login', {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: JSON.stringify({ username: normalizedUsername, password: normalizedPassword })
	});

	const payload = await response.json().catch(() => ({}));
	if (!response.ok || typeof payload.token !== 'string' || !payload.token) {
		return { ok: false, role: null, message: payload.error ?? 'Invalid username or password.' };
	}

	setCookie('toggly.auth', payload.token);
	const claims = parseJwtPayload(payload.token);
	const role = claims?.role;

	if (role === 'admin' || role === 'user') {
		return { ok: true, role: role as UserRole, token: payload.token };
	}

	return { ok: false, role: null, message: 'Invalid username or password.' };
}

export function getStoredAuthToken(): string | null {
	return getCookie('toggly.auth');
}

export function getStoredUserRole(): UserRole | null {
	const token = getStoredAuthToken();
	if (!token) {
		return null;
	}

	const claims = parseJwtPayload(token);
	const role = claims?.role;
	return role === 'admin' || role === 'user' ? (role as UserRole) : null;
}
