import { env } from '$env/dynamic/private';

export interface FlagSummary {
	key: string;
	enabled: boolean;
	value: string;
	version: number;
}

const API_ORIGIN = env.TOGGLY_API_ORIGIN?.trim() || 'http://127.0.0.1:8080';

export async function listFlags(token: string): Promise<FlagSummary[]> {
	const response = await fetch(`${API_ORIGIN}/api/flags`, {
		headers: { Authorization: `Bearer ${token}` }
	});
	if (!response.ok) {
		return [];
	}

	const payload = await response.json().catch(() => null);
	return Array.isArray(payload?.flags) ? payload.flags : [];
}
