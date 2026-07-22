import { json } from '@sveltejs/kit';
import { getAuthToken, getSession } from '$lib/server/auth';
import { hasPermission } from '$lib/permissions';
import { deleteGroup, updateGroup } from '$lib/server/groups';
import type { RequestHandler } from './$types';

export const PUT: RequestHandler = async ({ request, cookies, params }) => {
	const session = await getSession(cookies);
	if (!session || !hasPermission(session, 'groups:write')) {
		return json({ error: 'You do not have permission to manage groups.' }, { status: 403 });
	}

	if (params.id === 'admin') {
		return json({ error: "The Admin group can't be modified." }, { status: 403 });
	}

	const body = await request.json().catch(() => null);
	const name = typeof body?.name === 'string' ? body.name.trim() : '';
	const permissions = Array.isArray(body?.permissions) ? body.permissions.map(String) : [];

	if (!name) {
		return json({ error: 'Name is required.' }, { status: 400 });
	}

	const token = getAuthToken(cookies);
	const result = token
		? await updateGroup(token, params.id, { name, permissions })
		: { error: 'Not authenticated.', status: 401 };
	if (result.error) {
		return json({ error: result.error }, { status: result.status });
	}

	return json(result.group, { status: 200 });
};

export const DELETE: RequestHandler = async ({ cookies, params }) => {
	const session = await getSession(cookies);
	if (!session || !hasPermission(session, 'groups:write')) {
		return json({ error: 'You do not have permission to manage groups.' }, { status: 403 });
	}

	if (params.id === 'admin') {
		return json({ error: "The Admin group can't be modified." }, { status: 403 });
	}

	const token = getAuthToken(cookies);
	const result = token
		? await deleteGroup(token, params.id)
		: { error: "Couldn't delete that group.", status: 401 };
	if (result) {
		return json({ error: result.error }, { status: result.status });
	}

	return json({ status: 'deleted' });
};
