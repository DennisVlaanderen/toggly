import { env } from '$env/dynamic/private';

export interface FlagSummary {
	key: string;
	enabled: boolean;
	value: string;
	version: number;
}

const API_ORIGIN = env.AERENDIL_API_ORIGIN?.trim() || 'http://127.0.0.1:8080';

export async function listFlags(token: string): Promise<FlagSummary[]> {
	const response = await fetch(`${API_ORIGIN}/api/flags`, {
		headers: { Authorization: `Bearer ${token}` }
	});
	if (!response.ok) {
		// See the identical comment in lib/server/groups.ts's listGroups.
		console.error(`listFlags: backend returned ${response.status}`);
		return [];
	}

	const payload = await response.json().catch(() => null);
	return Array.isArray(payload?.flags) ? payload.flags : [];
}
