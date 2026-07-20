import { redirect } from '@sveltejs/kit';
import { getAuthToken, getSession } from '$lib/server/auth';
import { listFlags } from '$lib/server/flags';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies }) => {
	const session = await getSession(cookies);
	if (!session) {
		redirect(303, '/login');
	}

	const token = getAuthToken(cookies);
	const flags = token ? await listFlags(token) : [];

	return { ...session, flags };
};
